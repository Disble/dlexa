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
			Command:      helpCommandEspanolAlDia,
			Summary:      "Consulta un artículo concreto de Español al día cuando ya conocés el slug público exacto.",
			Syntax:       syntaxEspanolAlDia,
			Capabilities: []string{"Abrir un artículo editorial específico de Español al día.", "Consultar contenido ya identificado por URL o por un resultado previo de `search`."},
			InputHints:   []string{"Recibe el slug público exacto del artículo.", "Si todavía necesitás descubrir el slug, `dlexa search <consulta>` es la entrada adecuada."},
			Examples:     []string{"dlexa espanol-al-dia el-adverbio-solo-y-los-pronombres-demostrativos-sin-tilde"},
			AgentNotes:   []string{"El argumento es un slug estable de URL, no una frase libre ni el título completo.", "Usá este comando cuando la ruta ya esté identificada en tu flujo."},
			NextSteps:    []string{"Si todavía no conocés el slug exacto del artículo, seguí con `dlexa search <consulta>`."},
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
