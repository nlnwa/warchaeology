package nedlibreader

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/nlnwa/gowarc/v2"
	"github.com/spf13/afero"
)

var (
	testDataDir = filepath.Join("..", "testdata")
)

var testFiles = map[string]string{
	"b863a630196bce1a15ca86b40f34a2d5": filepath.Join(testDataDir, "nedlib", "nb-image", "b863a630196bce1a15ca86b40f34a2d5.meta"),
	"e4a2d28bdf4c38b8f6f291f7c8c958d5": filepath.Join(testDataDir, "nedlib", "nb-image", "e4a2d28bdf4c38b8f6f291f7c8c958d5.meta"),
	"df94a25cc254ba3c9765c5263b7ca6aa": filepath.Join(testDataDir, "nedlib", "bad", "df94a25cc254ba3c9765c5263b7ca6aa.meta"),
}

func TestNedlibReader(t *testing.T) {
	tests := []struct {
		name        string
		warcHeaders map[string]string
		httpHeaders map[string]string
		err         error
	}{
		{
			name: "b863a630196bce1a15ca86b40f34a2d5",
			warcHeaders: map[string]string{
				"WARC-Date":       "2003-01-11T14:37:37Z",
				"WARC-Target-URI": "http://www.nb.no:80/assets/images/Collett9.jpeg",
			},
			httpHeaders: map[string]string{
				"Content-Length": "10920",
				"Content-Type":   "image/jpeg",
			},
		},
		{
			name: "df94a25cc254ba3c9765c5263b7ca6aa",
			err:  errors.New(`bad Content-Length "9726Fri, 04 Apr 2003 08:36:47 GMT"`),
		},
		{
			name: "e4a2d28bdf4c38b8f6f291f7c8c958d5",
			warcHeaders: map[string]string{
				"WARC-Date":       "2003-01-11T14:37:37Z",
				"WARC-Target-URI": "http://www.nb.no:80/assets/images/auto_generated_images/Robert_Collett_KNAPP_1.gif",
			},
			httpHeaders: map[string]string{
				"Content-Length": "1802",
				"Content-Type":   "image/gif",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// resolve test file path
			testFile, err := filepath.Abs(testFiles[test.name])
			if err != nil {
				t.Fatal(err)
			}

			nedlibReader := &NedlibReader{
				fs:                afero.NewReadOnlyFs(afero.NewOsFs()),
				metaFilename:      testFile,
				defaultTime:       time.Time{},
				warcRecordOptions: nil,
			}

			wr, _, _, err := nedlibReader.Next()
			if err != nil {
				if test.err == nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			for key, value := range test.warcHeaders {
				if wr.WarcHeader().Get(key) != value {
					t.Errorf("expected %s: %s, got %s: %s", key, value, key, wr.WarcHeader().Get(key))
				}
			}
			switch v := wr.Block().(type) {
			case gowarc.HttpResponseBlock:
				for key, value := range test.httpHeaders {
					httpHeader := v.HttpHeader()
					if httpHeader.Get(key) != value {
						t.Errorf("expected %s: %s, got %s: %s", key, value, key, httpHeader.Get(key))
					}
				}
			default:
				t.Errorf("expected HttpResponseBlock, got %T", wr.Block())
			}
		})
	}
}
