package Updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

func CheckFileHash(filepath string, expectedHash string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("failed to compute hash: %w", err)
	}

	calculatedHash := hasher.Sum(nil)
	calculatedHashStr := hex.EncodeToString(calculatedHash)

	if strings.ToLower(calculatedHashStr) != strings.ToLower(expectedHash) {
		return fmt.Errorf("file hash does not match expected hash (calculated: %s, expected: %s)", calculatedHashStr, expectedHash)
	}

	return nil
}
