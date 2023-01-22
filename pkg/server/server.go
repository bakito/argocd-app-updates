package server

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/bakito/argocd-app-updates/pkg/client"
	"github.com/gin-gonic/gin"
)

var (
	//go:embed template.tpl.html
	pageTemplate string

	//go:embed helm.png
	iconHelm []byte

	//go:embed git.png
	iconGit []byte

	//go:embed favicon.png
	favicon []byte
	//go:embed styles.css
	styles string
)

func Start(cl client.Client, port int) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.SetHTMLTemplate(template.Must(template.New("index").Funcs(map[string]any{
		"mod":        func(a, b int) int { return a % b },
		"lower":      func(val string) string { return strings.ToLower(val) },
		"healthIcon": healthIcon,
		"syncIcon":   syncIcon,
	}).Parse(pageTemplate)))
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{
			"apps":    cl.Applications().WithUpdates(c.Query("project")),
			"updates": true,
		})
	})
	r.GET("/all", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{
			"apps": cl.Applications().ForProject(c.Query("project")),
		})
	})
	r.GET("/helm", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{
			"apps": cl.Applications().WithSourceType("Helm", c.Query("project")),
		})
	})
	r.GET("/git", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{
			"apps": cl.Applications().WithSourceType("Git", c.Query("project")),
		})
	})

	r.GET("/helm.png", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/png", iconHelm)
	})
	r.GET("/git.png", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/png", iconGit)
	})
	r.GET("/favicon.png", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/png", favicon)
	})
	r.GET("/styles.css", func(c *gin.Context) {
		c.String(http.StatusOK, styles)
	})
	log.Printf("Starting server on port %d", port)
	return r.Run(fmt.Sprintf(":%d", port))
}

func healthIcon(status string) string {
	var icon string
	switch status {
	case "Healthy":
		icon = "fa-heart"
	case "Progressing":
		icon = "fa-circle-notch"
	case "Degraded":
		icon = "fa-heart-broken"
	case "Missing":
		icon = "fa-ghost"
	case "Suspended":
		icon = "fa-pause-circle"
	case "Unknown":
		icon = "fa-question-circle"
	}
	return icon
}

func syncIcon(status string) string {
	var icon string
	switch status {
	case "Synced":
		icon = "fa-check-circle"
	case "OutOfSync":
		icon = "fa-arrow-alt-circle-up"
	}
	return icon
}
