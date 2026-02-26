package console

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/nationallibraryofnorway/warchaeology/v4/cmd/internal/flag"
	"github.com/nlnwa/gowarc/v3"
	"github.com/spf13/viper"
)

type record struct {
	id         string
	offset     int64
	size       int64
	recordType gowarc.RecordType
	hasError   bool
	errMsg     string
}

func (r record) titleWithSize() string {
	if r.size > 0 {
		return fmt.Sprintf("%s (%dB)", r.id, r.size)
	}
	return r.id
}

func (r record) String() string {
	result := strings.Builder{}
	title := r.titleWithSize()
	if r.hasError {
		result.WriteString(escapeFgColor(ErrorColor))
		result.WriteString(title)
		result.WriteString(escapeFgColor(gocui.ColorDefault))
	} else {
		reset := escapeFgColor(gocui.ColorDefault)
		switch r.recordType {
		case gowarc.Warcinfo:
			fmt.Fprintf(&result, "%s%s%s", escapeFgColor(WarcInfoColor), title, reset)
		case gowarc.Request:
			fmt.Fprintf(&result, "%s%s%s", escapeFgColor(RequestColor), title, reset)
		case gowarc.Response:
			fmt.Fprintf(&result, "%s%s%s", escapeFgColor(ResponseColor), title, reset)
		case gowarc.Metadata:
			fmt.Fprintf(&result, "%s%s%s", escapeFgColor(MetadataColor), title, reset)
		case gowarc.Resource:
			fmt.Fprintf(&result, "%s%s%s", escapeFgColor(ResourceColor), title, reset)
		case gowarc.Revisit:
			fmt.Fprintf(&result, "%s%s%s", escapeFgColor(RevisitColor), title, reset)
		case gowarc.Continuation:
			fmt.Fprintf(&result, "%s%s%s", escapeFgColor(ContinuationColor), title, reset)
		case gowarc.Conversion:
			fmt.Fprintf(&result, "%s%s%s", escapeFgColor(ConversionColor), title, reset)
		default:
			result.WriteString(title)
		}
	}
	return result.String()
}

func populateRecords(gui *gocui.Gui, ctx context.Context, finishedCb func(), widgetList *ListWidget, data any) {
	warcFileReader, err := gowarc.NewWarcFileReader(data.(string), 0, gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)))
	if err != nil {
		widgetList.records = append(widgetList.records, record{hasError: true, errMsg: err.Error()})
		finishedCb()
		return
	}
	defer func() { _ = warcFileReader.Close() }()

	for {
		select {
		case <-ctx.Done():
			goto end
		default:
			rec, err := warcFileReader.Next()
			if err == io.EOF {
				goto end
			}
			if err != nil {
				widgetList.records = append(widgetList.records, record{hasError: true, errMsg: err.Error()})
				goto end
			}
			warcRecord := rec.WarcRecord
			offset := rec.Offset
			validation := append([]error{}, rec.Validation...)
			digestValidation, err := warcRecord.ValidateDigest()
			validation = append(validation, digestValidation...)
			if err != nil {
				validation = append(validation, err)
			}

			warcRecordRecord := record{
				id:         warcRecord.WarcHeader().Get(gowarc.WarcRecordID),
				offset:     offset,
				size:       rec.Size,
				recordType: warcRecord.Type(),
			}

			if err := rec.Close(); err != nil {
				validation = append(validation, err)
			}
			warcRecordRecord.hasError = len(validation) > 0

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
			for _, suffix := range state.suffixes {
				if strings.HasSuffix(entry.Name(), suffix) {
					widgetList.records = append(widgetList.records, entry.Name())
					break
				}
			}
		}
	}
end:
	finishedCb()
}
