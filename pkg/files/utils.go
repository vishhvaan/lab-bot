package files

import (
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
