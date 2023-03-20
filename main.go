package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bakito/argocd-app-updates/pkg/client"
	ss "github.com/bakito/argocd-app-updates/pkg/server"
	"github.com/bakito/argocd-app-updates/pkg/terminal"
	"github.com/bakito/argocd-app-updates/pkg/types"
	"github.com/bakito/argocd-app-updates/version"
	"github.com/robfig/cron/v3"
)

const (
	envArgoUser     = "ARGOCD_USER"
	envArgoPassword = "ARGOCD_PASSWORD"
)

func main() {
	var (
		argoURL        string
		serverMode     bool
		ver            bool
		port           int
		metricsPort    int
		cronExpression string
		project        string
	)

	flag.StringVar(&argoURL, "argo-server", "http://localhost:8080", "Define the argo-cd target server URL")
	flag.BoolVar(&serverMode, "server", false, "run as server")
	flag.BoolVar(&ver, "version", false, "Print the version")
	flag.IntVar(&port, "port", 8080, "Server port")
	flag.IntVar(&metricsPort, "metrics-port", 8081, "Metrics port")
	flag.StringVar(&cronExpression, "cron", "*/15 * * * *", "The cron expression to sync the apps in server mode")
	flag.StringVar(&project, "project", "", "Optional define the project to search applications in CLI mode")
	flag.Parse()

	if ver {
		fmt.Printf("argocd-app-updates %s\n", version.Version)
		return
	}

	cl := client.NewClient(argoURL, os.Getenv(envArgoUser), os.Getenv(envArgoPassword))
	if err := cl.Update(); err != nil {
		log.Fatal(err)
	}

	if !serverMode {
		terminal.Render(cl.Applications().WithRepoType(types.RepoTypeHelm, project))
		return
	}

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
	log.Fatal(ss.Start(cl, port, metricsPort))
}
