package arcreader

import (
	"bufio"
	"io"
	"iter"

	"github.com/nlnwa/gowarc/v3"
	"github.com/spf13/afero"
)

type ArcFileReader struct {
	file           afero.File
	initialOffset  int64
	warcReader     gowarc.Unmarshaler
	countingReader *CountingReader
	bufferedReader *bufio.Reader
}

func NewArcFileReader(fs afero.Fs, filename string, offset int64, opts ...gowarc.WarcRecordOption) (*ArcFileReader, error) {
	file, err := fs.Open(filename)
	if err != nil {
		return nil, err
	}

	wf := &ArcFileReader{
		file:       file,
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

func (wf *ArcFileReader) Next() (gowarc.Record, error) {
	positionBefore := wf.initialOffset + wf.countingReader.N() - int64(wf.bufferedReader.Buffered())

	record, recordOffset, validation, err := wf.warcReader.Unmarshal(wf.bufferedReader)

	positionAfter := wf.initialOffset + wf.countingReader.N() - int64(wf.bufferedReader.Buffered())
	offset := positionBefore + recordOffset
	size := positionAfter - offset

	return gowarc.Record{
		WarcRecord: record,
		Offset:     offset,
		Size:       size,
		Validation: validation,
	}, err
}

func (wf *ArcFileReader) Records() iter.Seq2[gowarc.Record, error] {
	return func(yield func(gowarc.Record, error) bool) {
		for {
			rec, err := wf.Next()
			if err == io.EOF {
				return
			}
			if !yield(rec, err) {
				return
			}
			if err != nil {
				return
			}
		}
	}
}

func (wf *ArcFileReader) Close() error {
	return wf.file.Close()
}
