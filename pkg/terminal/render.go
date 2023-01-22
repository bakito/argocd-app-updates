package terminal

import (
	"fmt"
	"os"
	"strings"

	"github.com/bakito/argocd-app-updates/pkg/types"
	"github.com/fatih/color"
	"github.com/juju/ansiterm/tabwriter"
)

var (
	colorGreen   = color.New(color.FgGreen)
	colorYellow  = color.New(color.FgYellow)
	colorBlue    = color.New(color.FgCyan)
	colorRed     = color.New(color.FgRed)
	colorMagenta = color.New(color.FgHiMagenta)
	colorHiCyan  = color.New(color.FgHiCyan)
)

func Render(apps types.Applications) {
	w := tabwriter.NewWriter(os.Stdout, 6, 4, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, strings.Join([]string{
		"PROJECT",
		"NAME",
		"HEALTH STATUS",
		"SYNC STATUS",
		"AUTO SYNC",
		"CHART",
		"VERSION",
		"LATEST",
	}, "\t"))

	for _, app := range apps {
		var version string
		if app.LatestVersion != "" {
			version = colorYellow.Sprint(app.LatestVersion)
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
