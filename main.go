package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bakito/argocd-app-updates/pkg/client"
	ss "github.com/bakito/argocd-app-updates/pkg/server"
	"github.com/bakito/argocd-app-updates/pkg/types"
	"github.com/bakito/argocd-app-updates/version"
	"github.com/fatih/color"
	"github.com/juju/ansiterm/tabwriter"
	"github.com/robfig/cron/v3"
)

var (
	colorGreen   = color.New(color.FgGreen)
	colorYellow  = color.New(color.FgYellow)
	colorBlue    = color.New(color.FgCyan)
	colorRed     = color.New(color.FgRed)
	colorMagenta = color.New(color.FgHiMagenta)
	colorHiCyan  = color.New(color.FgHiCyan)
)

func main() {
	var (
		argoURL        string
		serverMode     bool
		ver            bool
		port           int
		cronExpression string
		project        string
	)

	flag.StringVar(&argoURL, "argo-server", "http://localhost:8080", "Define the argo-cd target server URL")
	flag.BoolVar(&serverMode, "server", false, "run as server")
	flag.BoolVar(&ver, "version", false, "Print the version")
	flag.IntVar(&port, "port", 8080, "Server port")
	flag.StringVar(&cronExpression, "cron", "*/15 * * * *", "The cron expression to sync the apps in server mode")
	flag.StringVar(&project, "project", "", "Optional define the project to search applications in CLI mode")
	flag.Parse()

	if ver {
		fmt.Printf("argocd-app-updates %s\n", version.Version)
		return
	}

	cl := client.NewClient(argoURL)
	if err := cl.Update(); err != nil {
		log.Fatal(err)
	}

	if serverMode {

		log.Printf("Starting argocd-app-updates %q", version.Version)
		log.Printf("Using cron expression %q to update applications", cronExpression)
		c := cron.New()
		_, err := c.AddFunc(cronExpression, func() {
			log.Printf("Updating applications")
			if err := cl.Update(); err != nil {
				log.Printf("Error during application update: %v", err)
			}
		})
		c.Start()
		if err != nil {
			log.Fatal(err)
		}

		log.Fatal(ss.Start(cl, port))
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 6, 4, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, strings.Join([]string{
		"PROJECT",
		"NAME",
		"HEALTH STATUS",
		"SYNC STATUS",
		"AUTO SYNC",
		"CHART",
		"VERSION",
		"NEWEST VERSION",
	}, "\t"))

	apps := cl.Applications().WithRepoType(types.RepoTypeHelm, project)
	for _, app := range apps {
		var version string
		if app.NewestVersion != "" {
			version = colorYellow.Sprint(app.NewestVersion)
		} else {
			version = colorGreen.Sprint(app.Version)
		}

		_, _ = fmt.Fprintln(w, strings.Join([]string{
			app.Project,
			app.Name,
			healthStatus(app),
			syncStatus(app),
			autoSync(app),
			app.Chart,
			app.Revision,
			version,
		}, "\t"))
	}
	_ = w.Flush()
}

func autoSync(app types.Application) string {
	if !app.Automated {
		return ""
	}
	return colorHiCyan.Sprintf("%v", true)
}

func syncStatus(app types.Application) string {
	syncStatus := app.SyncStatus
	switch syncStatus {
	case "Synced":
		syncStatus = colorGreen.Sprint(syncStatus)
	case "OutOfSync":
		syncStatus = colorYellow.Sprint(syncStatus)
	}
	return syncStatus
}

func healthStatus(app types.Application) string {
	health := app.HealthStatus
	switch health {
	case "Healthy":
		health = colorGreen.Sprint(health)
	case "Progressing":
		health = colorBlue.Sprint(health)
	case "Degraded":
		health = colorRed.Sprint(health)
	case "Missing":
		health = colorYellow.Sprint(health)
	case "Suspended":
		health = colorMagenta.Sprint(health)
	}
	return health
}
