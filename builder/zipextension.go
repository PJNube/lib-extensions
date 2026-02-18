package builder

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/PJNube/lib-extensions/manifest"
	"github.com/PJNube/lib-extensions/naming"
)

const (
	ExecutableName   = "extension"
	ZippedFolderName = "out"
	BuildPath        = "executable"
	ConfigFileName   = "config.yaml"
)

// PackageExtension packages the extension into a zip file.
// The 'architecture' field is optional and mainly used for testing.
func PackageExtension(opts Opts) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	metadata, err := manifest.GetMetadata()
	if err != nil {
		return err
	}

	version := metadata.Version
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	ldFlags := fmt.Sprintf("-s -w -X main.Version=%s", version)
	executablePath := path.Join(BuildPath, ExecutableName)
	cmd := exec.Command(
		"go", "build",
		"-trimpath",
		"-ldflags", ldFlags,
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

	metadata.BuildTime = getBuildTime()
	if opts.Architecture != "" {
		metadata.Dependencies.Architecture = opts.Architecture
	} else {
		metadata.Dependencies.Architecture = runtime.GOARCH
	}

	executableFullPath := filepath.Join(cwd, executablePath)
	metadataFilePath := path.Join(cwd, manifest.MetadataFileName)
	filePaths := []string{executableFullPath, metadataFilePath}
	for _, schema := range metadata.OpenAPISchemas {
		filePaths = append(filePaths, schema.Path)
	}

	for _, filePath := range filePaths {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("required file %s does not exist", filePath)
		}
	}

	configFilePath, err := getConfigPath(opts)
	if err != nil {
		return fmt.Errorf("failed to get config file path: %w", err)
	} else if configFilePath != "" {
		filePaths = append(filePaths, configFilePath)
	}

	commentInfo, _ := json.Marshal(metadata)
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

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
