package server

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
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

func Start(cl client.Client, port int, metricsPort int) error {
	if !isDebug() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	if isDebug() {
		r.Use(gin.Logger())
	}

	metricRouter := gin.Default()
	metricRouter.GET("/metrics", prometheusHandler())

	r.SetHTMLTemplate(template.Must(template.New("index").Funcs(map[string]any{
		"mod":        func(a, b int) int { return a % b },
		"lower":      func(val interface{}) string { return strings.ToLower(fmt.Sprintf("%v", val)) },
		"healthIcon": healthIcon,
		"syncIcon":   syncIcon,
	}).Parse(pageTemplate)))
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{
			"apps":        cl.Applications().WithUpdates(c.Query("project")),
			"url":         encodeBaseURL(cl),
			"updates":     true,
			"titleSuffix": "Updates",
		})
	})
	r.GET("/all", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{
			"apps":        cl.Applications().ForProject(c.Query("project")),
			"url":         encodeBaseURL(cl),
			"titleSuffix": "All",
		})
	})
	r.GET("/helm", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{
			"apps":        cl.Applications().WithRepoType(types.RepoTypeHelm, c.Query("project")),
			"url":         encodeBaseURL(cl),
			"titleSuffix": types.RepoTypeHelm,
		})
	})
	r.GET("/git", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{
			"apps":        cl.Applications().WithRepoType(types.RepoTypeGit, c.Query("project")),
			"url":         encodeBaseURL(cl),
			"titleSuffix": types.RepoTypeGit,
		})
	})
	r.GET("/health", func(c *gin.Context) {
		if cl.Ready() {
			c.JSON(http.StatusOK, map[string]string{"status": "OK"})
		} else {
			c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "AppsNotUpdated"})
		}
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

	go func() {
		log.Printf("Starting metrics on port %d", metricsPort)
		log.Fatal(metricRouter.Run(fmt.Sprintf(":%d", metricsPort)))
	}()

	log.Printf("Starting server on port %d", port)
	return r.Run(fmt.Sprintf(":%d", port))
}

func encodeBaseURL(cl client.Client) string {
	return base64.StdEncoding.EncodeToString([]byte(cl.URL()))
}

func isDebug() bool {
	return os.Getenv("DEBUG") == "true"
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
