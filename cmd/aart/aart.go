package aart

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"

	"github.com/nfnt/resize"
	"github.com/nlnwa/gowarc/v2"
	"github.com/nlnwa/warchaeology/v3/cmd/internal/flag"
	"github.com/nlnwa/warchaeology/v3/internal/filter"
	"github.com/nlnwa/warchaeology/v3/internal/warc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type AartOptions struct {
	offset            int64
	recordNum         int
	recordCount       int
	filter            *filter.RecordFilter
	fileName          string
	warcRecordOptions []gowarc.WarcRecordOption
	width             int
}

type AartFlags struct {
	FilterFlags           flag.FilterFlags
	WarcIteratorFlags     flag.WarcIteratorFlags
	WarcRecordOptionFlags flag.WarcRecordOptionFlags
}

func (f AartFlags) AddFlags(cmd *cobra.Command) {
	f.FilterFlags.AddFlags(cmd, flag.WithDefaultMimeType([]string{"image/jpeg", "image/png", "image/gif"}))
	f.WarcIteratorFlags.AddFlags(cmd)
	f.WarcRecordOptionFlags.AddFlags(cmd)

	cmd.Flags().IntP("width", "w", 100, "Width of image")
}

func (f AartFlags) Width() int {
	return viper.GetInt("width")
}

func (f AartFlags) ToOptions() (*AartOptions, error) {
	filter, err := f.FilterFlags.ToFilter()
	if err != nil {
		return nil, err
	}

	warcRecordOptions := f.WarcRecordOptionFlags.ToWarcRecordOptions()

	return &AartOptions{
		filter:            filter,
		offset:            f.WarcIteratorFlags.Offset(),
		recordNum:         f.WarcIteratorFlags.RecordNum(),
		recordCount:       f.WarcIteratorFlags.Limit(),
		width:             f.Width(),
		warcRecordOptions: warcRecordOptions,
	}, nil
}

func NewCmdAart() *cobra.Command {
	flags := AartFlags{}

	cmd := &cobra.Command{
		Use:    "aart",
		Short:  "Show images",
		Long:   ``,
		Hidden: true,
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
			return o.Run()
		},
	}

	flags.AddFlags(cmd)

	return cmd
}

func (o *AartOptions) Complete(cmd *cobra.Command, args []string) error {
	o.fileName = args[0]

	return nil

}
func (o *AartOptions) Validate() error {
	if len(o.fileName) == 0 {
		return errors.New("missing file name")
	}
	return nil
}

func (o *AartOptions) Run() error {
	wf, err := gowarc.NewWarcFileReader(o.fileName, o.offset, o.warcRecordOptions...)
	defer func() {
		if wf != nil {
			_ = wf.Close()
		}
	}()
	if err != nil {
		return fmt.Errorf("failed to create WARC reader: %v", err)
	}

	for record := range warc.NewIterator(context.TODO(), wf, o.filter, o.recordNum, o.recordCount) {
		if record.Err != nil {
			return fmt.Errorf("failed reading records: %w", record.Err)
		}
		wr := record.WarcRecord

		if b, ok := wr.Block().(gowarc.HttpResponseBlock); ok {
			fmt.Printf("\u001B[2J\u001B[HUrl: %s\n\n", wr.WarcHeader().Get(gowarc.WarcTargetURI))
			r, err := b.PayloadBytes()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get payload bytes: %v\n", err)
			}
			err = display(r, o.width)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to display: %v\n", err)
				continue
			}
			fmt.Printf("Hit enter to continue\n")
			_, _ = fmt.Scanln()
		}
	}
	return nil
}

var asciiChar = "MND8OZ$7I?+=~:,.."

func asciiArt(img image.Image, w, h int) []byte {
	table := []byte(asciiChar)
	buffer := new(bytes.Buffer)
	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			g := grayscale(img.At(j, i))
			pos := len(asciiChar) * g / 65536
			_ = buffer.WriteByte(table[pos])
		}
		_ = buffer.WriteByte('\n')
	}
	return buffer.Bytes()
}

func grayscale(c color.Color) int {
	r, g, b, _ := c.RGBA()
	return int(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))
}

func getHeight(img image.Image, w int) (image.Image, int, int) {
	sz := img.Bounds()
	height := (sz.Max.Y * w * 10) / (sz.Max.X * 16)
	img = resize.Resize(uint(w), uint(height), img, resize.Lanczos3)
	return img, w, height
}

func display(r io.Reader, width int) error {
	img, _, err := image.Decode(r)
	if err != nil {
		return err
	}

	finalASCIIArt := asciiArt(getHeight(img, width))
	fmt.Println(string(finalASCIIArt))
	return nil
}
