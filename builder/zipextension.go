package builder

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/PJNube/lib-extensions/manifest"
)

const (
	ExecutableName   = "extension"
	ZippedFolderName = "out"
	ConfigFileName   = "config.yaml"
)

// Opts controls optional packaging behaviour.
type Opts struct {
	// Architecture overrides the target architecture in the zip filename
	// (e.g. "arm64", "amd64"). When empty, unipack defaults to "arm64"
	// for BE and CE profiles.
	Architecture string

	// ConfigDir is the directory that contains config.yaml (the extension's
	// default runtime config). When set, config.yaml is included in the zip.
	ConfigDir string
}

// PackageExtension compiles the extension binary and packages it using unipack.
//
// It expects `unipack/unipack.sh` to be present relative to the caller's working
// directory (add unipack as a git submodule: git submodule add <url> unipack).
//
// The output zip is written to `out/`.
func PackageExtension(opts Opts) error {
	if err := os.MkdirAll(ZippedFolderName, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 1. Compile the Go binary into out/
	executablePath := filepath.Join(ZippedFolderName, ExecutableName)
	compile := exec.Command("go", "build", "-trimpath", "-ldflags", "-s -w", "-o", executablePath, ".")
	compile.Stdout = os.Stdout
	compile.Stderr = os.Stderr
	if err := compile.Run(); err != nil {
		return fmt.Errorf("failed to compile binary: %w", err)
	}

	// 2. Copy openAPI schema files into out/ so unipack's BE convention picks them up.
	//    extractBE on the server side expects schema files at schema.Path within the zip.
	metadata, err := manifest.GetMetadata()
	if err != nil {
		return fmt.Errorf("failed to read extension.json: %w", err)
	}
	for _, schema := range metadata.OpenAPISchemas {
		dst := filepath.Join(ZippedFolderName, schema.Path)
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return fmt.Errorf("failed to create directory for schema %s: %w", schema.Path, err)
		}
		if err := copyFile(schema.Path, dst); err != nil {
			return fmt.Errorf("failed to copy openAPI schema %s: %w", schema.Path, err)
		}
	}

	// 3. Copy config.yaml into out/ if a ConfigDir is provided.
	//    extractBE expects config.yaml at the zip root.
	if opts.ConfigDir != "" {
		configSrc := filepath.Join(opts.ConfigDir, ConfigFileName)
		if _, err := os.Stat(configSrc); err == nil {
			configDst := filepath.Join(ZippedFolderName, ConfigFileName)
			if err := copyFile(configSrc, configDst); err != nil {
				return fmt.Errorf("failed to copy config file: %w", err)
			}
		}
	}

	// 4. Pack with unipack (unipack must be a git submodule at ./unipack/).
	if _, err := os.Stat("unipack/unipack.sh"); err != nil {
		return fmt.Errorf(
			"unipack not found at unipack/unipack.sh — add it as a submodule:\n"+ 
				"  git submodule add git@github.com:PJNube/unipack.git unipack",
		)
	}
	packArgs := []string{"unipack/unipack.sh", ".", "--out", ZippedFolderName}
	if opts.Architecture != "" {
		packArgs = append(packArgs, "--arch", opts.Architecture)
	}
	pack := exec.Command("bash", packArgs...)
	pack.Stdout = os.Stdout
	pack.Stderr = os.Stderr
	if err := pack.Run(); err != nil {
		return fmt.Errorf("failed to pack with unipack: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
