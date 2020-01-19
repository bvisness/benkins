package server

import (
	"fmt"
	"html/template"
	"net/url"

	"github.com/frc-2175/benkins/shared"
)

var TemplateFuncs = template.FuncMap{
	"projectUrl": ProjectUrl,
	"commitUrl":  CommitUrl,
	"fileUrl":    FileUrl,
	"short":      Short,
}

func ProjectUrl(name string) string {
	return fmt.Sprintf("/p/%s", shared.Base64Encode(name))
}

func CommitUrl(projectName, hash string) string {
	return fmt.Sprintf("/p/%s/%s", shared.Base64Encode(projectName), hash)
}

func FileUrl(projectName, hash, filename string) string {
	return fmt.Sprintf("/p/%s/%s/f/%s", shared.Base64Encode(projectName), hash, url.PathEscape(filename))
}

func Short(hash string) string {
	return hash[0:7]
}
