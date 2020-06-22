package settings

import (
	"github.com/flosch/pongo2"
	"path"
)

var (
	CacheLocation string
	WorkDir string
)

func templatePath(name string) string {
	return path.Join(WorkDir, "templates", name)
}

var AlbumTemplate *pongo2.Template
var PictureTemplate *pongo2.Template

func InitSettings() {
	AlbumTemplate = pongo2.Must(pongo2.FromFile(templatePath("albums.html")))
	PictureTemplate = pongo2.Must(pongo2.FromFile(templatePath("pictures.html")))
}