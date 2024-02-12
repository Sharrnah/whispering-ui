package Utilities

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func MigrateProfileSettingsLocation1704429446() {
	newProfilesDir := filepath.Join(".", "Profiles")

	// Create the Profiles directory if it does not exist
	if err := os.MkdirAll(newProfilesDir, 0755); err != nil {
		println("Error creating Profiles directory:", err.Error())
		return
	}

	// Check if Profiles directory exists and is a directory
	dirInfo, err := os.Stat(newProfilesDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// If the directory does not exist, create it
			if err := os.Mkdir(newProfilesDir, 0755); err != nil {
				fmt.Println("Error creating Profiles directory:", err)
				return
			}
		} else {
			fmt.Println("Error checking Profiles directory:", err)
			return
		}
	} else if !dirInfo.IsDir() {
		fmt.Println("Profiles exists but is not a directory")
		return
	}

	// Check if Profiles directory is empty, ignoring .gitkeep
	files, err := os.ReadDir(newProfilesDir)
	if err != nil {
		fmt.Println("Error reading Profiles directory:", err)
		return
	}
	for _, file := range files {
		if file.Name() != ".gitkeep" {
			fmt.Println("Profiles directory is not empty")
			return
		}
	}

	// Move .yml and .yaml files
	files, err = os.ReadDir(".")
	if err != nil {
		fmt.Println("Error reading current directory:", err)
		return
	}
	for _, file := range files {
		if !file.IsDir() && !strings.HasPrefix(file.Name(), ".") &&
			(strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			oldPath := file.Name()
			newPath := filepath.Join(newProfilesDir, file.Name())
			if err := os.Rename(oldPath, newPath); err != nil {
				fmt.Printf("Error moving %s to Profiles: %s\n", file.Name(), err)
			} else {
				fmt.Printf("Moved %s to Profiles\n", file.Name())
			}
		}
	}
}
