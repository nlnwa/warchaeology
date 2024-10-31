package ls

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/nlnwa/gowarc/v2"
	"github.com/spf13/afero"
)

var (
	testDataDir    = filepath.Join("..", "..", "testdata")
	warcWithErrors = filepath.Join(testDataDir, "warc", "samsung-with-error", "rec-33318048d933-20240317162652059-0.warc.gz")
)

func TestListFile(t *testing.T) {
	got := new(bytes.Buffer)

	fields := ""
	delimiter := " "
	writer, err := NewRecordWriter(got, fields, delimiter)
	if err != nil {
		t.Fatal(err)
	}

	opts := &ListOptions{
		writer:            writer,
		warcRecordOptions: []gowarc.WarcRecordOption{gowarc.WithBufferTmpDir(t.TempDir())},
	}

	err = opts.listFile(context.TODO(), afero.NewOsFs(), warcWithErrors)
	if err == nil {
		t.Error("expected error, got nil")
	}
	// TODO: check that the result contains the expected values
}

func BenchmarkListFile(b *testing.B) {
	got := new(bytes.Buffer)
	fields := ""
	delimiter := " "
	writer, err := NewRecordWriter(got, fields, delimiter)
	if err != nil {
		b.Fatal(err)
	}

	opts := &ListOptions{
		writer:            writer,
		warcRecordOptions: []gowarc.WarcRecordOption{gowarc.WithBufferTmpDir(b.TempDir())},
	}

	for i := 0; i < b.N; i++ {
		_ = opts.listFile(context.TODO(), afero.NewOsFs(), warcWithErrors)
	}
}
