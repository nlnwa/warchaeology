package filewalker

import (
	"context"
	"fmt"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/nlnwa/warchaeology/internal/workerpool"
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
	cmd             string
	paths           []string
	recursive       bool
	followSymlinks  bool
	suffixes        []string
	concurrency     int
	processor       func(path string) Result
	fn              func(path string)
	logFileName     string
	logFile         *os.File
	logfileTypes    logType
	logConsoleTypes logType
	processedPaths  map[string]bool
	fileIndex       *FileIndex
}

func New(paths []string, recursive, followSymlinks bool, suffixes []string, concurrency int, fn func(path string) Result) FileWalker {
	return &filewalker{paths: paths, recursive: recursive, followSymlinks: followSymlinks, suffixes: suffixes, concurrency: concurrency, processor: fn, processedPaths: map[string]bool{}}
}

func NewFromViper(cmd string, paths []string, fn func(path string) Result) FileWalker {
	var consoleType logType
	var fileType logType
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
	for _, t := range viper.GetStringSlice(flag.LogFile) {
		switch strings.ToLower(t) {
		case "info":
			fileType = fileType | info
		case "error":
			fileType = fileType | err
		case "summary":
			fileType = fileType | summary
		default:
			panic("Illegal config value '" + t + "' for " + flag.LogFile)
		}
	}
	return &filewalker{
		cmd:             cmd,
		paths:           paths,
		recursive:       viper.GetBool(flag.Recursive),
		followSymlinks:  viper.GetBool(flag.FollowSymlinks),
		suffixes:        viper.GetStringSlice(flag.Suffixes),
		concurrency:     viper.GetInt(flag.Concurrency),
		processor:       fn,
		logFileName:     viper.GetString(flag.LogFileName),
		logfileTypes:    fileType,
		logConsoleTypes: consoleType,
		processedPaths:  map[string]bool{},
	}
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
	f.fn = func(path string) {
		wp.Submit(func() {
			if f.fileIndex != nil {
				if r := f.fileIndex.GetFileStats(path); r != nil {
					resultChan <- r
					return
				}
			}

			r := f.processor(path)

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
		if !f.processedPaths[p] {
			if err := f.walkDir(ctx, p, p); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *filewalker) walkDir(ctx context.Context, root, dirName string) error {
	return filepath.WalkDir(dirName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
			return filepath.SkipDir
		}
		if d.IsDir() {
			f.processedPaths[path] = true
			if !f.recursive && root != path {
				return filepath.SkipDir
			}
		} else if !d.IsDir() && !d.Type().IsRegular() {
			if f.followSymlinks {
				s, _ := filepath.EvalSymlinks(path)
				if f.processedPaths[s] {
					return nil
				}
				return f.walkDir(ctx, root, s)
			}
		} else if f.hasSuffix(path) {
			f.fn(path)
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

var anim string = `-\|/`
var animPos int
