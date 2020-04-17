/*
	Copyright (c) 2020 Michael Saigachenko
*/

package dpk

import (
	"os"

	gc "github.com/untillpro/gochips"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	gc.ExitIfError(err)
	return true
}
