package builder

import (
	"fmt"
	"os"
	"os/exec"
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
}

// PackageExtension compiles the extension binary and packages it using unipack.
//
// It expects `unipack/unipack.sh` to be present relative to the caller's working
// directory (add unipack as a git submodule: git submodule add <url> unipack).
//
// The binary is compiled to `out/extension`. pack.toml (or unipack's built-in
// convention) controls which files are zipped; add `"out"` to the strip list so
// the binary appears as `extension` at the zip root. The output zip is written
// to `out/`.
func PackageExtension(opts Opts) error {
	// 1. Compile the Go binary into out/ so it doesn't conflict with the
	//    extension/ package subdirectory that most extensions contain.
	if err := os.MkdirAll(ZippedFolderName, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	executablePath := ZippedFolderName + "/" + ExecutableName
	compile := exec.Command("go", "build", "-trimpath", "-ldflags", "-s -w", "-o", executablePath, ".")
	compile.Stdout = os.Stdout
	compile.Stderr = os.Stderr
	if err := compile.Run(); err != nil {
		return fmt.Errorf("failed to compile binary: %w", err)
	}

	// 2. Pack with unipack (unipack must be a git submodule at ./unipack/).
	//    pack.toml in the extension directory is the single source of truth for
	//    which files are included in the zip.
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
