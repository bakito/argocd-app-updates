package client

import (
	"fmt"
	"net/url"

	"github.com/bakito/argocd-app-updates/pkg/types"
	"github.com/go-resty/resty/v2"
	"golang.org/x/mod/semver"
)

const (
	urlApplications = "/api/v1/applications"
	urlHelmCharts   = "/api/v1/repositories/%s/helmcharts"
)

func NewClient(argoServer string) Client {
	return &client{
		client: resty.New().SetBaseURL(argoServer),
	}
}

type Client interface {
	Update() error
	Applications() types.Applications
}

type client struct {
	client *resty.Client
	apps   types.Applications
}

func (c *client) Applications() types.Applications {
	return c.apps
}

func (c *client) Update() error {
	apps, err := c.readApplications()
	if err != nil {
		return err
	}

	charts := make(map[string]*types.HelmChartsResponse)
	helmApps := apps.Helm()

	var myApps types.Applications

	for _, app := range helmApps {
		hc, err := c.getHelmCharts(app, charts)
		if err != nil {
			return err
		}
		myApp := types.Application{
			Name:         app.Metadata.Name,
			Project:      app.Spec.Project,
			SourceType:   app.Status.SourceType,
			RepoURL:      app.Spec.Source.RepoURL,
			Revision:     app.Spec.Source.TargetRevision,
			Chart:        app.Spec.Source.Chart,
			Version:      app.Spec.Source.TargetRevision,
			HealthStatus: app.Status.Health.Status,
			SyncStatus:   app.Status.Sync.Status,
			Automated:    app.Spec.SyncPolicy.Automated != nil,
		}
		updateAvailable := semver.Compare("v"+app.Spec.Source.TargetRevision, "v"+hc.Versions[0]) < 0

		if updateAvailable {
			myApp.NewestVersion = hc.Versions[0]
		}
		myApps = append(myApps, myApp)
	}
	c.apps = myApps
	return nil
}

func (c *client) getHelmCharts(app types.ApplicationResponse, charts map[string]*types.HelmChartsResponse) (*types.HelmChartResponse, error) {
	if hc, ok := charts[app.Spec.Source.RepoURL]; ok {
		return hc.Chart(app.Spec.Source.Chart), nil
	}
	hc := &types.HelmChartsResponse{}
	_, err := c.client.R().SetResult(hc).Get(fmt.Sprintf(urlHelmCharts, url.QueryEscape(app.Spec.Source.RepoURL)))
	if err != nil {
		return nil, err
	}
	charts[app.Spec.Source.RepoURL] = hc
	return hc.Chart(app.Spec.Source.Chart), err
}

func (c *client) readApplications() (*types.ApplicationListResponse, error) {
	apps := &types.ApplicationListResponse{}
	_, err := c.client.R().SetResult(apps).Get(urlApplications)
	return apps, err
}
