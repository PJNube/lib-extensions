package builder

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/PJNube/lib-extensions/manifest"
	"github.com/PJNube/lib-extensions/naming"
)

const ChecksumFileName = "checksums.sha256"

const (
	ExecutableName   = "extension"
	ZippedFolderName = "out"
	BuildPath        = "executable"
)

// PackageExtension packages the extension into a zip file.
// The 'arch' parameter is optional and mainly used for testing.
func PackageExtension(arch ...string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	executablePath := path.Join(BuildPath, ExecutableName)
	cmd := exec.Command(
		"go", "build",
		"-trimpath",
		"-ldflags=-s -w",
		"-o", executablePath,
		"main.go",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to build utility script: %v", err)
	}

	fmt.Println("Creating ZIP file...")
	outputFolder := path.Join(cwd, ZippedFolderName)
	err = os.Mkdir(outputFolder, 0755)
	if err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	metadata, err := manifest.GetMetadata()
	if err != nil {
		return err
	}

	metadata.BuildTime = getBuildTime()
	if len(arch) > 0 {
		metadata.Dependencies.Architecture = arch[0]
	} else {
		metadata.Dependencies.Architecture = runtime.GOARCH
	}

	executableFullPath := filepath.Join(cwd, executablePath)
	metadataFilePath := path.Join(cwd, manifest.MetadataFileName)

	filePaths := []string{executableFullPath}
	for _, schema := range metadata.OpenAPISchemas {
		filePaths = append(filePaths, schema.Path)
	}

	allFiles := append(filePaths, metadataFilePath)
	for _, filePath := range allFiles {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("required file %s does not exist", filePath)
		}
	}

	commentInfo, _ := json.Marshal(metadata)

	out, err := os.Create(ChecksumFileName)
	if err != nil {
		return err
	}
	defer out.Close()

	err = GenerateChecksumsFile(out, commentInfo, filePaths...)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Add enriched metadata content directly to the zip (same content as the comment)
	err = addContentToZip(zipWriter, manifest.MetadataFileName, commentInfo)
	if err != nil {
		return err
	}

	filePaths = append(filePaths, ChecksumFileName)
	err = addFilesToZip(zipWriter, filePaths...)
	if err != nil {
		return err
	}

	err = zipWriter.SetComment(string(commentInfo))
	if err != nil {
		return fmt.Errorf("failed to set zip comment: %w", err)
	}

	err = zipWriter.Close()
	if err != nil {
		return err
	}

	outputZipPath := getOutputZipPath(metadata, outputFolder)
	err = os.WriteFile(outputZipPath, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	err = os.RemoveAll(BuildPath)
	if err != nil {
		fmt.Println("Warning: failed to remove temporary executable:", err)
	}

	_ = os.Remove(ChecksumFileName)

	fmt.Printf("ZIP file created at: %s\n", outputZipPath)
	return nil
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

func addContentToZip(zipWriter *zip.Writer, filename string, content []byte) error {
	header := &zip.FileHeader{
		Name:   filename,
		Method: zip.Deflate,
	}
	header.SetMode(0644)

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = writer.Write(content)
	if err != nil {
		return err
	}

	return nil
}

func getOutputZipPath(metadata *manifest.Metadata, outputFolder string) string {
	id := naming.GetId(metadata.Profile, metadata.Vendor, metadata.Name)
	zipFileName := strings.Join([]string{id, metadata.Version, metadata.Dependencies.Architecture}, naming.IdSeparator)
	return path.Join(outputFolder, strings.Join([]string{zipFileName, ".zip"}, ""))
}
