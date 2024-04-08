package listRecordImplementation

import (
	"fmt"
	"io"

	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/filewalker2"
	"github.com/spf13/afero"
)

type WarcRecord struct {
	Record       gowarc.WarcRecord
	WarcFilePath string
	Offset       int64
	Validation   *gowarc.Validation
}

func ListRecords(warcFilePath string) ([]WarcRecord, error) {
	aferoFileSystemReplaceLater := afero.NewOsFs()                  // TODO: replace later
	fileToProcessChannel := make(chan filewalker2.FileToProcess, 1) // TODO: replace later
	err := filewalker2.PopulateChannelWithFilesToProcess(aferoFileSystemReplaceLater, warcFilePath, fileToProcessChannel)
	if err != nil {
		return nil, fmt.Errorf("failed to populate channel with files to process, original error: `%w`", err)
	}
	results := []WarcRecord{}
	for fileToProcess := range fileToProcessChannel {
		if fileToProcess.Error != nil {
			return nil, fmt.Errorf("error while processing file '%s', original error: `%w`", fileToProcess.Path, fileToProcess.Error)
		}
		warcFileReader, err := createWarcRecordReader(fileToProcess.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to create warc record reader, original error: `%w`", err)
		}
		defer closeWarcFileReader(warcFileReader)

		for {
			warcRecord, offset, validation, err := warcFileReader.Next()

			if err != nil {
				if err == io.ErrUnexpectedEOF {
					break
				}
				return nil, fmt.Errorf("error while processing warc record, original error: `%w`", err)
			}
			results = append(results, WarcRecord{
				Record:       warcRecord,
				WarcFilePath: fileToProcess.Path,
				Offset:       offset,
				Validation:   validation,
			})
		}
	}
	return results, nil
}

func createWarcRecordReader(filename string) (*gowarc.WarcFileReader, error) {
	warcFileReader, err := gowarc.NewWarcFileReader(filename, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to create warc file reader, original error: `%w`", err)
	}
	return warcFileReader, err
}

func closeWarcFileReader(warcFileReader *gowarc.WarcFileReader) {
	err := warcFileReader.Close()
	if err != nil {
		fmt.Printf("error closing warc file reader, original error: '%v'\n", err)
	}
}
