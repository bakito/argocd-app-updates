package types

import "golang.org/x/mod/semver"

type RepoType string

const (
	RepoTypeHelm RepoType = "Helm"
	RepoTypeGit  RepoType = "Git"
)

type HelmChartsResponse struct {
	Items []HelmChartResponse `json:"items"`
}

func (c HelmChartsResponse) Chart(chart string) *HelmChartResponse {
	for i := range c.Items {
		hc := c.Items[i]
		if hc.Name == chart {
			return &hc
		}
	}
	return nil
}

type HelmChartResponse struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

func (c HelmChartResponse) ReleasedVersions() []string {
	var filtered []string
	for _, v := range c.Versions {
		pr := semver.Prerelease("v" + v)
		if pr == "" {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

type ApplicationResponse struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec struct {
		Source struct {
			RepoURL        string `json:"repoURL"`
			TargetRevision string `json:"targetRevision"`
			Chart          string `json:"chart"`
			Path           string `json:"path"`
		} `json:"source"`
		Destination struct {
			Namespace string `json:"namespace"`
			Name      string `json:"name"`
			Server    string `json:"server"`
		} `json:"destination"`
		Project    string `json:"project"`
		SyncPolicy struct {
			Automated *struct{} `json:"automated"`
		} `json:"syncPolicy"`
	} `json:"spec"`
	Status struct {
		Health struct {
			Status string `json:"status"`
		} `json:"health"`
		SourceType string `json:"sourceType"`
		Summary    struct {
			ExternalURLs []string `json:"externalURLs"`
			Images       []string `json:"images"`
		} `json:"summary"`
		Sync struct {
			Status string `json:"status"`
		} `json:"sync"`
	} `json:"status"`
}
type ApplicationListResponse struct {
	Items []ApplicationResponse `json:"items"`
}

type Applications []Application

func (a Applications) WithUpdates(project string) Applications {
	filtered := Applications{}
	for _, app := range a.ForProject(project) {
		if app.LatestVersion != "" {
			filtered = append(filtered, app)
		}
	}
	return filtered
}

func (a Applications) WithRepoType(repoType RepoType, project string) Applications {
	filtered := Applications{}
	for _, app := range a.ForProject(project) {
		if app.RepoType == repoType {
			filtered = append(filtered, app)
		}
	}
	return filtered
}

func (a Applications) ForProject(project string) Applications {
	filtered := Applications{}
	for _, app := range a {
		if project == "" || app.Project == project {
			filtered = append(filtered, app)
		}
	}
	return filtered
}

type Application struct {
	Name          string
	Project       string
	Cluster       string
	RepoType      RepoType
	RepoURL       string
	Revision      string
	Path          string
	Chart         string
	Version       string
	LatestVersion string

	HealthStatus string
	SyncStatus   string

	Automated bool
}
