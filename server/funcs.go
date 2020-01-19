package server

import (
	"fmt"
	"html/template"

	"github.com/frc-2175/benkins/shared"
)

var TemplateFuncs = template.FuncMap{
	"projectUrl": ProjectUrl,
	"commitUrl":  CommitUrl,
	"short":      Short,
}

func ProjectUrl(name string) string {
	return fmt.Sprintf("/p/%s", shared.Base64Encode(name))
}

func CommitUrl(projectName, hash string) string {
	return fmt.Sprintf("/p/%s/%s", shared.Base64Encode(projectName), hash)
}

func Short(hash string) string {
	return hash[0:7]
}
