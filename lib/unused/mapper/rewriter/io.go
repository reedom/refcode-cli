package refcode

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"regexp"
)

// CountMatch counts how many times the specified regexp matches
// against the content in the read stream.
func CountMatch(ctx context.Context, r io.Reader, re *regexp.Regexp) (c int, err error) {
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
				//Verbose.Printf("skip binary file %q", c.filepath)
				err = ErrBinaryFile
				return
			}
			if re.Match(s.Bytes()) {
				c++
			}
		}
	}
	return
}

type transformContentFunc func([]byte) ([]byte, error)

// TransformContent transforms the content of the reading stream
// through regexp replace function and writes the result into the
// write stream.
func TransformContent(ctx context.Context, r io.Reader, w io.Writer, re *regexp.Regexp, repl transformContentFunc) error {
	var err error
	fn := func(line []byte) []byte {
		var ret []byte
		ret, err = repl(line)
		return ret
	}

	s := bufio.NewScanner(r)
	s.Split(scanLines)
	for s.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			line := re.ReplaceAllFunc(s.Bytes(), fn)
			if err != nil {
				return err
			}
			if _, err = w.Write(line); err != nil {
				return err
			}
		}
	}

	return nil
}

// scanLines is a split function for a Scanner and modified version of
// bufio.ScanLines. The difference is that this won't stlip
// end-of-line markers.
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); 0 <= i {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
