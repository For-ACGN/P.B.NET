package config

import (
	"io/ioutil"
	"os"

	"project/internal/logger"
	"project/internal/system"

	"project/script/internal/log"
)

// CreateBackup is used to create backup file.
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

// RestoreBackup is used to restore backup file.
func RestoreBackup(path string) error {
	return os.Rename(path+".bak", path)
}

// RemoveBackup is used to remove backup file.
func RemoveBackup(path string) error {
	return os.Remove(path + ".bak")
}

// CreateGoModBackup is used to create go.mod backup, go command will change it(before go1.16).
func CreateGoModBackup() bool {
	log.Println(logger.Info, "create backup of go.mod")
	_, err := CreateBackup("go.mod")
	if err != nil {
		log.Println(logger.Error, "failed to create backup of go.mod:", err)
		return false
	}
	return true
}

// RestoreGoModBackup is used to restore go.mod backup.
func RestoreGoModBackup() bool {
	log.Println(logger.Info, "restore backup of go.mod")
	err := RestoreBackup("go.mod")
	if err != nil {
		log.Println(logger.Error, "failed to restore backup of go.mod:", err)
		return false
	}
	return true
}

// RemoveGoModBackup is used to remove go.mod backup.
func RemoveGoModBackup() bool {
	log.Println(logger.Info, "remove backup of go.mod")
	err := RemoveBackup("go.mod")
	if err != nil {
		log.Println(logger.Error, "failed to remove backup of go.mod:", err)
		return false
	}
	return true
}

// CreateGoSumBackup is used to create go.sum backup.
func CreateGoSumBackup() bool {
	log.Println(logger.Info, "create backup of go.sum")
	_, err := CreateBackup("go.sum")
	if err != nil {
		log.Println(logger.Error, "failed to create backup of go.sum:", err)
		return false
	}
	return true
}

// RestoreGoSumBackup is used to restore go.sum backup.
func RestoreGoSumBackup() bool {
	log.Println(logger.Info, "restore backup of go.sum")
	err := RestoreBackup("go.sum")
	if err != nil {
		log.Println(logger.Error, "failed to restore backup of go.sum:", err)
		return false
	}
	return true
}

// RemoveGoSumBackup is used to remove go.sum backup.
func RemoveGoSumBackup() bool {
	log.Println(logger.Info, "remove backup of go.sum")
	err := RemoveBackup("go.sum")
	if err != nil {
		log.Println(logger.Error, "failed to remove backup of go.sum:", err)
		return false
	}
	return true
}
