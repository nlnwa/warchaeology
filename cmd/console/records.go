package console

import (
	"context"
	"fmt"
	"github.com/awesome-gocui/gocui"
	"github.com/nlnwa/gowarc"
	"io"
	"os"
	"strings"
)

type record struct {
	id         string
	offset     int64
	recordType gowarc.RecordType
	hasError   bool
}

func (r record) String() string {
	sb := strings.Builder{}
	if r.hasError {
		sb.WriteString(escapeFgColor(ErrorColor))
		sb.WriteString(r.id)
		sb.WriteString(escapeFgColor(gocui.ColorDefault))
	} else {
		reset := escapeFgColor(gocui.ColorDefault)
		switch r.recordType {
		case gowarc.Warcinfo:
			sb.WriteString(fmt.Sprintf("%s%s%s", escapeFgColor(WarcInfoColor), r.id, reset))
		case gowarc.Request:
			sb.WriteString(fmt.Sprintf("%s%s%s", escapeFgColor(RequestColor), r.id, reset))
		case gowarc.Response:
			sb.WriteString(fmt.Sprintf("%s%s%s", escapeFgColor(ResponseColor), r.id, reset))
		case gowarc.Metadata:
			sb.WriteString(fmt.Sprintf("%s%s%s", escapeFgColor(MetadataColor), r.id, reset))
		case gowarc.Resource:
			sb.WriteString(fmt.Sprintf("%s%s%s", escapeFgColor(ResourceColor), r.id, reset))
		case gowarc.Revisit:
			sb.WriteString(fmt.Sprintf("%s%s%s", escapeFgColor(RevisitColor), r.id, reset))
		case gowarc.Continuation:
			sb.WriteString(fmt.Sprintf("%s%s%s", escapeFgColor(ContinuationColor), r.id, reset))
		case gowarc.Conversion:
			sb.WriteString(fmt.Sprintf("%s%s%s", escapeFgColor(ConversionColor), r.id, reset))
		default:
			sb.WriteString(r.id)
		}
	}
	return sb.String()
}

func populateRecords(g *gocui.Gui, ctx context.Context, finishedCb func(), widget *ListWidget, data interface{}) {
	r, err := gowarc.NewWarcFileReader(data.(string), 0)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	for {
		select {
		case <-ctx.Done():
			goto end
		default:
			rec, offset, validate, err := r.Next()
			if err == io.EOF {
				goto end
			}
			if err != nil {
				goto end
			}
			rec.ValidateDigest(validate)

			wr := record{
				id:         rec.WarcHeader().Get(gowarc.WarcRecordID),
				offset:     offset,
				recordType: rec.Type(),
				hasError:   !validate.Valid(),
			}
			rec.Close()
			widget.records = append(widget.records, wr)
		}
	}
end:
	finishedCb()
}

func populateFiles(g *gocui.Gui, ctx context.Context, finishedCb func(), widget *ListWidget, data interface{}) {
	state.dir = data.(string)
	if v, err := g.View("dir"); err == nil {
		v.Title = state.dir
	}
	entries, err := os.ReadDir(state.dir)
	if err != nil {
		panic(err)
	}

	for _, e := range entries {
		select {
		case <-ctx.Done():
			goto end
		default:
			if strings.HasSuffix(e.Name(), ".warc") || strings.HasSuffix(e.Name(), ".warc.gz") {
				widget.records = append(widget.records, e.Name())
			}
		}
	}
end:
	finishedCb()
}
