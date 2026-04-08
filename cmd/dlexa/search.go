package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/spf13/cobra"
)

func newSearchCommand(ctx context.Context, runtime runtimeRunner, format *string, noCache *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "search <query>",
		Short:         "Busca rutas lingüísticas y devuelve siguientes pasos accionables.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runtime.RunModule(ctx, "search", modules.Request{Query: strings.TrimSpace(strings.Join(args, " ")), Format: strings.TrimSpace(*format), NoCache: *noCache, Args: append([]string(nil), args...)})
		},
	}
	cmd.SetHelpFunc(func(*cobra.Command, []string) {
		_ = runtime.RenderHelp(ctx, model.HelpEnvelope{
			Command:     "dlexa search",
			Summary:     "Explora contenido normativo y devuelve comandos literales para profundizar.",
			Syntax:      "dlexa search <consulta>",
			Examples:    []string{"dlexa search solo o sólo", "dlexa search tilde en qué"},
			NextSteps:   []string{"Copiá el `next_command` sugerido para entrar al módulo correcto."},
			RecoveryTip: "Si todavía no sabés el módulo adecuado, search es el primer paso correcto.",
		})
	})
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return runtime.HandleSyntaxError(ctx, err, "dlexa search <consulta>")
	})
	cmd.Args = func(_ *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runtime.HandleSyntaxError(ctx, fmt.Errorf("search command requires a query"), "dlexa search <consulta>")
		}
		return nil
	}
	return cmd
}
