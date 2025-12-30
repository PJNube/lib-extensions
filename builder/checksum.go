package builder

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func generateChecksums(outputFile *os.File, filePaths ...string) error {
	for _, filePath := range filePaths {
		err := generateChecksum(outputFile, filePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateChecksum(outputFile *os.File, filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if info.IsDir() || info.Name() == outputFile.Name() {
		return nil
	}

	// Calculate the SHA-256 hash
	hash, err := hashFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to hash %s: %w", filePath, err)
	}

	fileName := filepath.Base(filePath)
	_, err = fmt.Fprintf(outputFile, "%s  %s\n", hash, fileName)

	if err != nil {
		return err
	}

	return nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
