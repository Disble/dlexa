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
		Short:         "Busca rutas lingüísticas y devuelve sugerencias; algunas son guía diferida.",
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
			Summary:     "Explora contenido normativo y devuelve sugerencias literales para profundizar; algunas son guía diferida y no comandos CLI ejecutables todavía.",
			Syntax:      "dlexa search <consulta>",
			Examples:    []string{"dlexa search solo o sólo", "dlexa search tilde en qué"},
			NextSteps:   []string{"Leé la salida: `- sugerencia:` indica un siguiente paso ejecutable; `- More info:` indica guía diferida.", "No copies ni ejecutes a ciegas cada `next_command`; algunas sugerencias son guía diferida hasta que exista ese subcomando."},
			RecoveryTip: "Si todavía no sabés el módulo adecuado, search es el primer paso correcto; después verificá si la sugerencia es ejecutable o solo informativa.",
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
