package validate

import (
	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/nlnwa/warchaeology/v3/internal/hooks"
	"github.com/nlnwa/warchaeology/v3/internal/stat"
	"github.com/spf13/afero"
)

type teeReader struct {
	io.Reader
	r                   afero.File
	w                   *os.File
	inputFileName       string
	outputFileName      string
	closeOutputFileHook hooks.CloseOutputFileHook
}

func (reader *teeReader) Close(warcInfoId *string, result stat.Result) error {
	if reader.r != nil {
		_ = reader.r.Close()
		reader.r = nil
	}
	if reader.w != nil {
		_ = reader.w.Close()

		return reader.closeOutputFileHook.
			WithSrcFileName(reader.inputFileName).
			WithHash(reader.Hash()).
			WithErrorCount(result.ErrorCount()).
			Run(reader.outputFileName, reader.Size(), *warcInfoId)
	}

	return nil
}

func (reader *teeReader) Size() int64 {
	if countingReader, ok := reader.Reader.(*countingReader); ok {
		return countingReader.size
	}
	return 0
}

func (reader *teeReader) Hash() string {
	if countingReader, ok := reader.Reader.(*countingReader); ok && countingReader.hash != nil {
		return fmt.Sprintf("%0x", countingReader.hash.Sum(nil))
	}
	return ""
}

func NewCountingReader(ioReader io.Reader, hashFunction string) io.Reader {
	countingReader := &countingReader{Reader: ioReader}
	switch hashFunction {
	case "md5":
		countingReader.hash = crypto.MD5.New()
	case "sha1":
		countingReader.hash = crypto.SHA1.New()
	case "sha256":
		countingReader.hash = crypto.SHA256.New()
	case "sha512":
		countingReader.hash = crypto.SHA512.New()
	}
	return countingReader
}

type countingReader struct {
	io.Reader
	size int64
	hash hash.Hash
}

func (reader *countingReader) Read(byteSlice []byte) (length int, err error) {
	length, err = reader.Reader.Read(byteSlice)
	reader.size += int64(length)
	if reader.hash != nil {
		reader.hash.Write(byteSlice[:length])
	}
	return
}
