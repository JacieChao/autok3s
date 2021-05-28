package utils

import (
	"fmt"
	"os"
)

const (
	HomeEnv        = "HOME"
	HomeDriveEnv   = "HOMEDRIVE"
	HomePathEnv    = "HOMEPATH"
	UserProfileEnv = "USERPROFILE"
)

// EnsureFolderExist will create folder if not exist
func EnsureFolderExist(path string) error {
	if path == "" {
		return fmt.Errorf("path %s cannot be empty", path)
	}
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

// EnsureFileExist will create new file if not exist
func EnsureFileExist(path, file string) error {
	if err := EnsureFolderExist(path); err != nil {
		return err
	}
	n := fmt.Sprintf("%s/%s", path, file)
	if _, err := os.Stat(n); os.IsNotExist(err) {
		_, fileE := os.Create(n)
		if fileE != nil {
			return fileE
		}
	}

	return nil
}

// UserHome return user home path
func UserHome() string {
	if home := os.Getenv(HomeEnv); home != "" {
		return home
	}
	homeDrive := os.Getenv(HomeDriveEnv)
	homePath := os.Getenv(HomePathEnv)
	if homeDrive != "" && homePath != "" {
		return homeDrive + homePath
	}
	return os.Getenv(UserProfileEnv)
}
