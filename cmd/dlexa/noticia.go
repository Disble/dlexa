package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/spf13/cobra"
)

func newNoticiaCommand(ctx context.Context, runtime runtimeRunner, format *string, noCache *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "noticia <slug>",
		Short:         "Consulta una pregunta frecuente publicada como noticia en la RAE.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runtime.RunModule(ctx, "noticia", modules.Request{Query: strings.TrimSpace(strings.Join(args, " ")), Format: strings.TrimSpace(*format), NoCache: *noCache, Args: append([]string(nil), args...), Sources: []string{"noticia"}})
		},
	}
	cmd.SetHelpFunc(func(*cobra.Command, []string) {
		_ = runtime.RenderHelp(ctx, model.HelpEnvelope{
			Command:      "dlexa noticia",
			Summary:      "Consulta una FAQ normativa publicada bajo la superficie Noticia cuando ya conocés el slug exacto.",
			Syntax:       syntaxNoticia,
			Capabilities: []string{"Abrir una FAQ normativa ya identificada dentro de la superficie Noticia.", "Consultar una FAQ compatible a partir de un slug devuelto por `search` o por una URL conocida."},
			InputHints:   []string{"Recibe el slug público exacto de la FAQ.", "Si todavía necesitás descubrir la ruta normativa correcta, `dlexa search <consulta>` te ayuda a encontrarla."},
			Examples:     []string{"dlexa noticia preguntas-frecuentes-tilde-en-las-mayusculas"},
			AgentNotes:   []string{"Este comando apunta a FAQs normativas compatibles dentro de la superficie Noticia.", "Usalo cuando tu flujo ya tenga identificada una sugerencia ejecutable hacia `noticia`."},
			NextSteps:    []string{"Si todavía no conocés el slug exacto de la FAQ, seguí con `dlexa search <consulta>`."},
		})
	})
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return runtime.HandleSyntaxError(ctx, err, syntaxNoticia)
	})
	cmd.Args = func(_ *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runtime.HandleSyntaxError(ctx, fmt.Errorf("noticia command requires an article slug"), syntaxNoticia)
		}
		return nil
	}
	return cmd
}
