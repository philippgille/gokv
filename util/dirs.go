package util

import (
	"os"
	"path/filepath"
)

func CreateAllDirs(f string, dirMode os.FileMode) error {
	_, err := os.Stat(f)
	if err != nil {
		if os.IsNotExist(err) {
			if filepath.Dir(f) != "." {
				err = os.MkdirAll(filepath.Dir(f), dirMode)
				if err != nil {
					return err
				}
			}
		} else {
			return err
		}
	}
	return nil
}
