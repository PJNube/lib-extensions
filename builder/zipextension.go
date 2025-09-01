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
	"strings"
)

func getId(profile, vendor, name string) string {
	if profile == "" || vendor == "" || name == "" {
		return ""
	}
	return strings.ToLower(fmt.Sprintf("%s::%s::%s", profile, vendor, name))
}

func PackageExtension(windows bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	executableName := ExecutableName
	if windows {
		executableName += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", executableName, "main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to build utility script: %v", err)
	}

	outputFolder := path.Join(cwd, ZippedFolderName)
	outputZipPath := path.Join(outputFolder, ZipFileName)
	executablePath := filepath.Join(cwd, executableName)
	fmt.Println("Creating ZIP file...")
	err = os.Mkdir(outputFolder, 0755)
	if err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}
	file, err := os.Open(MetadataFileName)
	if err != nil {
		return fmt.Errorf("failed to open metadata file: %w", err)
	}
	defer file.Close()
	tempBytes := &bytes.Buffer{}
	_, err = file.WriteTo(tempBytes)
	if err != nil {
		return fmt.Errorf("failed to read metadata file: %w", err)
	}

	metadata := ZipMetadata{}
	err = json.Unmarshal(tempBytes.Bytes(), &metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	metadata.Id = getId(metadata.Profile, metadata.Vendor, metadata.Name)
	commentInfo, _ := json.Marshal(metadata)
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	err = addExecutableToZip(zipWriter, executablePath)
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

	err = os.WriteFile(outputZipPath, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	fmt.Printf("\n$ZIP file created at: %s\n", outputZipPath)
	return nil
}

func addExecutableToZip(zipWriter *zip.Writer, executablePath string) error {
	fileInfo, err := os.Stat(executablePath)
	if os.IsNotExist(err) {
		executablePath += ".exe"
		fileInfo, err = os.Stat(executablePath)
		if os.IsNotExist(err) {
			return fmt.Errorf("executable file does not exist: %s", executablePath)
		}
	}

	file, err := os.Open(executablePath)
	if err != nil {
		return fmt.Errorf("failed to open executable: %w", err)
	}
	defer file.Close()

	filename := filepath.Base(executablePath)
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
		return fmt.Errorf("failed to create zip entry: %w", err)
	}

	_, err = io.Copy(writer, file)
	if err != nil {
		return fmt.Errorf("failed to copy file to zip: %w", err)
	}

	return nil
}
