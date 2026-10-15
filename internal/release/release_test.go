package release

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"testing"
)

func TestChecksumFor(t *testing.T) {
	sums := "abc123  ghx_linux_amd64\nDEF456 *ghx_windows_amd64.exe\n"
	if got, _ := ChecksumFor(sums, "ghx_windows_amd64.exe"); got != "def456" {
		t.Fatalf("got %q", got)
	}
	if _, err := ChecksumFor(sums, "ghx_darwin_arm64"); err == nil {
		t.Fatal("expected error for unlisted file")
	}
}

func TestFindAndVersion(t *testing.T) {
	rel := Release{TagName: "v1.2.3", Assets: []Asset{{Name: "a"}}}
	if rel.Version() != "1.2.3" {
		t.Fatalf("got %s", rel.Version())
	}
	if _, ok := rel.Find("b"); ok {
		t.Fatal("found missing asset")
	}
}

func TestExtractZip(t *testing.T) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, body := range map[string]string{"gh_1_macOS_arm64/LICENSE": "l", "gh_1_macOS_arm64/bin/gh": "binary"} {
		file, _ := writer.Create(name)
		_, _ = file.Write([]byte(body))
	}
	_ = writer.Close()
	got, err := ExtractFile(buffer.Bytes(), "gh.zip", "bin/gh")
	if err != nil || string(got) != "binary" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestExtractTarGz(t *testing.T) {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	body := []byte("binary")
	_ = tarWriter.WriteHeader(&tar.Header{Name: "gh_1_linux_amd64/bin/gh", Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg})
	_, _ = tarWriter.Write(body)
	_ = tarWriter.Close()
	_ = gzipWriter.Close()
	got, err := ExtractFile(buffer.Bytes(), "gh.tar.gz", "bin/gh")
	if err != nil || string(got) != "binary" {
		t.Fatalf("got %q, %v", got, err)
	}
	if _, err := ExtractFile(buffer.Bytes(), "gh.tar.gz", "bin/missing"); err == nil {
		t.Fatal("expected error for missing entry")
	}
}

func TestExtractUnsupported(t *testing.T) {
	if _, err := ExtractFile(nil, "gh.rar", "bin/gh"); err == nil {
		t.Fatal("expected error")
	}
}
