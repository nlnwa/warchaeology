package cat

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nlnwa/gowarc/v2"
	"github.com/nlnwa/warchaeology/v4/cmd/internal/flag"
	"github.com/nlnwa/warchaeology/v4/internal/filewalker"
	"github.com/nlnwa/warchaeology/v4/internal/filter"
	"github.com/nlnwa/warchaeology/v4/internal/warc"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ShowWarcHeader      = "header"
	ShowWarcHeaderShort = "w"
	ShowWarcHeaderHelp  = "show WARC header"

	ShowProtocolHeader      = "protocol-header"
	ShowProtocolHeaderShort = "p"
	ShowProtocolHeaderHelp  = "show protocol header"

	ShowPayload      = "payload"
	ShowPayloadShort = "P"
	ShowPayloadHelp  = "show payload"

	Compression      = "compress"
	CompressionShort = "z"
	CompressionHelp  = "compress output (per record)"
)

type CatOptions struct {
	paths             []string
	offset            int64
	recordNum         int
	recordCount       int
	force             bool
	compress          bool
	continueOnError   bool
	filter            *filter.RecordFilter
	writer            *writer
	fileWalker        *filewalker.FileWalker
	warcRecordOptions []gowarc.WarcRecordOption
}

type CatFlags struct {
	FileWalkerFlags       flag.FileWalkerFlags
	FilterFlags           flag.FilterFlags
	WarcIteratorFlags     flag.WarcIteratorFlags
	WarcRecordOptionFlags flag.WarcRecordOptionFlags
	ErrorFlags            flag.ErrorFlags
}

func (f CatFlags) AddFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	f.FileWalkerFlags.AddFlags(cmd)
	f.FilterFlags.AddFlags(cmd)
	f.WarcIteratorFlags.AddFlags(cmd)
	f.WarcRecordOptionFlags.AddFlags(cmd)
	f.ErrorFlags.AddFlags(cmd)

	flags.BoolP(ShowWarcHeader, ShowWarcHeaderShort, false, ShowWarcHeaderHelp)
	flags.BoolP(ShowProtocolHeader, ShowProtocolHeaderShort, false, ShowProtocolHeaderHelp)
	flags.BoolP(ShowPayload, ShowPayloadShort, false, ShowPayloadHelp)
	flags.BoolP(Compression, CompressionShort, false, CompressionHelp)
}

func (f CatFlags) ShowWarcHeader() bool {
	return viper.GetBool(ShowWarcHeader)
}

func (f CatFlags) ShowProtocolHeader() bool {
	return viper.GetBool(ShowProtocolHeader)
}

func (f CatFlags) ShowPayload() bool {
	return viper.GetBool(ShowPayload)
}

func (f CatFlags) Compress() bool {
	return viper.GetBool("compress")
}

func (f CatFlags) ToOptions() (*CatOptions, error) {
	filter, err := f.FilterFlags.ToFilter()
	if err != nil {
		return nil, err
	}

	fileList, err := flag.ReadSrcFileList(f.FileWalkerFlags.SrcFileListFlags.SrcFileList())
	if err != nil {
		return nil, fmt.Errorf("failed to read from source file: %w", err)
	}

	fileWalker, err := f.FileWalkerFlags.ToFileWalker()
	if err != nil {
		return nil, fmt.Errorf("failed to create file walker: %w", err)
	}

	writer := &writer{
		showWarcHeader:     f.ShowWarcHeader(),
		showProtocolHeader: f.ShowProtocolHeader(),
		showPayload:        f.ShowPayload(),
	}

	return &CatOptions{
		paths:             fileList,
		offset:            f.WarcIteratorFlags.Offset(),
		recordCount:       f.WarcIteratorFlags.Limit(),
		recordNum:         f.WarcIteratorFlags.RecordNum(),
		force:             f.WarcIteratorFlags.Force(),
		filter:            filter,
		continueOnError:   f.ErrorFlags.ContinueOnError(),
		fileWalker:        fileWalker,
		compress:          f.Compress(),
		writer:            writer,
		warcRecordOptions: f.WarcRecordOptionFlags.ToWarcRecordOptions(),
	}, nil
}

func NewCmdCat() *cobra.Command {
	flags := CatFlags{}

	cmd := &cobra.Command{
		Use:   "cat FILE/DIR ...",
		Short: "Concatenate and print warc files",
		Long:  ``,
		Example: `
# Print all content from a WARC file (in principle the same as zcat)
warc cat file1.warc.gz

# Pipe the payload of the 4th record into the image viewer feh
warc cat -n4 -P file1.warc.gz | feh -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o, err := flags.ToOptions()
			if err != nil {
				return err
			}
			err = o.Complete(cmd, args)
			if err != nil {
				return err
			}
			err = o.Validate()
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			err = o.Run()
			if errors.Is(err, context.Canceled) {
				os.Exit(1)
			}
			return err
		},
	}

	flags.AddFlags(cmd)

	return cmd
}

func (o *CatOptions) Complete(cmd *cobra.Command, args []string) error {
	o.paths = append(o.paths, args...)

	// If no output is specified, show everything.
	// This way we can specify a single flag to show just that part of the record.
	if !(o.writer.showWarcHeader || o.writer.showProtocolHeader || o.writer.showPayload) {
		o.writer.showWarcHeader = true
		o.writer.showProtocolHeader = true
		o.writer.showPayload = true
	}

	return nil
}

// Validate validates the options
func (o *CatOptions) Validate() error {
	if len(o.paths) == 0 {
		return errors.New("missing file or directory")
	}
	return nil
}

// Run runs the cat command
func (o *CatOptions) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	for _, path := range o.paths {
		err := o.fileWalker.Walk(ctx, path, func(fs afero.Fs, path string, err error) error {
			if err != nil {
				return err
			}

			err = o.handleFile(ctx, fs, path)
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

			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// handleFile reads a WARC file and writes the content to stdout
func (o *CatOptions) handleFile(ctx context.Context, fs afero.Fs, path string) error {
	f, err := fs.Open(path)
	if err != nil {
		return err
	}
	warcFileReader, err := gowarc.NewWarcFileReaderFromStream(f, o.offset, o.warcRecordOptions...)
	if err != nil {
		return err
	}
	defer func() {
		_ = warcFileReader.Close()
	}()

	var lastOffset int64 = -1

	for record, err := range warc.Records(warcFileReader, o.filter, o.recordNum, o.recordCount) {
		if ctx.Err() != nil {
			return ctx.Err()
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
		if err := o.handleRecord(record); err != nil {
			return warc.Error(record, err)
		}
	}
	return nil
}

func (o *CatOptions) handleRecord(record warc.Record) error {
	defer record.Close()
	var w io.Writer
	if o.compress {
		gw := gzip.NewWriter(os.Stdout)
		defer gw.Close()
		w = gw
	} else {
		w = os.Stdout
	}
	return o.writer.writeWarcRecord(w, record.WarcRecord)
}

type writer struct {
	showWarcHeader     bool
	showProtocolHeader bool
	showPayload        bool
}

const CRLF = "\r\n"

// writeWarcRecord writes a WARC record to the writer
func (c *writer) writeWarcRecord(w io.Writer, warcRecord gowarc.WarcRecord) error {
	if c.showWarcHeader {
		// Write WARC record version
		_, err := fmt.Fprintf(w, "%v%s", warcRecord.Version(), CRLF)
		if err != nil {
			return fmt.Errorf("failed to write WARC record version: %w", err)
		}

		// Write WARC header
		_, err = warcRecord.WarcHeader().Write(w)
		if err != nil {
			return fmt.Errorf("failed to write WARC header: %w", err)
		}

		// Write newline
		_, err = w.Write([]byte(CRLF))
		if err != nil {
			return fmt.Errorf("failed to write separator: %w", err)
		}
	}

	if c.showProtocolHeader {
		if headerBlock, ok := warcRecord.Block().(gowarc.ProtocolHeaderBlock); ok {
			_, err := w.Write(headerBlock.ProtocolHeaderBytes())
			if err != nil {
				return fmt.Errorf("failed to write protocol header: %w", err)
			}
		}
	}

	if c.showPayload {
		if payloadBlock, ok := warcRecord.Block().(gowarc.PayloadBlock); ok {
			reader, err := payloadBlock.PayloadBytes()
			if err != nil {
				return fmt.Errorf("failed to read payload: %w", err)
			}
			_, err = io.Copy(w, reader)
			if err != nil {
				return fmt.Errorf("failed to write payload: %w", err)
			}
		} else {
			reader, err := warcRecord.Block().RawBytes()
			if err != nil {
				return fmt.Errorf("failed to write raw bytes of record block: %w", err)
			}
			_, err = io.Copy(w, reader)
			if err != nil {
				return fmt.Errorf("failed to write raw bytes of record block: %w", err)
			}
		}
	}

	// Write end of record separator
	_, err := w.Write([]byte(CRLF + CRLF))
	if err != nil {
		return fmt.Errorf("failed to write end of record separator: %w", err)
	}

	return nil
}
