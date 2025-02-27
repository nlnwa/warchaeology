package flag

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/nlnwa/gowarc/v2"
	"github.com/nlnwa/warchaeology/v3/internal/filter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	RecordId     = "id"
	RecordIdHelp = `filter record ID's. For more than one, repeat flag or use comma separated list.`

	RecordType     = "record-type"
	RecordTypeHelp = `filter records by type. For more than one, repeat the flag or use a comma separated list.
Legal values:
	warcinfo, request, response, metadata, revisit, resource, continuation and conversion`

	ResponseCode     = "response-code"
	ResponseCodeHelp = `filter records by http response code
Example:
	200	- only records with a 200 response
	200-300	- records with response codes between 200 (inclusive) and 300 (exclusive)
	500-	- response codes from 500 and above
	-400	- all response codes below 400`

	MimeType     = "mime-type"
	MimeTypeHelp = `filter records with given mime-types. For more than one, repeat flag or use a comma separated list.`
)

type FilterFlags struct {
	defaultMimeType []string
}

func WithDefaultMimeType(mimeTypes []string) func(*FilterFlags) {
	return func(f *FilterFlags) {
		f.defaultMimeType = mimeTypes
	}
}

func (f FilterFlags) AddFlags(cmd *cobra.Command, opts ...func(*FilterFlags)) {
	for _, opt := range opts {
		opt(&f)
	}

	flags := cmd.Flags()
	flags.StringSlice(RecordId, []string{}, RecordIdHelp)
	flags.StringSliceP(RecordType, "t", []string{}, RecordTypeHelp)
	flags.StringP(ResponseCode, "S", "", ResponseCodeHelp)
	flags.StringSliceP(MimeType, "m", f.defaultMimeType, MimeTypeHelp)

	if err := cmd.RegisterFlagCompletionFunc(RecordType, SliceCompletion{
		"warcinfo",
		"request",
		"response",
		"metadata",
		"revisit",
		"resource",
		"continuation",
		"conversion",
	}.CompletionFn); err != nil {
		fmt.Fprintf(os.Stderr, "failed to register completion function for flag %s: %v\n", RecordType, err)
		os.Exit(1)
	}
}

func (f FilterFlags) ResponseCode() string {
	return viper.GetString(ResponseCode)
}

func (f FilterFlags) RecordId() []string {
	return viper.GetStringSlice(RecordId)
}

func (f FilterFlags) RecordType() []string {
	return viper.GetStringSlice(RecordType)
}

func (f FilterFlags) MimeType() []string {
	return viper.GetStringSlice(MimeType)
}

func (f FilterFlags) ToFilter() (*filter.RecordFilter, error) {
	from, to, err := parseResponseCode(f.ResponseCode())
	if err != nil {
		return nil, fmt.Errorf("failed to parse response code: %w", err)
	}

	recordTypes, err := toRecordTypes(f.RecordType())
	if err != nil {
		return nil, fmt.Errorf("failed to parse record types: %w", err)
	}

	return filter.New(
		filter.WithMimeType(f.MimeType()),
		filter.WithCodeRange(from, to),
		filter.WithRecordIds(f.RecordId()),
		filter.WithRecordTypes(recordTypes),
	), nil
}

func toRecordTypes(recordTypes []string) (recordType gowarc.RecordType, err error) {
	for _, r := range recordTypes {
		switch strings.ToLower(r) {
		case "warcinfo":
			recordType = recordType | gowarc.Warcinfo
		case "request":
			recordType = recordType | gowarc.Request
		case "response":
			recordType = recordType | gowarc.Response
		case "metadata":
			recordType = recordType | gowarc.Metadata
		case "revisit":
			recordType = recordType | gowarc.Revisit
		case "resource":
			recordType = recordType | gowarc.Resource
		case "continuation":
			recordType = recordType | gowarc.Continuation
		case "conversion":
			recordType = recordType | gowarc.Conversion
		default:
			err = errors.New("unknown record type")
		}
	}
	return
}

func parseResponseCode(responseCode string) (fromStatus int, toStatus int, err error) {
	rc := strings.Split(responseCode, "-")
	var code int
	switch len(rc) {
	case 1:
		if len(rc[0]) == 0 {
			toStatus = math.MaxInt32
		} else if code, err = strconv.Atoi(rc[0]); err == nil {
			fromStatus = code
			toStatus = code + 1
		}
	case 2:
		if len(rc[0]) > 0 {
			if code, err = strconv.Atoi(rc[0]); err == nil {
				fromStatus = code
			} else {
				break
			}
		}
		if len(rc[1]) == 0 {
			toStatus = math.MaxInt32
		} else {
			if code, err = strconv.Atoi(rc[1]); err == nil {
				toStatus = code
			} else {
				break
			}
		}
	default:
		err = errors.New("illegal response code")
	}
	return
}
