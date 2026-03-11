package kit

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/PondWader/kit/pkg/lang/values"
	"github.com/klauspost/compress/zstd"
)

type readableBuffer struct {
	reader *bytes.Reader
}

func (r readableBuffer) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func TestTarLayerOpenNormalizesControlPath(t *testing.T) {
	var archive bytes.Buffer
	tw := tar.NewWriter(&archive)
	if err := tw.WriteHeader(&tar.Header{Name: "./control", Mode: 0o644, Size: int64(len("Package: demo\n"))}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := tw.Write([]byte("Package: demo\n")); err != nil {
		t.Fatalf("write body: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}

	var compressed bytes.Buffer
	gz := gzip.NewWriter(&compressed)
	if _, err := gz.Write(archive.Bytes()); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}

	layer := tarLayer{newReader: func(r io.Reader) (io.Reader, error) {
		return gzip.NewReader(r)
	}}
	opened, err := layer.Open(values.Of(readableBuffer{reader: bytes.NewReader(compressed.Bytes())}))
	if err != nil {
		t.Fatalf("open tar archive: %v", err)
	}

	archiveObj, ok := opened.ToObject()
	if !ok {
		t.Fatal("opened archive is not an object")
	}
	tarArchive := archiveObj.Binding.(tarArchive)

	hasFile, hasFileErr := tarArchive.HasFile(values.Of("control"))
	if hasFileErr != nil {
		t.Fatalf("has_file failed: %v", hasFileErr)
	}
	if ok, _ := hasFile.ToBool(); !ok {
		t.Fatal("expected has_file to find control")
	}

	fileValue, fileErr := tarArchive.File(values.Of("control"))
	if fileErr != nil {
		t.Fatalf("file lookup failed: %v", fileErr)
	}
	fileObj, ok := fileValue.ToObject()
	if !ok {
		t.Fatal("tar file is not an object")
	}
	file := fileObj.Binding.(tarFile)
	text, textErr := file.Text()
	if textErr != nil {
		t.Fatalf("read text failed: %v", textErr)
	}
	textStr, ok := text.ToString()
	if !ok {
		t.Fatal("text result is not a string")
	}
	if got := textStr.String(); got != "Package: demo\n" {
		t.Fatalf("unexpected control contents: %q", got)
	}
}

func TestTarLayerOpenSupportsZstd(t *testing.T) {
	var archive bytes.Buffer
	tw := tar.NewWriter(&archive)
	if err := tw.WriteHeader(&tar.Header{Name: "control", Mode: 0o644, Size: int64(len("Version: 1\n"))}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := tw.Write([]byte("Version: 1\n")); err != nil {
		t.Fatalf("write body: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}

	var compressed bytes.Buffer
	zw, err := zstd.NewWriter(&compressed)
	if err != nil {
		t.Fatalf("create zstd writer: %v", err)
	}
	if _, err := zw.Write(archive.Bytes()); err != nil {
		t.Fatalf("zstd write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zstd writer: %v", err)
	}

	layer := tarLayer{newReader: func(r io.Reader) (io.Reader, error) {
		return zstd.NewReader(r)
	}}
	opened, openErr := layer.Open(values.Of(readableBuffer{reader: bytes.NewReader(compressed.Bytes())}))
	if openErr != nil {
		t.Fatalf("open zstd tar archive: %v", openErr)
	}

	archiveObj, ok := opened.ToObject()
	if !ok {
		t.Fatal("opened archive is not an object")
	}
	fileValue, fileErr := archiveObj.Binding.(tarArchive).File(values.Of("control"))
	if fileErr != nil {
		t.Fatalf("file lookup failed: %v", fileErr)
	}
	fileObj, ok := fileValue.ToObject()
	if !ok {
		t.Fatal("tar file is not an object")
	}
	text, textErr := fileObj.Binding.(tarFile).Text()
	if textErr != nil {
		t.Fatalf("read text failed: %v", textErr)
	}
	textStr, ok := text.ToString()
	if !ok || textStr.String() != "Version: 1\n" {
		t.Fatalf("unexpected control text: %#v", text)
	}
}
