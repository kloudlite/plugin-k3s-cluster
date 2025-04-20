package templates

import (
	"embed"
	"path/filepath"

	"github.com/kloudlite/operator/toolkit/templates"
)

//go:embed *
var templatesDir embed.FS

type templateFile string

const (
	ClusterLifeycleSpec templateFile = "./cluster-lifecycle.spec.yml.tpl"
)

func Read(t templateFile) ([]byte, error) {
	return templatesDir.ReadFile(filepath.Join(string(t)))
}

var ParseBytes = templates.ParseBytes
