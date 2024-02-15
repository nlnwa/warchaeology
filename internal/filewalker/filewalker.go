package filewalker

import (
	"bufio"
	"context"
	"fmt"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/nlnwa/warchaeology/internal/ftpfs"
	"github.com/nlnwa/warchaeology/internal/hooks"
	"github.com/nlnwa/warchaeology/internal/utils"
	"github.com/nlnwa/warchaeology/internal/workerpool"
	"github.com/nlnwa/whatwg-url/url"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type FileWalker interface {
	Walk(ctx context.Context, stats Stats) error
}

type logType uint8

const (
	info     logType = 1
	err      logType = 2
	summary  logType = 4
	progress logType = 8
)

type filewalker struct {
	cmd                string
	fs                 afero.Fs
	paths              []string
	recursive          bool
	followSymlinks     bool
	suffixes           []string
	concurrency        int
	processor          func(fs afero.Fs, path string) Result
	logFileName        string
	logFile            *os.File
	logfileTypes       logType
	logConsoleTypes    logType
	processedPaths     StringSet
	fileIndex          *FileIndex
	openInputFileHook  hooks.OpenInputFileHook
	closeInputFileHook hooks.CloseInputFileHook
}

func New(paths []string, recursive, followSymlinks bool, suffixes []string, concurrency int,
	fn func(fs afero.Fs, path string) Result) FileWalker {
	return &filewalker{
		fs:             resolveFs(),
		paths:          paths,
		recursive:      recursive,
		followSymlinks: followSymlinks,
		suffixes:       suffixes,
		concurrency:    concurrency,
		processor:      fn,
		processedPaths: NewStringSet(),
	}
}

func NewFromViper(cmd string, paths []string, fn func(fs afero.Fs, path string) Result) (FileWalker, error) {
	var consoleType logType
	var fileType logType
	if utils.StdoutIsTerminal() {
		for _, t := range viper.GetStringSlice(flag.LogConsole) {
			switch strings.ToLower(t) {
			case "info":
				consoleType = consoleType | info
			case "error":
				consoleType = consoleType | err
			case "summary":
				consoleType = consoleType | summary
			case "progress":
				consoleType = consoleType | progress
			default:
				panic("Illegal config value '" + t + "' for " + flag.LogConsole)
			}
		}
	}
	for _, t := range viper.GetStringSlice(flag.LogFile) {
		switch strings.ToLower(t) {
		case "info":
			fileType = fileType | info
		case "error":
			fileType = fileType | err
		case "summary":
			fileType = fileType | summary
		default:
			return nil, fmt.Errorf("illegal config value '%s' for %s", t, flag.LogFile)
		}
	}
	var beforeProcessFile hooks.OpenInputFileHook
	var afterProcessFile hooks.CloseInputFileHook
	var err error
	if hook := viper.GetString(flag.OpenInputFileHook); hook != "" {
		beforeProcessFile, err = hooks.NewOpenInputFileHook(cmd, hook)
		if err != nil {
			return nil, err
		}
	}
	if hook := viper.GetString(flag.CloseInputFileHook); hook != "" {
		afterProcessFile, err = hooks.NewCloseInputFileHook(cmd, hook)
		if err != nil {
			return nil, err
		}
	}
	return &filewalker{
		cmd:                cmd,
		fs:                 resolveFs(),
		paths:              paths,
		recursive:          viper.GetBool(flag.Recursive),
		followSymlinks:     viper.GetBool(flag.FollowSymlinks),
		suffixes:           viper.GetStringSlice(flag.Suffixes),
		concurrency:        viper.GetInt(flag.Concurrency),
		processor:          fn,
		logFileName:        viper.GetString(flag.LogFileName),
		logfileTypes:       fileType,
		logConsoleTypes:    consoleType,
		processedPaths:     NewStringSet(),
		openInputFileHook:  beforeProcessFile,
		closeInputFileHook: afterProcessFile,
	}, nil
}

func resolveFs() afero.Fs {
	fsDef := viper.GetString(flag.SrcFilesystem)
	if fsDef == "" {
		return afero.NewOsFs()
	}

	u, err := url.Parse(fsDef)
	if err != nil {
		panic(err)
	}

	//ftp://user:password@host:port/path
	if u.Protocol() == "ftp:" {
		hostPort := fmt.Sprintf("%s:%d", u.Host(), u.DecodedPort())
		return ftpfs.New(hostPort, u.Username(), u.Password(), int32(viper.GetInt(flag.Concurrency)))
	}

	panic("Unsupported filesystem: " + fsDef)
}

func (f *filewalker) Walk(ctx context.Context, stats Stats) error {
	if viper.GetBool(flag.KeepIndex) {
		if fileIndex, err := NewFileIndex(viper.GetBool(flag.NewIndex), f.cmd); err != nil {
			return err
		} else {
			f.fileIndex = fileIndex
		}
		defer f.fileIndex.Close()
	}

	startTime := time.Now()

	if f.logFileName != "" {
		var err error
		f.logFile, err = os.Create(f.logFileName)
		if err != nil {
			return err
		}
		defer func() { _ = f.logFile.Close() }()
	}

	wp := workerpool.New(ctx, f.concurrency)
	resultChan := make(chan Result, 32)
	fn := func(fs afero.Fs, path string) {
		wp.Submit(func() {
			if f.fileIndex != nil {
				if r := f.fileIndex.GetFileStats(path); r != nil {
					resultChan <- r
					return
				}
			}

			// run openInputFileHook hook
			if err := f.openInputFileHook.Run(path); err != nil {
				panic(err)
			}

			// Process file
			r := f.processor(fs, path)

			// run closeInputFileHook hook
			if err := f.closeInputFileHook.Run(path, r.ErrorCount()); err != nil {
				panic(err)
			}

			if f.fileIndex != nil {
				f.fileIndex.SaveFileStats(path, r)
			}
			resultChan <- r
		})
	}

	allResults := &sync.WaitGroup{}
	allResults.Add(1)
	defer func() {
		wp.CloseWait()
		resultChan <- nil
		allResults.Wait()
		timeSpent := time.Since(startTime)
		if f.isLog(summary) {
			s := fmt.Sprintf("Total time: %v, %s", timeSpent, stats)
			f.logSummary(s)
		} else if f.isLog(progress) {
			fmt.Printf("                                                                                     \r")
		}
	}()
	go func() {
		count := 0
		for {
			res := <-resultChan
			if res == nil {
				allResults.Done()
				break
			}
			count++
			if res.ErrorCount() > 0 && f.isLog(err) {
				f.logError(res, count)
			} else if f.isLog(info) {
				f.logInfo(res, count)
			}

			stats.Merge(res.GetStats())
			if res.Fatal() != nil {
				fmt.Printf("ERROR: %s\n", res.Fatal())
			}

			if f.isLog(progress) {
				fmt.Printf("  %s %s\r", string(anim[animPos]), stats.String())
				animPos++
				if animPos >= len(anim) {
					animPos = 0
				}
			}
		}
	}()
	for _, p := range f.paths {
		if !f.processedPaths.Contains(p) {
			if err := f.walkDir(ctx, p, p, fn); err != nil {
				return err
			}
		}
	}
	srcFileList := viper.GetString(flag.SrcFileList)
	if srcFileList != "" {
		sfl, err := os.Open(srcFileList) //open the file
		if err != nil {
			fmt.Println("Error opening file:", err)
			return nil
		}
		defer func(sfl *os.File) {
			_ = sfl.Close()
		}(sfl)

		scanner := bufio.NewScanner(sfl) //scan the contents of a file and print line by line
		for scanner.Scan() {
			p := scanner.Text()
			if !f.processedPaths.Contains(p) {
				if err := f.walkDir(ctx, p, p, fn); err != nil {
					return err
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading from file:", err) //print error if scanning is not done properly
		}
	}
	return nil
}

func (f *filewalker) walkDir(ctx context.Context, root, dirName string, fn func(fs afero.Fs, path string)) error {
	return afero.Walk(f.fs, dirName, func(path string, d fs.FileInfo, err error) error {
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}
		if d.IsDir() {
			f.processedPaths.Add(path)
			if !f.recursive && root != path {
				return filepath.SkipDir
			}
		} else if !d.IsDir() && !d.Mode().IsRegular() {
			if f.followSymlinks {
				if lr, ok := f.fs.(afero.LinkReader); ok {
					s, err := lr.ReadlinkIfPossible(path)
					if err != nil {
						_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
						return filepath.SkipDir
					}
					s = filepath.Join(filepath.Dir(path), s)
					if f.processedPaths.Contains(s) {
						return nil
					}
					return f.walkDir(ctx, root, s, fn)
				} else {
					_, _ = fmt.Fprintln(os.Stderr, "Error: Symlinks not supported by filesystem")
					return filepath.SkipDir
				}
			}
		} else if f.hasSuffix(path) && !f.processedPaths.Contains(path) {
			f.processedPaths.Add(path)
			fn(f.fs, path)
		}

		select {
		case <-ctx.Done():
			return filepath.SkipDir
		default:
			return nil
		}
	})
}

func (f *filewalker) hasSuffix(path string) bool {
	if f.suffixes == nil || len(f.suffixes) == 0 {
		return true
	}
	for _, s := range f.suffixes {
		if strings.HasSuffix(path, s) {
			return true
		}
	}
	return false
}

func (f *filewalker) isLog(l logType) bool {
	return f.logConsoleTypes&l != 0 || f.logFile != nil && f.logfileTypes&l != 0
}

func (f *filewalker) logSummary(s string) {
	if f.logConsoleTypes&summary != 0 {
		fmt.Println(s)
	}
	if f.logFile != nil && f.logfileTypes&summary != 0 {
		_, _ = fmt.Fprintln(f.logFile, s)
	}
}

func (f *filewalker) logInfo(res Result, recNum int) {
	s := res.Log(recNum)
	if f.logConsoleTypes&info != 0 {
		fmt.Println(s)
	}
	if f.logFile != nil && f.logfileTypes&info != 0 {
		_, _ = fmt.Fprintln(f.logFile, s)
	}
}

func (f *filewalker) logError(res Result, recNum int) {
	s := res.Log(recNum)
	e := strings.ReplaceAll(res.Error(), "\n", "\n  ")
	if f.logConsoleTypes&err != 0 {
		fmt.Println(s)
		fmt.Println(e)
	}
	if f.logFile != nil && f.logfileTypes&err != 0 {
		_, _ = fmt.Fprintln(f.logFile, s)
		_, _ = fmt.Fprintln(f.logFile, e)
	}
}

var anim = `-\|/`
var animPos int
