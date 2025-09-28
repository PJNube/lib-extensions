package builder

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PJNube/lib-extensions/manifest"
	"github.com/PJNube/lib-extensions/naming"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	ExecutableName   = "extension"
	ZippedFolderName = "out"
	MetadataFileName = "extension.json"
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

	metadata := manifest.Metadata{}
	err = json.Unmarshal(tempBytes.Bytes(), &metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	metadata.BuildTime = getBuildTime()

	if len(arch) > 0 {
		metadata.Dependencies.Architecture = arch[0]
	} else {
		metadata.Dependencies.Architecture = runtime.GOARCH
	}

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

	id := naming.GetId(metadata.Profile, metadata.Vendor, metadata.Name)
	zipFileName := strings.Join([]string{id, metadata.Version, metadata.Dependencies.Architecture}, naming.IdSeparator)
	outputZipPath := path.Join(outputFolder, strings.Join([]string{zipFileName, ".zip"}, ""))
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
