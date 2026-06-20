package web

import (
	"io/fs"

	"embed"
)

//go:embed all:dist
var dist embed.FS

// Dist, build edilmiş React SPA'sını dist/ kökünden sunulabilir fs.FS olarak döner.
func Dist() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}
