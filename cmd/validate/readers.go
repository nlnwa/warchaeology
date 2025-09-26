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
)

// NewCountingReader creates a new countingReader using the given reader and hash function.
func NewCountingReader(r io.Reader, hashFunction string) *countingReader {
	var hash hash.Hash
	switch hashFunction {
	case "md5":
		hash = crypto.MD5.New()
	case "sha1":
		hash = crypto.SHA1.New()
	case "sha256":
		hash = crypto.SHA256.New()
	case "sha512":
		hash = crypto.SHA512.New()
	}
	return &countingReader{
		Reader: r,
		hash:   hash,
	}
}

// countingReader wraps an io.Reader and counts the number of bytes read and calculates a hash.
type countingReader struct {
	io.Reader

	size int64
	hash hash.Hash
}

// Read reads data from the underlying reader and updates the size and hash.
func (reader *countingReader) Read(byteSlice []byte) (length int, err error) {
	length, err = reader.Reader.Read(byteSlice)
	reader.size += int64(length)
	if reader.hash != nil {
		reader.hash.Write(byteSlice[:length])
	}
	return
}

// Size returns the number of bytes read so far.
func (reader *countingReader) Size() int64 {
	return reader.size
}

// Hash returns the hash of the data read so far in hexadecimal format.
func (reader *countingReader) Hash() string {
	if reader.hash == nil {
		return ""
	}
	return fmt.Sprintf("%0x", reader.hash.Sum(nil))
}
