package std

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/PondWader/kit/pkg/lang/values"
)

type readableArchive struct {
	reader *bytes.Reader
}

func (r readableArchive) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func TestArArchiveHasFile(t *testing.T) {
	archive := append([]byte("!<arch>\n"), arMember("debian-binary", []byte("2.0\n"))...)
	archive = append(archive, arMember("control.tar.zst", []byte("control"))...)

	decoded, err := arDecode(values.Of(readableArchive{reader: bytes.NewReader(archive)}))
	if err != nil {
		t.Fatalf("decode ar archive: %v", err)
	}

	archiveObj, ok := decoded.ToObject()
	if !ok {
		t.Fatal("decoded archive is not an object")
	}
	arc := archiveObj.Binding.(arArchive)

	hasControl, hasControlErr := arc.HasFile(values.Of("control.tar.zst"))
	if hasControlErr != nil {
		t.Fatalf("has_file failed: %v", hasControlErr)
	}
	if ok, _ := hasControl.ToBool(); !ok {
		t.Fatal("expected has_file to find control.tar.zst")
	}

	hasData, hasDataErr := arc.HasFile(values.Of("data.tar.xz"))
	if hasDataErr != nil {
		t.Fatalf("has_file missing check failed: %v", hasDataErr)
	}
	if ok, _ := hasData.ToBool(); ok {
		t.Fatal("expected has_file to report missing data.tar.xz")
	}

	file, fileErr := arc.File(values.Of("control.tar.zst"))
	if fileErr != nil {
		t.Fatalf("file lookup failed: %v", fileErr)
	}
	fileObj, ok := file.ToObject()
	if !ok {
		t.Fatal("file result is not an object")
	}
	buf := make([]byte, len("control"))
	if _, err := fileObj.Binding.(arFile).Read(buf); err != nil {
		t.Fatalf("read file contents failed: %v", err)
	}
	if got := string(buf); got != "control" {
		t.Fatalf("unexpected file contents: %q", got)
	}
}

func arMember(name string, body []byte) []byte {
	header := fmt.Sprintf("%-16s%-12d%-6d%-6d%-8o%-10d`\n", name, 0, 0, 0, 0o100644, len(body))
	member := append([]byte(header), body...)
	if len(body)%2 != 0 {
		member = append(member, '\n')
	}
	return member
}
