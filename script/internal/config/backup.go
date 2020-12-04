package config

import (
	"io/ioutil"
	"os"

	"project/internal/logger"
	"project/internal/system"

	"project/script/internal/log"
)

// CreateBackup is used to create file backup.
func CreateBackup(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(path) // #nosec
	if err != nil {
		return nil, err
	}
	err = os.Rename(path, path+".bak")
	if err != nil {
		return nil, err
	}
	err = system.WriteFile(path, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// RestoreBackup is used to restore file backup.
func RestoreBackup(path string) error {
	return os.Rename(path+".bak", path)
}

// CreateGoModBackup is used to create go.mod backup, go command will change it(before go1.16).
func CreateGoModBackup() bool {
	_, err := CreateBackup("go.mod")
	if err != nil {
		log.Println(logger.Error, "failed to create backup of go.mod:", err)
		return false
	}
	return true
}

// RestoreGoModBackup is used to restore go.mod backup.
func RestoreGoModBackup() bool {
	err := RestoreBackup("go.mod")
	if err != nil {
		log.Println(logger.Error, "failed to restore backup of go.mod:", err)
		return false
	}
	return true
}
