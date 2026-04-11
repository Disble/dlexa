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
			Command:     "dlexa noticia",
			Summary:     "Consulta una FAQ normativa publicada bajo la superficie Noticia de la RAE.",
			Syntax:      syntaxNoticia,
			Examples:    []string{"dlexa noticia preguntas-frecuentes-tilde-en-las-mayusculas"},
			NextSteps:   []string{"Usá `dlexa search <consulta>` si todavía no conocés el slug exacto de la FAQ."},
			RecoveryTip: "Este comando acepta solo slugs de noticias FAQ compatibles; no sirve para noticias institucionales generales.",
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
