package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/spf13/cobra"
)

func newEspanolAlDiaCommand(ctx context.Context, runtime runtimeRunner, format *string, noCache *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           commandEspanolAlDia + " <slug>",
		Short:         "Consulta un artículo específico de Español al día.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runtime.RunModule(ctx, commandEspanolAlDia, modules.Request{Query: strings.TrimSpace(strings.Join(args, " ")), Format: strings.TrimSpace(*format), NoCache: *noCache, Args: append([]string(nil), args...), Sources: []string{commandEspanolAlDia}})
		},
	}
	cmd.SetHelpFunc(func(*cobra.Command, []string) {
		_ = runtime.RenderHelp(ctx, model.HelpEnvelope{
			Command:     helpCommandEspanolAlDia,
			Summary:     "Consulta un artículo concreto de la superficie Español al día de la RAE.",
			Syntax:      syntaxEspanolAlDia,
			Examples:    []string{"dlexa espanol-al-dia el-adverbio-solo-y-los-pronombres-demostrativos-sin-tilde"},
			NextSteps:   []string{"Usá `dlexa search <consulta>` si todavía no conocés el slug exacto del artículo."},
			RecoveryTip: "Este comando recibe el slug público del artículo, no una consulta libre.",
		})
	})
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return runtime.HandleSyntaxError(ctx, err, syntaxEspanolAlDia)
	})
	cmd.Args = func(_ *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runtime.HandleSyntaxError(ctx, fmt.Errorf("espanol-al-dia command requires an article slug"), syntaxEspanolAlDia)
		}
		return nil
	}
	return cmd
}
