package arcreader

import (
	"bufio"

	"github.com/nlnwa/gowarc"
	"github.com/spf13/afero"
)

type ArcFileReader struct {
	file           afero.File
	initialOffset  int64
	offset         int64
	warcReader     gowarc.Unmarshaler
	countingReader *CountingReader
	bufferedReader *bufio.Reader
	currentRecord  gowarc.WarcRecord
}

func NewArcFileReader(fs afero.Fs, filename string, offset int64, opts ...gowarc.WarcRecordOption) (*ArcFileReader, error) {
	file, err := fs.Open(filename) // For read access.
	if err != nil {
		return nil, err
	}

	wf := &ArcFileReader{
		file:       file,
		offset:     offset,
		warcReader: NewUnmarshaler(opts...),
	}
	_, err = file.Seek(offset, 0)
	if err != nil {
		return nil, err
	}

	wf.countingReader = NewCountingReader(file)
	wf.initialOffset = offset
	wf.bufferedReader = bufio.NewReaderSize(wf.countingReader, 4*1024)
	return wf, nil
}

func (wf *ArcFileReader) Next() (gowarc.WarcRecord, int64, *gowarc.Validation, error) {
	var validation *gowarc.Validation
	if wf.currentRecord != nil {
		if err := wf.currentRecord.Close(); err != nil {
			return nil, wf.offset, validation, err
		}
	}
	wf.offset = wf.initialOffset + wf.countingReader.N() - int64(wf.bufferedReader.Buffered())
	fs, _ := wf.file.Stat()
	if fs.Size() <= wf.offset {
		wf.offset = 0
	}

	var err error
	var recordOffset int64
	wf.currentRecord, recordOffset, validation, err = wf.warcReader.Unmarshal(wf.bufferedReader)
	return wf.currentRecord, wf.offset + recordOffset, validation, err
}

func (wf *ArcFileReader) Close() error {
	return wf.file.Close()
}
