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
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/filter"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/nlnwa/warchaeology/internal/warc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type conf struct {
	offset    int64
	recordNum int
	filter    *filter.Filter
}

func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:    "aart",
		Short:  "Show images",
		Long:   ``,
		Hidden: true,
		RunE:   parseArgumentsAndCallAsciiArt}

	cmd.Flags().IntP("width", "w", 100, "Width of image")

	return cmd
}

func parseArgumentsAndCallAsciiArt(cmd *cobra.Command, args []string) error {
	config := &conf{}
	if len(args) == 0 {
		return errors.New("missing file name")
	}
	fileName := args[0]
	config.offset = viper.GetInt64(flag.Offset)
	config.recordNum = viper.GetInt(flag.RecordNum)

	if config.offset < 0 {
		config.offset = 0
	}

	viper.Set(flag.MimeType, []string{"image/gif", "image/jpeg", "image/png"})
	config.filter = filter.NewFromViper()

	readFile(config, fileName)
	return nil

}

func readFile(c *conf, fileName string) {
	warcFileReader, err := gowarc.NewWarcFileReader(fileName, c.offset, gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)))
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer func() { _ = warcFileReader.Close() }()

	records := make(chan warc.Record)

	iterator := warc.Iterator{
		WarcFileReader: warcFileReader,
		Filter:         c.filter,
		Nth:            c.recordNum,
		Limit:          0,
		Records:        records,
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	go iterator.Iterate(ctx)

	count := 0

	for record := range records {
		count++

		fmt.Println("Record number:", count)
		if record.Err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error in record at offset %v: %v\n", c.offset, record.Err)
			break
		}

		warcRecord := record.WarcRecord

		b, ok := warcRecord.Block().(gowarc.HttpResponseBlock)
		if !ok {
			continue
		}

		fmt.Printf("\u001B[2J\u001B[HUrl: %s\n\n", warcRecord.WarcHeader().Get(gowarc.WarcTargetURI))
		r, err := b.PayloadBytes()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		err = display(r, viper.GetInt("width"))
		if err != nil {
			fmt.Println("Couldn't decode image,\nError:", err.Error())
			continue
		}

		fmt.Printf("Hit enter to continue\n")
		_, _ = fmt.Scanln()
	}
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
