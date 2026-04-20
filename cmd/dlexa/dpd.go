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
			return runtime.RunModule(ctx, commandDPD, modules.Request{Query: strings.TrimSpace(strings.Join(args, " ")), Format: strings.TrimSpace(*format), NoCache: *noCache, Args: append([]string(nil), args...)})
		},
	}
	cmd.AddCommand(newDPDSearchCommand(ctx, runtime, format, noCache))
	cmd.SetHelpFunc(func(*cobra.Command, []string) {
		_ = runtime.RenderHelp(ctx, model.HelpEnvelope{
			Command:      "dlexa dpd",
			Summary:      "Consulta una entrada concreta del Diccionario panhispánico de dudas cuando ya conocés el término o locución que querés abrir.",
			Syntax:       syntaxDPD,
			Capabilities: []string{"Consultar una entrada puntual del DPD para dudas ortográficas, gramaticales o léxico-semánticas.", "Ir directo a una entrada cuando ya conocés el término o una forma muy cercana."},
			InputHints:   []string{"Recibe el término de entrada o una forma muy cercana.", "Si todavía estás explorando la superficie correcta, `dlexa search <consulta>` te ayuda a descubrirla."},
			Examples:     []string{"dlexa dpd basto", "dlexa dpd que"},
			AgentNotes:   []string{"La consulta se interpreta como término de entrada, no como instrucción conversacional larga.", "Si el término exacto cambia de forma o deriva en otra ruta, usá `search` para descubrir la superficie correcta."},
			NextSteps:    []string{"Si querés descubrir variantes o superficies relacionadas, seguí con `dlexa search <consulta>`."},
		})
	})
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return runtime.HandleSyntaxError(ctx, err, syntaxDPD)
	})
	cmd.Args = func(_ *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runtime.HandleSyntaxError(ctx, fmt.Errorf("dpd command requires a query"), syntaxDPD)
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
			return runtime.RunModule(ctx, commandSearch, modules.Request{
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
			Command:      "dlexa dpd search",
			Summary:      "Busca en el índice específico del DPD cuando querés discovery acotado al propio DPD.",
			Syntax:       syntaxDPDSearch,
			Capabilities: []string{"Explorar coincidencias del índice DPD.", "Descubrir candidatos dentro del DPD antes de abrir una entrada final."},
			InputHints:   []string{"Recibe un término o fragmento para buscar coincidencias en el índice DPD.", "Mantiene el contrato de `search`, pero limitado al proveedor DPD."},
			Examples:     []string{"dlexa dpd search Abu Dhabi", "dlexa dpd search solo"},
			AgentNotes:   []string{"La salida puede incluir sugerencias ejecutables o guía para el siguiente paso.", "Si ya conocés el término exacto, podés ir directo a `dlexa dpd <termino>`."},
			NextSteps:    []string{"Si querés federación entre buscadores, seguí con `dlexa search <consulta>`.", "Si un candidato ya te da la entrada exacta, seguí con `dlexa dpd <termino>`."},
		})
	})
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return runtime.HandleSyntaxError(ctx, err, syntaxDPDSearch)
	})
	cmd.Args = func(_ *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runtime.HandleSyntaxError(ctx, fmt.Errorf("dpd search command requires a query"), syntaxDPDSearch)
		}
		return nil
	}
	return cmd
}
