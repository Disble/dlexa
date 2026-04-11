package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/spf13/cobra"
)

func newDudaLinguisticaCommand(ctx context.Context, runtime runtimeRunner, format *string, noCache *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           commandDudaLinguistica + " <slug>",
		Short:         "Consulta una duda rápida específica de la RAE.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runtime.RunModule(ctx, commandDudaLinguistica, modules.Request{Query: strings.TrimSpace(strings.Join(args, " ")), Format: strings.TrimSpace(*format), NoCache: *noCache, Args: append([]string(nil), args...), Sources: []string{commandDudaLinguistica}})
		},
	}
	cmd.SetHelpFunc(func(*cobra.Command, []string) {
		_ = runtime.RenderHelp(ctx, model.HelpEnvelope{
			Command:     helpCommandDudaLinguistica,
			Summary:     "Consulta una duda rápida concreta de la superficie Duda lingüística de la RAE.",
			Syntax:      syntaxDudaLinguistica,
			Examples:    []string{"dlexa duda-linguistica cuando-se-escriben-con-tilde-los-adverbios-en-mente"},
			NextSteps:   []string{"Usá `dlexa search <consulta>` si todavía no conocés el slug exacto de la duda."},
			RecoveryTip: "Este comando recibe el slug público de la duda, no una consulta libre.",
		})
	})
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return runtime.HandleSyntaxError(ctx, err, syntaxDudaLinguistica)
	})
	cmd.Args = func(_ *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runtime.HandleSyntaxError(ctx, fmt.Errorf("duda-linguistica command requires an article slug"), syntaxDudaLinguistica)
		}
		return nil
	}
	return cmd
}
