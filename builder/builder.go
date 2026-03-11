package builder

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/PJNube/lib-extensions/manifest"
)

type Opts struct {
	Architecture string
	ConfigDir    string
}

func addFilesToZip(zipWriter *zip.Writer, paths ...string) error {
	for _, path := range paths {
		err := addFileToZip(zipWriter, path)
		if err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) && filepath.Ext(filePath) != ".exe" {
		filePath += ".exe"
		fileInfo, err = os.Stat(filePath)
	}

	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filepath.Base(filePath), err)
	}
	defer file.Close()

	filename := filepath.Base(filePath)

	header := &zip.FileHeader{
		Name:   filename,
		Method: zip.Deflate,
	}

	mode := fileInfo.Mode()
	if mode&0111 == 0 {
		mode |= 0755
	}
	header.SetMode(mode)

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip entry for %s: %w", filename, err)
	}

	_, err = io.Copy(writer, file)
	if err != nil {
		return fmt.Errorf("failed to copy file %s to zip: %w", filename, err)
	}

	return nil
}

func getBuildTime() string {
	currentTime := time.Now().UTC()
	return currentTime.Format(time.RFC3339)
}

func addMetadataToZip(zipWriter *zip.Writer, metadata *manifest.Metadata) error {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(metadata); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}
	data := buf.Bytes()
	data = bytes.TrimSuffix(data, []byte{'\n'})

	header := &zip.FileHeader{
		Name:   manifest.MetadataFileName,
		Method: zip.Deflate,
	}
	header.SetMode(0644)

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip entry for %s: %w", manifest.MetadataFileName, err)
	}

	if _, err = writer.Write(data); err != nil {
		return fmt.Errorf("failed to write metadata to zip: %w", err)
	}

	return nil
}
