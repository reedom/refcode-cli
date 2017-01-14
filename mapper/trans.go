package refcode

import (
	"bufio"
	"bytes"
	"context"
	"io"
)

// CountMarkerInContent counts how many times the specified marker
// found in the content.
func CountMarkerInContent(ctx context.Context, r io.Reader, marker []byte) (c int, err error) {
	l := len(marker)
	if l == 0 {
		panic("CountMarkerInContent() arg marker is empty")
	}

	s := bufio.NewScanner(r)
	s.Split(scanLines)

	for s.Scan() {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		default:
			b := s.Bytes()
			if 0 <= bytes.IndexByte(b, 0x00) {
				err = ErrBinaryFile
				return
			}
			for {
				i := bytes.Index(b, marker)
				if i < 0 {
					break
				}
				c++
				b = b[i+l:]
			}
		}
	}
	return
}

type TransFn func(ctx context.Context) ([]byte, error)

func TransformContent(ctx context.Context, r io.Reader, w io.Writer, marker []byte, fn TransFn) error {
	l := len(marker)
	if l == 0 {
		panic("TransformContent() arg marker is empty")
	}
	if fn == nil {
		panic("TransformContent() arg fn is nil")
	}

	s := bufio.NewScanner(r)
	s.Split(scanLines)
	for s.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			b := s.Bytes()
			for {
				i := bytes.Index(b, marker)
				if i < 0 {
					if _, err := w.Write(b); err != nil {
						return err
					}
					break
				}

				_, err := w.Write(b[0:i])
				if err != nil {
					return err
				}
				data, err := fn(ctx)
				if err != nil {
					return err
				}
				_, err = w.Write(data)
				if err != nil {
					return err
				}
				b = b[i+l:]
				if len(b) == 0 {
					break
				}
			}
		}
	}

	return nil
}
