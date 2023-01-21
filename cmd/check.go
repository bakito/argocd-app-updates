package cmd

import (
	"context"
	"github.com/argoproj/argo-cd/v2/cmd/argocd/commands/headless"
	argocdclient "github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	argoio "github.com/argoproj/argo-cd/v2/util/io"
	"github.com/argoproj/pkg/errors"
	"github.com/spf13/cobra"
)

func newCheckCommand(globalClientOpts *argocdclient.ClientOptions) *cobra.Command {

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Check Argo CD Applications for Updates",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()

			acdClient := headless.NewClientOrDie(globalClientOpts, cmd)
			conn, appIf := acdClient.NewApplicationClientOrDie()
			defer argoio.Close(conn)

			apps, err := appIf.List(ctx, &application.ApplicationQuery{})
			errors.CheckError(err)
			println(apps)

		},
	}
	return checkCmd
}
