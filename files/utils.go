package files

import (
	"errors"
	"os"
	"path"
)

func CreateFolder(base string, directory string) (folder string, err error) {
	fullFolder := path.Join(base, directory)
	if _, err := os.Stat(fullFolder); os.IsNotExist(err) {
		err = os.Mkdir(fullFolder, 0700)
		return fullFolder, err
	}
	return fullFolder, nil
}

func OpenFile(base string, filename string) (file *os.File, err error) {
	fullFile := path.Join(base, filename)
	if _, err := os.Stat(fullFile); os.IsNotExist(err) {
		file, err = os.Create(fullFile)
		return file, err
	} else {
		file, err = os.OpenFile(fullFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0700)
		return file, err
	}
}

func DeleteFile(filepath string) (err error) {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return errors.New("File doesn't exist")
	} else {
		err = os.Remove(filepath)
		return err
	}
}

func FindExeDir() (exePath string) {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return path.Dir(ex)
}
