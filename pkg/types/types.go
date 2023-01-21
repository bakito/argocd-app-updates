package types

type HelmCharts struct {
	Items []HelmChart `json:"items"`
}

func (c HelmCharts) Chart(chart string) *HelmChart {
	for i := range c.Items {
		hc := c.Items[i]
		if hc.Name == chart {
			return &hc
		}
	}
	return nil
}

type HelmChart struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

type Application struct {
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
		Project string `json:"project"`
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
	} `json:"status"`
}
type ApplicationList struct {
	Items []Application `json:"items"`
}

func (l ApplicationList) Helm() []Application {
	var helmApps []Application

	for i := range l.Items {
		app := l.Items[i]
		if app.Status.SourceType == "Helm" {
			helmApps = append(helmApps, app)
		}
	}

	return helmApps
}
