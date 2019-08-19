// +build !darwin

package clonefile

import (
	"errors"
	"os"
)

func CloneFile(src *os.File, dst *os.File) (success bool, err error) {
	return false, errors.New("not supported")
}

func CloneFileByPath(src string, dst string) (success bool, err error) {
	return false, errors.New("not supported")
}
