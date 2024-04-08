package filewalker

import (
	"fmt"
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
