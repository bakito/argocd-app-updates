package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/bakito/argocd-app-updates/pkg/types"
	"github.com/fatih/color"
	"github.com/go-resty/resty/v2"
	"github.com/juju/ansiterm/tabwriter"
	"golang.org/x/mod/semver"
)

const (
	urlApplications = "/api/v1/applications"
	urlHelmCharts   = "/api/v1/repositories/%s/helmcharts"
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
	var server string

	flag.StringVar(&server, "server", "http://localhost:8080", "Define the argo-cd target server")
	flag.Parse()

	client := resty.New().SetBaseURL(server)

	apps, err := readApplications(client)
	if err != nil {
		log.Fatal(err)
	}

	helmApps := apps.Helm()
	w := tabwriter.NewWriter(os.Stdout, 6, 4, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, strings.Join([]string{
		"PROJECT",
		"NAME",
		"HEALTH STATUS",
		"SYNC STATUS",
		"AUTO SYNC",
		"SOURCE TYPE",
		"CHART",
		"TARGET REVISION",
		"NEWEST VERSION",
	}, "\t"))

	charts := make(map[string]*types.HelmCharts)

	for _, app := range helmApps {
		hc, err := getHelmCharts(client, app, charts)
		if err != nil {
			log.Fatal(err)
		}
		updateAvailable := semver.Compare("v"+app.Spec.Source.TargetRevision, "v"+hc.Versions[0]) < 0

		var version string
		if updateAvailable {
			version = colorYellow.Sprint(hc.Versions[0])
		} else {
			version = colorGreen.Sprint(hc.Versions[0])
		}

		_, _ = fmt.Fprintln(w, strings.Join([]string{
			app.Spec.Project,
			app.Metadata.Name,
			healthStatus(app),
			syncStatus(app),
			autoSync(app),
			app.Status.SourceType,
			app.Spec.Source.Chart,
			app.Spec.Source.TargetRevision,
			version,
		}, "\t"))
	}
	_ = w.Flush()
}

func autoSync(app types.Application) string {
	if app.Spec.SyncPolicy.Automated == nil {
		return ""
	}
	return colorHiCyan.Sprintf("%v", true)
}

func syncStatus(app types.Application) string {
	syncStatus := app.Status.Sync.Status
	switch syncStatus {
	case "Synced":
		syncStatus = colorGreen.Sprint(syncStatus)
	case "OutOfSync":
		syncStatus = colorYellow.Sprint(syncStatus)
	}
	return syncStatus
}

func healthStatus(app types.Application) string {
	health := app.Status.Health.Status
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

func getHelmCharts(client *resty.Client, app types.Application, charts map[string]*types.HelmCharts) (*types.HelmChart, error) {
	if hc, ok := charts[app.Spec.Source.RepoURL]; ok {
		return hc.Chart(app.Spec.Source.Chart), nil
	}
	hc := &types.HelmCharts{}
	_, err := client.R().SetResult(hc).Get(fmt.Sprintf(urlHelmCharts, url.QueryEscape(app.Spec.Source.RepoURL)))
	if err != nil {
		return nil, err
	}
	charts[app.Spec.Source.RepoURL] = hc
	return hc.Chart(app.Spec.Source.Chart), err
}

func readApplications(client *resty.Client) (*types.ApplicationList, error) {
	apps := &types.ApplicationList{}
	_, err := client.R().SetResult(apps).Get(urlApplications)
	return apps, err
}
