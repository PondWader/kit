package deb

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	arMagic     = "!<arch>\n"
	headerSize  = 60
	headerMagic = "`\n"
)

var (
	// ErrInvalidMagic is returned when the ar archive does not start with the expected "!<arch>\n" magic bytes.
	ErrInvalidMagic = errors.New("ar: invalid global header magic")
	// ErrInvalidHeader is returned when an entry header is malformed or truncated.
	ErrInvalidHeader = errors.New("ar: invalid entry header")
)

// ArHeader contains the metadata for a single member in an ar archive.
type ArHeader struct {
	Name string
	Size int64
}

// ArReader reads members from an ar archive sequentially.
// It supports both BSD and GNU long filename extensions.
type ArReader struct {
	r         io.Reader
	remaining int64
	padded    bool
	gnuNames  []byte
}

// NewArReader returns a new ArReader reading from r.
// It reads and validates the global ar header from r.
func NewArReader(r io.Reader) (*ArReader, error) {
	magic := make([]byte, len(arMagic))
	if _, err := io.ReadFull(r, magic); err != nil {
		return nil, fmt.Errorf("ar: read global header: %w", err)
	}
	if string(magic) != arMagic {
		return nil, ErrInvalidMagic
	}
	return &ArReader{r: r}, nil
}

// Next advances to the next member in the archive and returns its header.
// The ArReader then acts as an io.Reader for that member's data.
// It returns io.EOF when no more members remain.
func (ar *ArReader) Next() (*ArHeader, error) {
	for {
		// Skip remaining data from previous entry.
		if ar.remaining > 0 {
			if _, err := io.CopyN(io.Discard, ar.r, ar.remaining); err != nil {
				if errors.Is(err, io.EOF) {
					return nil, ErrInvalidHeader
				}
				return nil, err
			}
			ar.remaining = 0
		}

		// Skip padding byte from previous entry.
		if ar.padded {
			if _, err := io.CopyN(io.Discard, ar.r, 1); err != nil {
				if errors.Is(err, io.EOF) {
					return nil, ErrInvalidHeader
				}
				return nil, err
			}
			ar.padded = false
		}

		// Read the 60-byte header.
		var hdr [headerSize]byte
		n, err := io.ReadFull(ar.r, hdr[:])
		if err != nil {
			if errors.Is(err, io.EOF) && n == 0 {
				return nil, io.EOF
			}
			if errors.Is(err, io.ErrUnexpectedEOF) {
				return nil, ErrInvalidHeader
			}
			return nil, err
		}

		// Validate header magic.
		if string(hdr[58:60]) != headerMagic {
			return nil, ErrInvalidHeader
		}

		size, err := strconv.ParseInt(strings.TrimSpace(string(hdr[48:58])), 10, 64)
		if err != nil || size < 0 {
			return nil, ErrInvalidHeader
		}
		origSize := size

		rawName := strings.TrimRight(string(hdr[0:16]), " ")

		if strings.HasPrefix(rawName, "#1/") {
			nameLen, err := strconv.ParseInt(strings.TrimPrefix(rawName, "#1/"), 10, 64)
			if err != nil || nameLen < 0 || nameLen > size {
				return nil, ErrInvalidHeader
			}

			name, err := readMemberBytes(ar.r, nameLen)
			if err != nil {
				return nil, err
			}

			size -= nameLen
			ar.remaining = size
			ar.padded = origSize%2 != 0
			return &ArHeader{Name: strings.TrimRight(string(name), "\x00"), Size: size}, nil
		}

		if rawName == "//" {
			names, err := readMemberBytes(ar.r, size)
			if err != nil {
				return nil, err
			}
			ar.gnuNames = names

			if origSize%2 != 0 {
				if _, err := io.CopyN(io.Discard, ar.r, 1); err != nil {
					if errors.Is(err, io.EOF) {
						return nil, ErrInvalidHeader
					}
					return nil, err
				}
			}

			// Skip GNU long-name table pseudo-entry.
			continue
		}

		name := strings.TrimSuffix(rawName, "/")
		if strings.HasPrefix(rawName, "/") && len(rawName) > 1 {
			offset, err := strconv.ParseInt(rawName[1:], 10, 64)
			if err == nil {
				resolved, ok := resolveGNULongName(ar.gnuNames, offset)
				if !ok {
					return nil, ErrInvalidHeader
				}
				name = resolved
			}
		}

		ar.remaining = size
		ar.padded = origSize%2 != 0
		return &ArHeader{Name: name, Size: size}, nil
	}
}

// Read reads from the current member's data. It returns io.EOF when the
// member's data has been fully consumed.
func (ar *ArReader) Read(p []byte) (int, error) {
	if ar.remaining <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > ar.remaining {
		p = p[:ar.remaining]
	}
	n, err := ar.r.Read(p)
	ar.remaining -= int64(n)
	return n, err
}

func readMemberBytes(r io.Reader, n int64) ([]byte, error) {
	if n < 0 || n > int64(^uint(0)>>1) {
		return nil, ErrInvalidHeader
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(r, b); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, ErrInvalidHeader
		}
		return nil, err
	}
	return b, nil
}

func resolveGNULongName(table []byte, offset int64) (string, bool) {
	if offset < 0 || offset >= int64(len(table)) {
		return "", false
	}
	rest := table[offset:]
	end := bytes.Index(rest, []byte("/\n"))
	if end < 0 {
		end = bytes.IndexByte(rest, '\n')
	}
	if end < 0 {
		end = len(rest)
	}
	return string(rest[:end]), true
}
