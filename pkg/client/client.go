package client

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/bakito/argocd-app-updates/pkg/types"
	"github.com/go-resty/resty/v2"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/mod/semver"
)

const (
	apiV1           = "/api/v1/"
	urlApplications = apiV1 + "applications"
	urlHelmCharts   = apiV1 + "repositories/%s/helmcharts"
	urlSettings     = apiV1 + "settings"
	urlSession      = apiV1 + "session"
)

func NewClient(argoServer string, username string, password string) Client {
	metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "argocd_app_update_available",
		Help: "1 if an update is available for the argocd application",
	}, []string{"project", "name", "current_version", "latest_version"},
	)
	prometheus.MustRegister(metric)

	cl := &client{
		client: resty.New().SetBaseURL(argoServer),
		url:    argoServer,
		metric: metric,
	}

	if username != "" && password != "" {
		cl.auth = &sessionRequest{
			Username: username,
			Password: password,
		}
	}
	return cl
}

type Client interface {
	Update() error
	Applications() types.Applications
	URL() string
}

type client struct {
	client *resty.Client
	apps   types.Applications
	url    string
	auth   *sessionRequest
	token  string
	metric *prometheus.GaugeVec
}

func (c *client) URL() string {
	return c.url
}

func (c *client) Applications() types.Applications {
	return c.apps
}

func (c *client) Update() error {
	if c.auth != nil {
		var err error
		c.token, err = c.login()
		if err != nil {
			return err
		}
	}

	s, err := c.readSettings()
	if err != nil {
		return err
	}
	if s.URL != "" {
		c.url = strings.TrimSuffix(s.URL, "/")
	}

	apps, err := c.readApplications()
	if err != nil {
		return err
	}

	charts := make(map[string]*types.HelmChartsResponse)

	var myApps types.Applications

	for _, app := range apps.Items {
		myApp := types.Application{
			Name:         app.Metadata.Name,
			Project:      app.Spec.Project,
			RepoURL:      app.Spec.Source.RepoURL,
			Revision:     app.Spec.Source.TargetRevision,
			Path:         app.Spec.Source.Path,
			Chart:        app.Spec.Source.Chart,
			Version:      app.Spec.Source.TargetRevision,
			HealthStatus: app.Status.Health.Status,
			SyncStatus:   app.Status.Sync.Status,
			Automated:    app.Spec.SyncPolicy.Automated != nil,
		}

		if app.Spec.Source.Path != "" {
			myApp.RepoType = types.RepoTypeGit
		} else {
			myApp.RepoType = types.RepoTypeHelm

			hc, err := c.getHelmCharts(app, charts)
			if err != nil {
				return err
			}

			if hc != nil {
				rv := hc.ReleasedVersions()
				if len(rv) != 0 {
					if semver.Compare("v"+app.Spec.Source.TargetRevision, "v"+rv[0]) < 0 {
						myApp.LatestVersion = rv[0]
					}
				}
			}
		}
		myApps = append(myApps, myApp)
	}
	for _, app := range c.apps {
		c.metric.DeleteLabelValues(app.Project, app.Name, app.Version, app.LatestVersion)
	}
	for _, app := range myApps {
		var val float64
		if app.LatestVersion != "" {
			val = 1
		}
		c.metric.WithLabelValues(app.Project, app.Name, app.Version, app.LatestVersion).Set(val)
	}
	c.apps = myApps
	return nil
}

func (c *client) getHelmCharts(app types.ApplicationResponse,
	charts map[string]*types.HelmChartsResponse,
) (*types.HelmChartResponse, error) {
	if hc, ok := charts[app.Spec.Source.RepoURL]; ok {
		return hc.Chart(app.Spec.Source.Chart), nil
	}
	hc := &types.HelmChartsResponse{}
	resp, err := c.client.R().
		SetAuthToken(c.token).
		SetResult(hc).
		Get(fmt.Sprintf(urlHelmCharts, url.QueryEscape(app.Spec.Source.RepoURL)))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("%s %s", resp.String(), resp.Request.URL)
	}
	charts[app.Spec.Source.RepoURL] = hc
	return hc.Chart(app.Spec.Source.Chart), err
}

func (c *client) readApplications() (*types.ApplicationListResponse, error) {
	apps := &types.ApplicationListResponse{}
	resp, err := c.client.R().
		SetAuthToken(c.token).
		SetResult(apps).
		Get(urlApplications)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("%s %s", resp.String(), resp.Request.URL)
	}
	sort.Slice(apps.Items, func(i, j int) bool {
		return apps.Items[i].Metadata.Name < apps.Items[j].Metadata.Name
	})
	return apps, err
}

func (c *client) readSettings() (*settings, error) {
	s := &settings{}
	resp, err := c.client.R().
		SetAuthToken(c.token).
		SetResult(s).
		Get(urlSettings)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("%s %s", resp.String(), resp.Request.URL)
	}
	return s, err
}

func (c *client) login() (string, error) {
	s := &sessionResponse{}
	resp, err := c.client.R().
		SetBody(c.auth).
		SetResult(s).
		Post(urlSession)
	if err != nil {
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("%s %s", resp.String(), resp.Request.URL)
	}
	return s.Token, err
}

type settings struct {
	URL string `json:"url"`
}
type sessionResponse struct {
	Token string `json:"token"`
}
type sessionRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
