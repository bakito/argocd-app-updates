package server

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/bakito/argocd-app-updates/pkg/client"
	"github.com/bakito/argocd-app-updates/pkg/types"
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
)

func Start(cl client.Client, port int) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.SetHTMLTemplate(template.Must(template.New("index").Funcs(map[string]any{
		"mod":        func(a, b int) int { return a % b },
		"lower":      func(val interface{}) string { return strings.ToLower(fmt.Sprintf("%v", val)) },
		"healthIcon": healthIcon,
		"syncIcon":   syncIcon,
	}).Parse(pageTemplate)))
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{
			"apps":        cl.Applications().WithUpdates(c.Query("project")),
			"updates":     true,
			"titleSuffix": "Updates",
		})
	})
	r.GET("/all", func(c *gin.Context) {
		apps := cl.Applications().ForProject(c.Query("project"))
		c.HTML(http.StatusOK, "index", gin.H{
			"apps":        apps,
			"titleSuffix": "All",
		})
	})
	r.GET("/helm", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{
			"apps":        cl.Applications().WithRepoType(types.RepoTypeHelm, c.Query("project")),
			"titleSuffix": types.RepoTypeHelm,
		})
	})
	r.GET("/git", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{
			"apps":        cl.Applications().WithRepoType(types.RepoTypeGit, c.Query("project")),
			"titleSuffix": types.RepoTypeGit,
		})
	})
	r.GET("/open/:app", func(c *gin.Context) {
		app := c.Param("app")
		redirect := fmt.Sprintf("%s/applications/argocd/%s?view=tree&resource=", cl.URL(), app)
		c.Redirect(http.StatusTemporaryRedirect, redirect)
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
