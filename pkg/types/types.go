package types

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
		} `json:"source"`
		Destination struct {
			Namespace string `json:"namespace"`
			Name      string `json:"name"`
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

func (l ApplicationListResponse) Helm() []ApplicationResponse {
	var helmApps []ApplicationResponse

	for i := range l.Items {
		app := l.Items[i]
		if app.Status.SourceType == "Helm" {
			helmApps = append(helmApps, app)
		}
	}

	return helmApps
}

type Applications []Application

func (a Applications) WithUpdates(project string) Applications {
	filtered := Applications{}
	for _, app := range a.ForProject(project) {
		if app.NewestVersion != "" {
			filtered = append(filtered, app)
		}
	}
	return filtered
}

func (a Applications) WithSourceType(sourceType string, project string) Applications {
	filtered := Applications{}
	for _, app := range a.ForProject(project) {
		if app.SourceType == sourceType {
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
	SourceType    string
	RepoURL       string
	Revision      string
	Chart         string
	Version       string
	NewestVersion string

	HealthStatus string
	SyncStatus   string

	Automated bool
}
