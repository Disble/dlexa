package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/spf13/cobra"
)

func newDPDCommand(ctx context.Context, runtime runtimeRunner, format *string, noCache *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "dpd <query>",
		Short:         "Consulta una entrada del Diccionario panhispánico de dudas.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runtime.RunModule(ctx, "dpd", modules.Request{Query: strings.TrimSpace(strings.Join(args, " ")), Format: strings.TrimSpace(*format), NoCache: *noCache, Args: append([]string(nil), args...)})
		},
	}
	cmd.SetHelpFunc(func(*cobra.Command, []string) {
		_ = runtime.RenderHelp(ctx, model.HelpEnvelope{
			Command:     "dlexa dpd",
			Summary:     "Consulta una entrada concreta del Diccionario panhispánico de dudas.",
			Syntax:      "dlexa dpd <termino>",
			Examples:    []string{"dlexa dpd basto", "dlexa dpd que"},
			NextSteps:   []string{"Si no aparece la entrada, probá primero `dlexa search <consulta>`."},
			RecoveryTip: "Si te equivocás de sintaxis, mirá esta ayuda y corregí el comando.",
		})
	})
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return runtime.HandleSyntaxError(ctx, err, "dlexa dpd <termino>")
	})
	cmd.Args = func(_ *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runtime.HandleSyntaxError(ctx, fmt.Errorf("dpd command requires a query"), "dlexa dpd <termino>")
		}
		return nil
	}
	return cmd
}
