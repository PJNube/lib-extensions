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

	if opts.ConfigDir != "" {
		configFilePath := path.Join(opts.ConfigDir, ConfigFileName)
		if _, err := os.Stat(configFilePath); err != nil {
			return fmt.Errorf("config file does not exist at specified path: %s", configFilePath)
		}
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

	updateReleaseVersion(metadata)

	parts := []string{id, metadata.Version}
	if metadata.DistVersion != "" {
		parts = append(parts, metadata.DistVersion)
	}
	parts = append(parts, metadata.Dependencies.Architecture)
	zipFileName := strings.Join(parts, naming.IdSeparator) + ".zip"
	outputZipPath := filepath.Join(outputFolder, zipFileName)
	err = os.WriteFile(outputZipPath, buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write zip file %s: %w", outputZipPath, err)
	}

	err = os.RemoveAll(BuildPath)
	if err != nil {
		fmt.Println("Warning: failed to remove temporary executable:", err)
	}

	fmt.Printf("ZIP file created at: %s\n", outputZipPath)
	return nil
}

func updateReleaseVersion(metadata *manifest.Metadata) {
	if metadata.DistVersion != "" {
		return
	}

	releaseVersion := getReleaseVersionFormGit()
	if releaseVersion == "" {
		return
	}

	metadata.DistVersion = releaseVersion
	metadata.CommitID = getCommitID()
}

func getReleaseVersionFormGit() string {
	cmd := exec.Command("git", "branch", "--show-current")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	branchName := strings.TrimSpace(string(out))
	if strings.HasPrefix(branchName, "release/") {
		return strings.TrimPrefix(branchName, "release/")
	}

	return ""
}

func getCommitID() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
