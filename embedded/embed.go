package embedded

import (
	"embed"
	"io/fs"
)

//go:embed templates static deploy vendor
var files embed.FS

var (
	Templates = mustSub("templates")
	Static    = mustSub("static")
	Deploy    = mustSub("deploy")
	Vendor    = mustSub("vendor")
)

func mustSub(dir string) fs.FS {
	sub, err := fs.Sub(files, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
