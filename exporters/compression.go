package exporters

import (
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	None = "none"
	GZIP = "gzip"
	ZIP  = "zip"
)

type compositeWriteCloser struct {
	io.Writer
	closeFunc func() error
}

// Close implements io.WriteCloser.
func (c *compositeWriteCloser) Close() error {
	if c.closeFunc == nil {
		return nil
	}
	return c.closeFunc()
}

func createOutputWriter(path string, options ExportOptions, format string) (io.WriteCloser, error) {
	compression := strings.ToLower(strings.TrimSpace(options.Compression))

	switch compression {
	case None:
		file, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("error creating file: %w", err)
		}
		return file, nil

	case GZIP:
		if !strings.HasSuffix(strings.ToLower(path), ".gz") {
			path += ".gz"
		}
		file, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("error creating file: %w", err)
		}
		gzipWriter := gzip.NewWriter(file)
		return &compositeWriteCloser{
			Writer: gzipWriter,
			closeFunc: func() error {
				var err error
				if cerr := gzipWriter.Close(); cerr != nil {
					err = cerr
				}
				if ferr := file.Close(); ferr != nil && err == nil {
					err = ferr
				}
				return err
			},
		}, nil

	case ZIP:
		if !strings.HasSuffix(strings.ToLower(path), ".zip") {
			path += ".zip"
		}
		file, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("error creating file: %w", err)
		}
		zipWriter := zip.NewWriter(file)
		entryName := determineZipEntryName(path, format)
		entryWriter, err := zipWriter.Create(entryName)
		if err != nil {
			zipWriter.Close()
			file.Close()
			return nil, fmt.Errorf("error creating zip entry: %w", err)
		}
		return &compositeWriteCloser{
			Writer: entryWriter,
			closeFunc: func() error {
				var err error
				if cerr := zipWriter.Close(); cerr != nil {
					err = cerr
				}
				if ferr := file.Close(); ferr != nil && err == nil {
					err = ferr
				}
				return err
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported compression: %s", options.Compression)
	}
}

func determineZipEntryName(outputPath, format string) string {
	base := filepath.Base(outputPath)
	lowerBase := strings.ToLower(base)

	name := strings.TrimSuffix(lowerBase, ".zip")

	if name == "" {
		name = "export"
	}

	if !strings.HasSuffix(name, "."+format) {
		name = fmt.Sprintf("%s.%s", name, format)
	}

	return name
}
