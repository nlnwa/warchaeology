package filewalker

import (
	"fmt"
	"strings"
	"sync"
	"time"

	workerPool "github.com/nlnwa/warchaeology/internal/workerpool"
)

func closePool(walker *fileWalker, pool *workerPool.WorkerPool, resultChan chan Result, allResults *sync.WaitGroup, startTime time.Time, stats Stats) {
	pool.CloseWait()
	resultChan <- nil
	allResults.Wait()
	timeSpent := time.Since(startTime)
	if walker.isLog(summaryLogType) {
		walker.logSummary(fmt.Sprintf("Total time: %v, %s", timeSpent, stats))
	} else if walker.isLog(progressLogType) {
		fmt.Printf("                                                                                     \r")
	}
}

func printResultsAndProgress(walker *fileWalker, resultChan chan Result, allResults *sync.WaitGroup, stats Stats) {
	count := 0
	for {
		result := <-resultChan
		if result == nil {
			allResults.Done()
			break
		}
		count++
		if result.ErrorCount() > 0 && walker.isLog(errorLogType) {
			walker.logError(result, count)
		} else if walker.isLog(infoLogType) {
			walker.logInfo(result, count)
		}

		stats.Merge(result.GetStats())
		if result.Fatal() != nil {
			fmt.Printf("ERROR: %s\n", result.Fatal())
		}

		if walker.isLog(progressLogType) {
			fmt.Printf("  %s %s\r", string(anim[animPos]), stats.String())
			animPos++
			if animPos >= len(anim) {
				animPos = 0
			}
		}
	}
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
	if walker.logConsoleTypes&summaryLogType != 0 {
		fmt.Println(str)
	}
	if walker.logFile != nil && walker.logfileTypes&summaryLogType != 0 {
		_, _ = fmt.Fprintln(walker.logFile, str)
	}
}

func (walker *fileWalker) logInfo(res Result, recNum int) {
	logString := res.Log(recNum)
	if walker.logConsoleTypes&infoLogType != 0 {
		fmt.Println(logString)
	}
	if walker.logFile != nil && walker.logfileTypes&infoLogType != 0 {
		_, _ = fmt.Fprintln(walker.logFile, logString)
	}
}

func (walker *fileWalker) logError(res Result, recordNumber int) {
	recordNumberLogString := res.Log(recordNumber)
	errorString := strings.ReplaceAll(res.Error(), "\n", "\n  ")
	if walker.logConsoleTypes&errorLogType != 0 {
		fmt.Println(recordNumberLogString)
		fmt.Println(errorString)
	}
	if walker.logFile != nil && walker.logfileTypes&errorLogType != 0 {
		_, _ = fmt.Fprintln(walker.logFile, recordNumberLogString)
		_, _ = fmt.Fprintln(walker.logFile, errorString)
	}
}
