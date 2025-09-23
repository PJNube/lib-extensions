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
	"time"
)

const IdSeparator = "-"

func GetId(profile, vendor, name string) string {
	if profile == "" {
		profile = "na"
		fmt.Println("WARNING: profile is empty...")
	}
	if vendor == "" {
		vendor = "na"
		fmt.Println("WARNING: vendor is empty...")
	}
	if name == "" {
		name = "na"
		fmt.Println("WARNING: name is empty...")
	}
	return strings.ToLower(strings.Join([]string{profile, vendor, name}, IdSeparator))
}

func PackageExtension() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	executablePath := path.Join(BuildPath, ExecutableName)

	cmd := exec.Command("go", "build", "-o", executablePath, "main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to build utility script: %v", err)
	}

	outputFolder := path.Join(cwd, ZippedFolderName)
	executableFullPath := filepath.Join(cwd, executablePath)
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

	metadata.BuildTime = getBuildTime()
	commentInfo, _ := json.Marshal(metadata)
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	err = addExecutableToZip(zipWriter, executableFullPath)
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

	id := GetId(metadata.Profile, metadata.Vendor, metadata.Name)
	outputZipPath := path.Join(outputFolder, strings.Join([]string{id, IdSeparator, metadata.Version, ".zip"}, ""))
	err = os.WriteFile(outputZipPath, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	err = os.RemoveAll(BuildPath)
	if err != nil {
		fmt.Println("Warning: failed to remove temporary executable:", err)
	}

	fmt.Printf("ZIP file created at: %s\n", outputZipPath)
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

func getBuildTime() string {
	currentTime := time.Now().UTC()
	return currentTime.Format(time.RFC3339)
}
