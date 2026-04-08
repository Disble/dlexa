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
	cmd.AddCommand(newDPDSearchCommand(ctx, runtime, format, noCache))
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

func newDPDSearchCommand(ctx context.Context, runtime runtimeRunner, format *string, noCache *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "search <termino-de-busqueda>",
		Short:         "Busca términos del índice específico del DPD.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runtime.RunModule(ctx, "search", modules.Request{
				Query:   strings.TrimSpace(strings.Join(args, " ")),
				Format:  strings.TrimSpace(*format),
				NoCache: *noCache,
				Args:    append([]string(nil), args...),
				Sources: []string{"dpd"},
			})
		},
	}
	cmd.SetHelpFunc(func(*cobra.Command, []string) {
		_ = runtime.RenderHelp(ctx, model.HelpEnvelope{
			Command:     "dlexa dpd search",
			Summary:     "Busca en el índice específico del DPD y devuelve candidatos en el formato del gateway search.",
			Syntax:      "dlexa dpd search <termino-de-busqueda>",
			Examples:    []string{"dlexa dpd search Abu Dhabi", "dlexa dpd search solo"},
			NextSteps:   []string{"Usá este comando cuando quieras consultar solo el buscador específico del DPD sin mezclarlo con el search general de la RAE."},
			RecoveryTip: "Si querés federación entre buscadores, usá `dlexa search <consulta>`; si querés solo el índice DPD, usá este subcomando.",
		})
	})
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return runtime.HandleSyntaxError(ctx, err, "dlexa dpd search <termino-de-busqueda>")
	})
	cmd.Args = func(_ *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runtime.HandleSyntaxError(ctx, fmt.Errorf("dpd search command requires a query"), "dlexa dpd search <termino-de-busqueda>")
		}
		return nil
	}
	return cmd
}
