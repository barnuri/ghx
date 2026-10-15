package release

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
)

// ExtractFile returns the first entry whose path ends with suffix, from a .zip or .tar.gz.
func ExtractFile(archive []byte, archiveName string, suffix string) ([]byte, error) {
	if strings.HasSuffix(archiveName, ".zip") {
		return extractZip(archive, suffix)
	}
	if strings.HasSuffix(archiveName, ".tar.gz") {
		return extractTarGz(archive, suffix)
	}
	return nil, fmt.Errorf("unsupported archive %s", archiveName)
}

func extractZip(archive []byte, suffix string) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, err
	}
	for _, entry := range reader.File {
		if !matches(entry.Name, suffix) {
			continue
		}
		file, err := entry.Open()
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return io.ReadAll(file)
	}
	return nil, fmt.Errorf("%s not found in archive", suffix)
}

func extractTarGz(archive []byte, suffix string) ([]byte, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("%s not found in archive", suffix)
		}
		if err != nil {
			return nil, err
		}
		if header.Typeflag == tar.TypeReg && matches(header.Name, suffix) {
			return io.ReadAll(tarReader)
		}
	}
}

func matches(entryName string, suffix string) bool {
	cleaned := path.Clean(strings.ReplaceAll(entryName, `\`, "/"))
	return cleaned == suffix || strings.HasSuffix(cleaned, "/"+suffix)
}
