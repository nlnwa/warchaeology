/*
 * Copyright 2023 National Library of Norway.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package aart

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/nfnt/resize"
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/filter"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"strconv"
)

type conf struct {
	offset    int64
	recordNum int
	filter    *filter.Filter
}

func NewCommand() *cobra.Command {
	c := &conf{}
	var cmd = &cobra.Command{
		Use:    "aart",
		Short:  "Show images",
		Long:   ``,
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing file name")
			}
			fileName := args[0]
			c.offset = viper.GetInt64(flag.Offset)
			c.recordNum = viper.GetInt(flag.RecordNum)

			if c.offset < 0 {
				c.offset = 0
			}

			viper.Set(flag.MimeType, []string{"image/gif", "image/jpeg", "image/png"})
			c.filter = filter.NewFromViper()

			readFile(c, fileName)
			return nil
		},
	}

	cmd.Flags().IntP("width", "w", 100, "Width of image")

	return cmd
}

func readFile(c *conf, fileName string) {
	wf, err := gowarc.NewWarcFileReader(fileName, c.offset, gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)))
	defer func() { _ = wf.Close() }()
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}

	num := 0
	count := 0

	for {
		wr, _, _, err := wf.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v, rec num: %v, Offset %v\n", err.Error(), strconv.Itoa(count), c.offset)
			break
		}

		if !c.filter.Accept(wr) {
			continue
		}

		// Find record number
		if c.recordNum > 0 && num < c.recordNum {
			num++
			continue
		}

		if b, ok := wr.Block().(gowarc.HttpResponseBlock); ok {
			fmt.Printf("\u001B[2J\u001B[HUrl: %s\n\n", wr.WarcHeader().Get(gowarc.WarcTargetURI))
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
