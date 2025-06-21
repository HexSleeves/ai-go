//go:build !js
// +build !js

package ui

import "codeberg.org/anaseto/gruid"

func GetDriver() gruid.Driver {
	return driver
}
