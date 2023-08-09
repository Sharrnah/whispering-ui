package Updater

import (
	"fmt"
	"os"
	"strings"
	"whispering-tiger-ui/Utilities"
)

func CheckFileHash(filepath string, expectedHash string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	calculatedHashStr, err := Utilities.FileHash(file)
	if err != nil {
		return fmt.Errorf("failed to compute hash: %w", err)
	}

	if strings.ToLower(calculatedHashStr) != strings.ToLower(expectedHash) {
		return fmt.Errorf("file hash does not match expected hash (calculated: %s, expected: %s)", calculatedHashStr, expectedHash)
	}

	return nil
}
