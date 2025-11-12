package console

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/nlnwa/gowarc/v2"
	"github.com/nlnwa/warchaeology/v4/cmd/internal/flag"
	"github.com/spf13/viper"
)

type record struct {
	id         string
	offset     int64
	recordType gowarc.RecordType
	hasError   bool
}

func (warcRecordRecord record) String() string {
	result := strings.Builder{}
	if warcRecordRecord.hasError {
		result.WriteString(escapeFgColor(ErrorColor))
		result.WriteString(warcRecordRecord.id)
		result.WriteString(escapeFgColor(gocui.ColorDefault))
	} else {
		reset := escapeFgColor(gocui.ColorDefault)
		switch warcRecordRecord.recordType {
		case gowarc.Warcinfo:
			result.WriteString(fmt.Sprintf("%s%s%s", escapeFgColor(WarcInfoColor), warcRecordRecord.id, reset))
		case gowarc.Request:
			result.WriteString(fmt.Sprintf("%s%s%s", escapeFgColor(RequestColor), warcRecordRecord.id, reset))
		case gowarc.Response:
			result.WriteString(fmt.Sprintf("%s%s%s", escapeFgColor(ResponseColor), warcRecordRecord.id, reset))
		case gowarc.Metadata:
			result.WriteString(fmt.Sprintf("%s%s%s", escapeFgColor(MetadataColor), warcRecordRecord.id, reset))
		case gowarc.Resource:
			result.WriteString(fmt.Sprintf("%s%s%s", escapeFgColor(ResourceColor), warcRecordRecord.id, reset))
		case gowarc.Revisit:
			result.WriteString(fmt.Sprintf("%s%s%s", escapeFgColor(RevisitColor), warcRecordRecord.id, reset))
		case gowarc.Continuation:
			result.WriteString(fmt.Sprintf("%s%s%s", escapeFgColor(ContinuationColor), warcRecordRecord.id, reset))
		case gowarc.Conversion:
			result.WriteString(fmt.Sprintf("%s%s%s", escapeFgColor(ConversionColor), warcRecordRecord.id, reset))
		default:
			result.WriteString(warcRecordRecord.id)
		}
	}
	return result.String()
}

func populateRecords(gui *gocui.Gui, ctx context.Context, finishedCb func(), widgetList *ListWidget, data any) {
	warcFileReader, err := gowarc.NewWarcFileReader(data.(string), 0, gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)))
	if err != nil {
		panic(err)
	}
	defer warcFileReader.Close()

	for {
		select {
		case <-ctx.Done():
			goto end
		default:
			warcRecord, offset, validate, err := warcFileReader.Next()
			if err == io.EOF {
				goto end
			}
			if err != nil {
				goto end
			}
			_ = warcRecord.ValidateDigest(validate)

			warcRecordRecord := record{
				id:         warcRecord.WarcHeader().Get(gowarc.WarcRecordID),
				offset:     offset,
				recordType: warcRecord.Type(),
			}

			if err := warcRecord.Close(); err != nil {
				*validate = append(*validate, err)
			}
			warcRecordRecord.hasError = !validate.Valid()

			widgetList.records = append(widgetList.records, warcRecordRecord)
		}
	}
end:
	finishedCb()
}

func populateFiles(gui *gocui.Gui, ctx context.Context, finishedCb func(), widgetList *ListWidget, data any) {
	state.dir = data.(string)
	if view, err := gui.View("dir"); err == nil {
		view.Title = state.dir
	}

	if len(state.files) > 0 {
		for _, file := range state.files {
			for _, suffix := range state.suffixes {
				if strings.HasSuffix(file, suffix) {
					widgetList.records = append(widgetList.records, file)
				}
			}
		}
		finishedCb()
		return
	}

	entries, err := os.ReadDir(state.dir)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			goto end
		default:
			if strings.HasSuffix(entry.Name(), ".warc") || strings.HasSuffix(entry.Name(), ".warc.gz") {
				widgetList.records = append(widgetList.records, entry.Name())
			}
		}
	}
end:
	finishedCb()
}

// Copied from standard lib for go 1.17 while we are waiting for 1.17 to be in common use
// dirInfo is a DirEntry based on a FileInfo.
type dirInfo struct {
	fileInfo fs.FileInfo
}

func (directoryInfo dirInfo) IsDir() bool {
	return directoryInfo.fileInfo.IsDir()
}

func (directoryInfo dirInfo) Type() fs.FileMode {
	return directoryInfo.fileInfo.Mode().Type()
}

func (directoryInfo dirInfo) Info() (fs.FileInfo, error) {
	return directoryInfo.fileInfo, nil
}

func (directoryInfo dirInfo) Name() string {
	return directoryInfo.fileInfo.Name()
}

// FileInfoToDirEntry returns a DirEntry that returns information from info.
// If info is nil, FileInfoToDirEntry returns nil.
func FileInfoToDirEntry(fileInfo fs.FileInfo) fs.DirEntry {
	if fileInfo == nil {
		return nil
	}
	return dirInfo{fileInfo: fileInfo}
}
