package ls

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nlnwa/gowarc/v2"
	"github.com/nlnwa/warchaeology/v3/cmd/internal/flag"
	"github.com/nlnwa/warchaeology/v3/internal/filewalker"
	"github.com/nlnwa/warchaeology/v3/internal/filter"
	"github.com/nlnwa/warchaeology/v3/internal/warc"
	"github.com/nlnwa/warchaeology/v3/internal/workerpool"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	Delimiter     = "delimiter"
	DelimiterHelp = `field delimiter`

	Fields     = "fields"
	FieldsHelp = `which fields to include in the output

Field specification letters are mostly the same as the fields in the CDX file specification (https://iipc.github.io/warc-specifications/specifications/cdx-format/cdx-2015/).

The following fields are supported:
	a - original URL
	b - date in 14 digit format
	B - date in RFC3339 format
	e - IP address
	g - filename
	h - original host
	i - record id
	k - checksum
	m - document mime type
	s - http response code
	S - record size
	T - record type
	V - offset

A number after the field letter restricts the field length. By adding a + or - sign before the number the field is padded to have the exact length. + is right aligned and - is left aligned.`
)

type Writer interface {
	WriteRecord(record warc.Record, fileName string) error
}

type ListOptions struct {
	paths             []string
	offset            int64
	recordNum         int
	recordCount       int
	force             bool
	concurrency       int
	continueOnError   bool
	filter            *filter.RecordFilter
	writer            Writer
	fileWalker        *filewalker.FileWalker
	warcRecordOptions []gowarc.WarcRecordOption
}

type ListFlags struct {
	FileWalkerFlags       flag.FileWalkerFlags
	FilterFlags           flag.FilterFlags
	WarcIteratorFlags     flag.WarcIteratorFlags
	WarcRecordOptionFlags flag.WarcRecordOptionFlags
	ConcurrencyFlags      flag.ConcurrencyFlags
	ErrorFlags            flag.ErrorFlags
}

func (f ListFlags) AddFlags(cmd *cobra.Command) {
	f.FileWalkerFlags.AddFlags(cmd)
	f.FilterFlags.AddFlags(cmd)
	f.WarcIteratorFlags.AddFlags(cmd)
	f.WarcRecordOptionFlags.AddFlags(cmd)
	f.ConcurrencyFlags.AddFlags(cmd)
	f.ErrorFlags.AddFlags(cmd)

	flags := cmd.Flags()

	flags.StringP(Delimiter, "d", " ", DelimiterHelp)
	flags.StringP(Fields, "F", "", FieldsHelp)
	flags.Bool("json", false, "output as JSON lines")
}

func (f ListFlags) Delimiter() string {
	return viper.GetString(Delimiter)
}

func (f ListFlags) Fields() string {
	return viper.GetString(Fields)
}

func (f ListFlags) JSON() bool {
	return viper.GetBool("json")
}

func (f ListFlags) ToListOptions() (*ListOptions, error) {
	filter, err := f.FilterFlags.ToFilter()
	if err != nil {
		return nil, err
	}

	paths, err := flag.ReadSrcFileList(f.FileWalkerFlags.SrcFileListFlags.SrcFileList())
	if err != nil {
		return nil, fmt.Errorf("failed to read from source file: %w", err)
	}

	opts := f.WarcRecordOptionFlags.ToWarcRecordOptions()

	var writer Writer
	if f.JSON() {
		writer = NewJSONWriter(os.Stdout, f.Fields())
	} else {
		writer, err = NewRecordWriter(os.Stdout, f.Fields(), f.Delimiter())
		if err != nil {
			return nil, err
		}
	}

	fileWalker, err := f.FileWalkerFlags.ToFileWalker()
	if err != nil {
		return nil, fmt.Errorf("failed to create file walker: %w", err)
	}

	return &ListOptions{
		paths:             paths,
		offset:            f.WarcIteratorFlags.Offset(),
		recordNum:         f.WarcIteratorFlags.RecordNum(),
		recordCount:       f.WarcIteratorFlags.Limit(),
		force:             f.WarcIteratorFlags.Force(),
		concurrency:       f.ConcurrencyFlags.Concurrency(),
		filter:            filter,
		continueOnError:   f.ErrorFlags.ContinueOnError(),
		fileWalker:        fileWalker,
		writer:            writer,
		warcRecordOptions: opts,
	}, nil
}

// NewCmdList creates the ls command
func NewCmdList() *cobra.Command {
	flags := ListFlags{}

	cmd := &cobra.Command{
		Use:   "ls FILE/DIR ...",
		Short: "List WARC record fields",
		Long:  `List information about WARC records`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o, err := flags.ToListOptions()
			if err != nil {
				return err
			}
			if err := o.Complete(cmd, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			cmd.SilenceUsage = true
			err = o.Run()
			if errors.Is(err, context.Canceled) {
				os.Exit(1)
			}
			return err
		},
		ValidArgsFunction: flag.SuffixCompletionFn,
	}

	flags.AddFlags(cmd)

	return cmd
}

// Complete completes the options
func (o *ListOptions) Complete(cmd *cobra.Command, args []string) error {
	o.paths = append(o.paths, args...)
	return nil
}

// Validate validates the options
func (o *ListOptions) Validate() error {
	if len(o.paths) == 0 {
		return errors.New("missing file or directory name")
	}
	return nil
}

// Run runs the ls command
func (o *ListOptions) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	workerPool := workerpool.New(ctx, o.concurrency)
	defer workerPool.CloseWait()

	for _, path := range o.paths {
		err := o.fileWalker.Walk(ctx, path, func(fs afero.Fs, path string, err error) error {
			if err != nil {
				return err
			}

			workerPool.Submit(func() {
				err := o.handleFile(ctx, fs, path)
				if err != nil {
					if !o.continueOnError {
						cancel()
					}
					var recordErr warc.RecordError
					if errors.As(err, &recordErr) {
						slog.Error(recordErr.Error(), "path", path, "offset", recordErr.Offset())
					} else {
						slog.Error(err.Error(), "path", path)
					}
				}
			})

			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// listFile reads a warc file and writes the records to the output
func (o *ListOptions) handleFile(ctx context.Context, fs afero.Fs, path string) error {
	f, err := fs.Open(path)
	if err != nil {
		return err
	}
	warcFileReader, err := gowarc.NewWarcFileReaderFromStream(f, o.offset, o.warcRecordOptions...)
	if err != nil {
		return err
	}
	defer func() { _ = warcFileReader.Close() }()

	var lastOffset int64 = -1

	for record, err := range warc.Records(warcFileReader, o.filter, o.recordNum, o.recordCount) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err != nil {
			// When forcing, avoid infinite loop by ensuring the iterator moves forward
			if o.force && lastOffset != record.Offset {
				slog.Warn(err.Error(), "offset", record.Offset, "path", path)
				lastOffset = record.Offset
				continue
			}
			return warc.Error(record, err)
		}
		if err := o.handleRecord(record, path); err != nil {
			return warc.Error(record, err)
		}
	}

	return nil
}

func (o *ListOptions) handleRecord(record warc.Record, path string) error {
	defer record.Close()
	if err := o.writer.WriteRecord(record, path); err != nil {
		return warc.Error(record, err)
	}
	return nil
}
