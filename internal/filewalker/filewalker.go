package filewalker

import (
	"archive/tar"
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/klauspost/compress/gzip"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/nlnwa/warchaeology/internal/ftpfs"
	"github.com/nlnwa/warchaeology/internal/hooks"
	"github.com/nlnwa/warchaeology/internal/utils"
	workerPool "github.com/nlnwa/warchaeology/internal/workerpool"
	"github.com/nlnwa/whatwg-url/url"
	"github.com/spf13/afero"
	"github.com/spf13/afero/tarfs"
	"github.com/spf13/viper"
)

type FileWalker interface {
	Walk(ctx context.Context, stats Stats) error
}

type logType uint8

var anim = `-\|/`
var animPos int

const (
	info     logType = 1
	err      logType = 2
	summary  logType = 4
	progress logType = 8
)

type fileWalker struct {
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
	fn func(fs afero.Fs, path string) Result) (FileWalker, error) {
	fileSystem, err := resolveFilesystem()
	if err != nil {
		return nil, fmt.Errorf("error resolving filesystem: original error: '%w'", err)
	}
	return &fileWalker{
		fs:             fileSystem,
		paths:          paths,
		recursive:      recursive,
		followSymlinks: followSymlinks,
		suffixes:       suffixes,
		concurrency:    concurrency,
		processor:      fn,
		processedPaths: NewStringSet(),
	}, nil
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
	fileSystem, err := resolveFilesystem()
	if err != nil {
		return nil, fmt.Errorf("error resolving filesystem: original error: '%w'", err)
	}
	return &fileWalker{
		cmd:                cmd,
		fs:                 fileSystem,
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

func resolveFilesystem() (afero.Fs, error) {
	filesystemDefinition := viper.GetString(flag.SrcFilesystem)
	if filesystemDefinition == "" {
		return afero.NewOsFs(), nil
	}

	url, err := url.Parse(filesystemDefinition)
	if err != nil {
		return nil, fmt.Errorf("error parsing filesystem definition: original error: '%w'", err)
	}

	//ftp://user:password@host:port/path
	if url.Protocol() == "ftp:" {
		hostPort := fmt.Sprintf("%s:%d", url.Host(), url.DecodedPort())
		return ftpfs.New(hostPort, url.Username(), url.Password(), int32(viper.GetInt(flag.Concurrency))), nil
	}

	// tar://path/to/archive.tar
	if url.Protocol() == "tar:" {
		filepath, err := afero.NewOsFs().Open(url.Hostname() + url.Pathname())
		if err != nil {
			return nil, fmt.Errorf("error opening tar file: original error: '%w'", err)
		}
		tarReader := tar.NewReader(filepath)
		return tarfs.New(tarReader), nil
	}

	// tgz://path/to/archive.tar.gz
	if url.Protocol() == "tgz:" {
		filepath, err := afero.NewOsFs().Open(url.Hostname() + url.Pathname())
		if err != nil {
			return nil, fmt.Errorf("error opening tar.gz file: original error: '%w'", err)
		}
		gzipReader, err := gzip.NewReader(filepath)
		if err != nil {
			return nil, fmt.Errorf("error creating gzip reader: original error: '%w'", err)
		}
		tarReader := tar.NewReader(gzipReader)
		return tarfs.New(tarReader), nil
	}

	return nil, fmt.Errorf("unsupported filesystem: %s", filesystemDefinition)
}

//nolint:gocyclo
func (walker *fileWalker) Walk(ctx context.Context, stats Stats) error {
	if viper.GetBool(flag.KeepIndex) {
		if fileIndex, err := NewFileIndex(viper.GetBool(flag.NewIndex), walker.cmd); err != nil {
			return err
		} else {
			walker.fileIndex = fileIndex
		}
		defer walker.fileIndex.Close()
	}

	startTime := time.Now()

	if walker.logFileName != "" {
		var err error
		walker.logFile, err = os.Create(walker.logFileName)
		if err != nil {
			return err
		}
		defer func() { _ = walker.logFile.Close() }()
	}

	pool := workerPool.New(ctx, walker.concurrency)
	resultChan := make(chan Result, 32)
	submitJobFunction := func(fs afero.Fs, path string) {
		pool.Submit(func() error {
			if walker.fileIndex != nil {
				if result := walker.fileIndex.GetFileStats(path); result != nil {
					resultChan <- result
					return nil
				}
			}

			// run openInputFileHook hook
			if err := walker.openInputFileHook.Run(path); err != nil {
				if err == hooks.ErrSkipFile {
					fmt.Printf("Skipping file: %s\n", path)
					return nil
				}
				return fmt.Errorf("error running openInputFileHook: original error: '%w'", err)
			}

			// Process file
			results := walker.processor(fs, path)

			// run closeInputFileHook hook
			if err := walker.closeInputFileHook.Run(path, results.ErrorCount()); err != nil {
				return fmt.Errorf("error running closeInputFileHook: original error: '%w'", err)
			}

			if walker.fileIndex != nil {
				walker.fileIndex.SaveFileStats(path, results)
			}
			resultChan <- results
			return nil
		})
	}

	allResults := &sync.WaitGroup{}
	allResults.Add(1)
	defer func() {
		err := pool.CloseWait()
		if err != nil {
			panic(err)
		}
		resultChan <- nil
		allResults.Wait()
		timeSpent := time.Since(startTime)
		if walker.isLog(summary) {
			walker.logSummary(fmt.Sprintf("Total time: %v, %s", timeSpent, stats))
		} else if walker.isLog(progress) {
			fmt.Printf("                                                                                     \r")
		}
	}()
	go func() {
		count := 0
		for {
			result := <-resultChan
			if result == nil {
				allResults.Done()
				break
			}
			count++
			if result.ErrorCount() > 0 && walker.isLog(err) {
				walker.logError(result, count)
			} else if walker.isLog(info) {
				walker.logInfo(result, count)
			}

			stats.Merge(result.GetStats())
			if result.Fatal() != nil {
				fmt.Printf("ERROR: %s\n", result.Fatal())
			}

			if walker.isLog(progress) {
				fmt.Printf("  %s %s\r", string(anim[animPos]), stats.String())
				animPos++
				if animPos >= len(anim) {
					animPos = 0
				}
			}
		}
	}()
	for _, path := range walker.paths {
		if !walker.processedPaths.Contains(path) {
			if err := walker.walkDir(ctx, path, path, submitJobFunction); err != nil {
				return err
			}
		}
	}
	srcFileList := viper.GetString(flag.SrcFileList)
	if srcFileList != "" {
		sourceFile, err := os.Open(srcFileList)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return nil
		}
		defer func(sfl *os.File) {
			_ = sfl.Close()
		}(sourceFile)

		scanner := bufio.NewScanner(sourceFile) //scan the contents of a file and print line by line
		for scanner.Scan() {
			token := scanner.Text()
			if !walker.processedPaths.Contains(token) {
				if err := walker.walkDir(ctx, token, token, submitJobFunction); err != nil {
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

func (walker *fileWalker) walkDir(ctx context.Context, root, dirName string, fn func(fs afero.Fs, path string)) error {
	return afero.Walk(walker.fs, dirName, func(path string, fileInfo fs.FileInfo, err error) error {
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}
		if fileInfo.IsDir() {
			walker.processedPaths.Add(path)
			if !walker.recursive && root != path {
				return filepath.SkipDir
			}
		} else if !fileInfo.IsDir() && !fileInfo.Mode().IsRegular() {
			if walker.followSymlinks {
				if linkReader, ok := walker.fs.(afero.LinkReader); ok {
					linkPath, err := linkReader.ReadlinkIfPossible(path)
					if err != nil {
						_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
						return filepath.SkipDir
					}
					linkPath = filepath.Join(filepath.Dir(path), linkPath)
					if walker.processedPaths.Contains(linkPath) {
						return nil
					}
					return walker.walkDir(ctx, root, linkPath, fn)
				} else {
					_, _ = fmt.Fprintln(os.Stderr, "Error: Symlinks not supported by filesystem")
					return filepath.SkipDir
				}
			}
		} else if walker.hasSuffix(path) && !walker.processedPaths.Contains(path) {
			walker.processedPaths.Add(path)
			fn(walker.fs, path)
		}

		select {
		case <-ctx.Done():
			return filepath.SkipDir
		default:
			return nil
		}
	})
}

func (walker *fileWalker) hasSuffix(path string) bool {
	if walker.suffixes == nil || len(walker.suffixes) == 0 {
		return true
	}
	for _, suffix := range walker.suffixes {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}
	return false
}

func (walker *fileWalker) isLog(log logType) bool {
	return walker.logConsoleTypes&log != 0 || walker.logFile != nil && walker.logfileTypes&log != 0
}

func (walker *fileWalker) logSummary(str string) {
	if walker.logConsoleTypes&summary != 0 {
		fmt.Println(str)
	}
	if walker.logFile != nil && walker.logfileTypes&summary != 0 {
		_, _ = fmt.Fprintln(walker.logFile, str)
	}
}

func (walker *fileWalker) logInfo(res Result, recNum int) {
	logString := res.Log(recNum)
	if walker.logConsoleTypes&info != 0 {
		fmt.Println(logString)
	}
	if walker.logFile != nil && walker.logfileTypes&info != 0 {
		_, _ = fmt.Fprintln(walker.logFile, logString)
	}
}

func (walker *fileWalker) logError(res Result, recordNumber int) {
	recordNumberLogString := res.Log(recordNumber)
	errorString := strings.ReplaceAll(res.Error(), "\n", "\n  ")
	if walker.logConsoleTypes&err != 0 {
		fmt.Println(recordNumberLogString)
		fmt.Println(errorString)
	}
	if walker.logFile != nil && walker.logfileTypes&err != 0 {
		_, _ = fmt.Fprintln(walker.logFile, recordNumberLogString)
		_, _ = fmt.Fprintln(walker.logFile, errorString)
	}
}
