package widget

import (
	"io"
)

const (
	reset = "\x1b[0m"
	green = "\x1b[32m"
)

var (
	repCR = append(append(append([]byte{}, green...), '\\', 'r'), reset...)
	repNL = append(append(append([]byte{}, green...), '\\', 'n', '\n'), reset...)
)

type visibleLineEndingReader struct {
	r io.Reader

	in      []byte
	out     []byte
	pending []byte

	repCR []byte
	repNL []byte

	upErr error
}

func newVisibleFilter(r io.Reader) *visibleLineEndingReader {
	return newVisibleLineEndingReader(r, repCR, repNL, 32*1024)
}

func newVisibleLineEndingReader(r io.Reader, repCR, repNL []byte, chunk int) *visibleLineEndingReader {
	if chunk <= 0 {
		chunk = 32 * 1024
	}
	return &visibleLineEndingReader{
		r:     r,
		in:    make([]byte, chunk),
		out:   make([]byte, 0, chunk+64),
		repCR: repCR,
		repNL: repNL,
	}
}

func (t *visibleLineEndingReader) Read(p []byte) (int, error) {
	if len(t.pending) > 0 {
		n := copy(p, t.pending)
		t.pending = t.pending[n:]
		if len(t.pending) == 0 && t.upErr != nil {
			err := t.upErr
			t.upErr = nil
			if n > 0 {
				return n, nil
			}
			return 0, err
		}
		return n, nil
	}

	if t.upErr != nil {
		err := t.upErr
		t.upErr = nil
		return 0, err
	}

	nIn, err := t.r.Read(t.in)
	if nIn == 0 {
		return 0, err
	}

	t.out = t.out[:0]
	in := t.in[:nIn]
	for _, b := range in {
		switch b {
		case '\r':
			t.out = append(t.out, t.repCR...)
		case '\n':
			t.out = append(t.out, t.repNL...)
		default:
			t.out = append(t.out, b)
		}
	}

	t.pending = t.out

	n := copy(p, t.pending)
	t.pending = t.pending[n:]

	if err != nil {
		t.upErr = err
	}
	return n, nil
}
