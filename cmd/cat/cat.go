package cat

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/filter"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/spf13/viper"
)

type config struct {
	offset             int64
	recordNum          int
	recordCount        int
	fileName           string
	filter             *filter.Filter
	showWarcHeader     bool
	showProtocolHeader bool
	showPayload        bool
}

func listRecords(catConfig *config, fileName string) {
	warcFileReader, err := gowarc.NewWarcFileReader(fileName, catConfig.offset, gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)))
	defer func() { _ = warcFileReader.Close() }()
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}

	num := 0
	count := 0

	for {
		warcRecord, _, _, err := warcFileReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v, rec num: %v, Offset %v\n", err.Error(), strconv.Itoa(count), catConfig.offset)
			break
		}

		if !catConfig.filter.Accept(warcRecord) {
			continue
		}

		// Find record number
		if catConfig.recordNum > 0 && num < catConfig.recordNum {
			num++
			continue
		}

		count++
		out := os.Stdout

		if catConfig.showWarcHeader {
			// Write WARC record version
			_, err = fmt.Fprintf(out, "%v\r\n", warcRecord.Version())
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}

			// Write WARC header
			_, err = warcRecord.WarcHeader().Write(out)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}

			// Write separator
			_, err = out.WriteString("\r\n")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}

		if catConfig.showProtocolHeader {
			if headerBlock, ok := warcRecord.Block().(gowarc.ProtocolHeaderBlock); ok {
				_, err = out.Write(headerBlock.ProtocolHeaderBytes())
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			}
		}

		if catConfig.showPayload {
			if payloadBlock, ok := warcRecord.Block().(gowarc.PayloadBlock); ok {
				reader, err := payloadBlock.PayloadBytes()
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
				_, err = io.Copy(out, reader)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			} else {
				reader, err := warcRecord.Block().RawBytes()
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
				_, err = io.Copy(out, reader)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			}
		}

		// Write end of record separator
		_, err = out.WriteString("\r\n\r\n")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		if catConfig.recordCount > 0 && count >= catConfig.recordCount {
			break
		}
	}
	_, _ = fmt.Fprintln(os.Stderr, "Count: ", count)
}
