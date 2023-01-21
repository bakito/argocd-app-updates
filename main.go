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
	"github.com/liggitt/tabwriter"
	"golang.org/x/mod/semver"
)

const (
	urlApplications = "/api/v1/applications"
	urlHelmCharts   = "/api/v1/repositories/%s/helmcharts"
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
	w := tabwriter.NewWriter(os.Stdout, 6, 4, 3, ' ', tabwriter.RememberWidths)
	_, _ = fmt.Fprintln(w, strings.Join([]string{"PROJECT", "NAME", "SOURCE TYPE", "CHART", "TARGET REVISION", "NEWEST VERSION"}, "\t"))

	charts := make(map[string]*types.HelmCharts)

	for _, app := range helmApps {
		hc, err := getHelmCharts(client, app, charts)
		if err != nil {
			log.Fatal(err)
		}
		updateAvailable := semver.Compare("v"+app.Spec.Source.TargetRevision, "v"+hc.Versions[0]) < 0

		var version string
		if updateAvailable {
			version = color.New(color.FgYellow).Sprint(hc.Versions[0])
		} else {
			version = color.New(color.FgGreen).Sprint(hc.Versions[0])
		}

		_, _ = fmt.Fprintln(w, strings.Join([]string{
			app.Spec.Project,
			app.Metadata.Name,
			app.Status.SourceType,
			app.Spec.Source.Chart,
			app.Spec.Source.TargetRevision,
			version,
		}, "\t"))
	}
	_ = w.Flush()
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
