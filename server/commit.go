package server

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"

	"github.com/frc-2175/benkins/ansicolors"

	"github.com/frc-2175/benkins/shared"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
)

func CommitIndex(r *gin.Engine, loader Loader) gin.HandlerFunc {
	r.HTMLRender.(multitemplate.Renderer).AddFromFilesFuncs("commit", TemplateFuncs, "server/tmpl/base.html", "server/tmpl/commit.html")

	return func(c *gin.Context) {
		projectName := shared.NewProjectNameFromEncoded(c.Param("project"))

		commit, err := loader.Commit(projectName, c.Param("hash"))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		logs, err := ioutil.ReadFile(filepath.Join(commit.Filepath, shared.ExecutionLogFilename))
		if err != nil {
			code := http.StatusInternalServerError
			if os.IsNotExist(err) {
				code = http.StatusNotFound
			}

			c.AbortWithError(code, err)
			return
		}

		blocks := ansicolors.Process(logs)

		var htmlBlocks []HTMLBlock
		for _, b := range blocks {
			var classes []string
			for _, a := range b.Attributes {
				if class, ok := Attribute2Class[a]; ok {
					classes = append(classes, class)
				}
			}

			htmlBlocks = append(htmlBlocks, HTMLBlock{
				Classes: strings.Join(classes, " "),
				Text:    string(b.Contents),
			})
		}

		c.HTML(http.StatusOK, "commit", v{
			"projectName": projectName,
			"commit":      commit,
			"logBlocks":   htmlBlocks,
		})
	}
}

type HTMLBlock struct {
	Classes string
	Text    string
}

var Attribute2Class = map[int]string{
	// Base attributes
	int(color.Bold):      "bold",
	int(color.Faint):     "faint",
	int(color.Italic):    "italic",
	int(color.Underline): "underline",

	// Foreground text colors
	int(color.FgBlack):   "fgblack",
	int(color.FgRed):     "fgred",
	int(color.FgGreen):   "fggreen",
	int(color.FgYellow):  "fgyellow",
	int(color.FgBlue):    "fgblue",
	int(color.FgMagenta): "fgmagenta",
	int(color.FgCyan):    "fgcyan",
	int(color.FgWhite):   "fgwhite",

	// Foreground Hi-Intensity text colors
	int(color.FgHiBlack):   "fghiblack",
	int(color.FgHiRed):     "fghired",
	int(color.FgHiGreen):   "fghigreen",
	int(color.FgHiYellow):  "fghiyellow",
	int(color.FgHiBlue):    "fghiblue",
	int(color.FgHiMagenta): "fghimagenta",
	int(color.FgHiCyan):    "fghicyan",
	int(color.FgHiWhite):   "fghiwhite",

	// Background text colors
	int(color.BgBlack):   "bgblack",
	int(color.BgRed):     "bgred",
	int(color.BgGreen):   "bggreen",
	int(color.BgYellow):  "bgyellow",
	int(color.BgBlue):    "bgblue",
	int(color.BgMagenta): "bgmagenta",
	int(color.BgCyan):    "bgcyan",
	int(color.BgWhite):   "bgwhite",

	// Background Hi-Intensity text colors
	int(color.BgHiBlack):   "bghiblack",
	int(color.BgHiRed):     "bghired",
	int(color.BgHiGreen):   "bghigreen",
	int(color.BgHiYellow):  "bghiyellow",
	int(color.BgHiBlue):    "bghiblue",
	int(color.BgHiMagenta): "bghimagenta",
	int(color.BgHiCyan):    "bghicyan",
	int(color.BgHiWhite):   "bghiwhite",
}
