package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/bakito/argocd-app-updates/pkg/types"
	"github.com/go-resty/resty/v2"
	"golang.org/x/mod/semver"
)

const (
	urlApplications = "http://localhost:8080/api/v1/applications"
	urlHelmCharts   = "https://argocd.k3s.bakito.net/api/v1/repositories/%s/helmcharts"
)

func main() {
	client := resty.New()

	apps, err := readApplications(client)
	if err != nil {
		log.Fatal(err)
	}

	helmApps := apps.Helm()

	charts := make(map[string]*types.HelmCharts)

	for _, app := range helmApps {
		hc, err := getHelmCharts(client, app, charts)
		if err != nil {
			log.Fatal(err)
		}
		if semver.Compare("v"+app.Spec.Source.TargetRevision, "v"+hc.Versions[0]) < 0 {
			log.Printf("Update available for %q: %s -> %s", app.Metadata.Name, app.Spec.Source.TargetRevision, hc.Versions[0])
		}

	}
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
