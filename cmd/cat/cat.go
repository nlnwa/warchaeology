package cat

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/filter"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/nlnwa/warchaeology/internal/warc"
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

func runCat(catConfig *config) error {
	warcFileReader, err := gowarc.NewWarcFileReader(catConfig.fileName, catConfig.offset, gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)))
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}

	defer func() { _ = warcFileReader.Close() }()

	records := make(chan warc.Record)

	iterator := warc.Iterator{
		WarcFileReader: warcFileReader,
		Filter:         catConfig.filter,
		Nth:            catConfig.recordNum,
		Limit:          catConfig.recordCount,
		Records:        records,
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	go iterator.Iterate(ctx)

	for record := range records {
		if record.Err != nil {
			return fmt.Errorf("error in record at offset %v: %w", record.Offset, record.Err)
		}
		if err := writeWarcRecord(os.Stdout, record.WarcRecord, catConfig); err != nil {
			return err
		}
	}
	return nil
}

func writeWarcRecord(out io.Writer, warcRecord gowarc.WarcRecord, catConfig *config) error {
	if catConfig.showWarcHeader {
		// Write WARC record version
		_, err := fmt.Fprintf(out, "%v\r\n", warcRecord.Version())
		if err != nil {
			return fmt.Errorf("error writing WARC record version: %w", err)
		}

		// Write WARC header
		_, err = warcRecord.WarcHeader().Write(out)
		if err != nil {
			return fmt.Errorf("error writing WARC header: %w", err)
		}

		// Write newline
		_, err = out.Write([]byte("\r\n"))
		if err != nil {
			return fmt.Errorf("error writing separator: %w", err)
		}
	}

	if catConfig.showProtocolHeader {
		if headerBlock, ok := warcRecord.Block().(gowarc.ProtocolHeaderBlock); ok {
			_, err := out.Write(headerBlock.ProtocolHeaderBytes())
			if err != nil {
				return fmt.Errorf("error writing protocol header: %w", err)
			}
		}
	}

	if catConfig.showPayload {
		if payloadBlock, ok := warcRecord.Block().(gowarc.PayloadBlock); ok {
			reader, err := payloadBlock.PayloadBytes()
			if err != nil {
				return fmt.Errorf("error reading payload: %w", err)
			}
			_, err = io.Copy(out, reader)
			if err != nil {
				return fmt.Errorf("error writing payload: %w", err)
			}
		} else {
			reader, err := warcRecord.Block().RawBytes()
			if err != nil {
				return fmt.Errorf("error reading raw bytes of record block: %w", err)
			}
			_, err = io.Copy(out, reader)
			if err != nil {
				return fmt.Errorf("error writing raw bytes of record block: %w", err)
			}
		}
	}

	// Write end of record separator
	_, err := out.Write([]byte("\r\n\r\n"))
	if err != nil {
		return fmt.Errorf("error writing end of record separator: %w", err)
	}

	return nil
}
