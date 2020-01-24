package server

import (
	"fmt"
	"html/template"
	"net/url"
	"time"

	"github.com/frc-2175/benkins/shared"
)

var TemplateFuncs = template.FuncMap{
	"projectUrl": ProjectUrl,
	"commitUrl":  CommitUrl,
	"fileUrl":    FileUrl,
	"short":      Short,

	"now":        time.Now,
	"timeSecond": func() time.Duration { return time.Second },
}

func ProjectUrl(name shared.ProjectName) string {
	return fmt.Sprintf("/p/%s", name.Encoded())
}

func CommitUrl(projectName shared.ProjectName, hash string) string {
	return fmt.Sprintf("/p/%s/%s", projectName.Encoded(), hash)
}

func FileUrl(projectName shared.ProjectName, hash, filename string) string {
	return fmt.Sprintf("/p/%s/%s/f/%s", projectName.Encoded(), hash, url.PathEscape(filename))
}

func Short(hash string) string {
	return hash[0:7]
}
