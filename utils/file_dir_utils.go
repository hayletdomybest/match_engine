package utils

import "os"

func FileExisted(name string) bool {
	_, err := os.Stat(name)

	return err == nil
}

func MkdirAll(path string) error {
	if FileExisted(path) {
		return nil
	}

	return os.MkdirAll(path, os.ModePerm)
}
