/*
 * Copyright 2021 National Library of Norway.
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
 */

package ls

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal"
)

type RecordWriter interface {
	Write(wr gowarc.WarcRecord, fileName string, offset int64) error
}

type CdxLegacy struct {
}

type CdxJ struct {
}

func (c *CdxLegacy) Write(wr gowarc.WarcRecord, fileName string, offset int64) error {
	return nil
}

func (c *CdxJ) Write(wr gowarc.WarcRecord, fileName string, offset int64) error {
	if wr.Type() != gowarc.Warcinfo {
		rec := internal.NewCdxRecord(wr, fileName, offset)
		cdxj, err := json.Marshal(rec)
		if err != nil {
			return err
		}
		fmt.Printf("%s %s %s %s\n", rec.Ssu, rec.Sts, rec.Srt, cdxj)
	}
	return nil
}

// filter bitmask
const (
	None       = 0b0000
	Offset     = 0b0001
	RecordID   = 0b0010
	RecordType = 0b0100
	TargetURI  = 0b1000
	All        = 0b1111
)

type defaultWriterBuilder struct {
	// a opt-out bitmask filter to stop the writer from printing certain record information
	filter int8
}

func NewDefaultWriterBuilder() DefaultWriterBuilder {
	return &defaultWriterBuilder{None}
}

func (d *defaultWriterBuilder) FilterOffset(enabled bool) DefaultWriterBuilder {
	d.filter |= getFilterOrNone(Offset, enabled)
	return d
}

func (d *defaultWriterBuilder) FilterRecordID(enabled bool) DefaultWriterBuilder {
	d.filter |= getFilterOrNone(RecordID, enabled)
	return d
}

func (d *defaultWriterBuilder) FilterRecordType(enabled bool) DefaultWriterBuilder {
	d.filter |= getFilterOrNone(RecordType, enabled)
	return d
}

func (d *defaultWriterBuilder) FilterRecordTargetURI(enabled bool) DefaultWriterBuilder {
	d.filter |= getFilterOrNone(TargetURI, enabled)
	return d
}

func (d *defaultWriterBuilder) Build() DefaultWriter {
	return DefaultWriter{
		filter: d.filter,
	}
}

func getFilterOrNone(value int8, enabled bool) int8 {
	if enabled {
		return value
	}
	return None
}

type DefaultWriterBuilder interface {
	FilterOffset(enabled bool) DefaultWriterBuilder
	FilterRecordID(enabled bool) DefaultWriterBuilder
	FilterRecordType(enabled bool) DefaultWriterBuilder
	FilterRecordTargetURI(enabled bool) DefaultWriterBuilder
	Build() DefaultWriter
}

type DefaultWriter struct {
	// a opt-out filter to stop the writer from printing certain record information
	filter int8
}

func (d DefaultWriter) Write(wr gowarc.WarcRecord, fileName string, offset int64) error {
	if d.filter == All {
		// probably a user error since the user is requesting a NOP
		return errors.New("filtering all record information")
	}

	var sb strings.Builder
	if d.filter&Offset != Offset {
		sb.WriteString(fmt.Sprintf("%11d ", offset))
	}
	if d.filter&RecordID != RecordID {
		recordID := wr.WarcHeader().Get(gowarc.WarcRecordID)
		sb.WriteString(fmt.Sprintf("%s ", recordID))
	}
	if d.filter&RecordType != RecordType {
		sb.WriteString(fmt.Sprintf("%-9.9s ", wr.Type()))
	}
	if d.filter&TargetURI != TargetURI {
		targetURI := internal.CropString(wr.WarcHeader().Get(gowarc.WarcTargetURI), 100)
		sb.WriteString(fmt.Sprintf("%s", targetURI))
	}
	sb.WriteString("\n")
	fmt.Print(sb.String())
	return nil
}
