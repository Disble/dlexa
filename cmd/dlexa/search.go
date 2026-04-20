package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/spf13/cobra"
)

var knownSearchSources = []string{commandSearch, commandDPD}

func validateSources(sources []string) error {
	known := make(map[string]struct{}, len(knownSearchSources))
	for _, source := range knownSearchSources {
		known[source] = struct{}{}
	}
	for _, source := range sources {
		if _, ok := known[source]; !ok {
			return fmt.Errorf("unknown source %q: valid sources are %s", source, strings.Join(knownSearchSources, ", "))
		}
	}
	return nil
}

func newSearchCommand(ctx context.Context, runtime runtimeRunner, format *string, noCache *bool) *cobra.Command {
	var sources []string

	cmd := &cobra.Command{
		Use:           "search <query>",
		Short:         "Busca rutas lingüísticas y devuelve sugerencias; algunas son guía diferida.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := validateSources(sources); err != nil {
				return err
			}
			return runtime.RunModule(ctx, commandSearch, modules.Request{Query: strings.TrimSpace(strings.Join(args, " ")), Format: strings.TrimSpace(*format), NoCache: *noCache, Args: append([]string(nil), args...), Sources: append([]string(nil), sources...)})
		},
	}
	cmd.Flags().StringArrayVar(&sources, "source", nil, "limit search to specific provider(s): search, dpd (repeatable; omit to federate all)")
	cmd.SetHelpFunc(func(*cobra.Command, []string) {
		_ = runtime.RenderHelp(ctx, model.HelpEnvelope{
			Command:      "dlexa search",
			Summary:      "Hace discovery semántico sobre superficies normativas y devuelve rutas de consulta dentro de `dlexa`.",
			Syntax:       "dlexa search [--source <id> ...] <consulta>",
			Capabilities: []string{"Transformar una duda en lenguaje natural en candidatos accionables.", "Descubrir entradas DPD, slugs editoriales y FAQs compatibles.", "Acotar la búsqueda por proveedor con `--source`."},
			InputHints:   []string{"Recibe una consulta libre; no necesitás conocer todavía el comando final.", "Con `--source` podés limitar la búsqueda a `search` o `dpd`.", "La salida distingue entre sugerencias ejecutables y guía diferida según la ruta encontrada."},
			Examples:     []string{"dlexa search solo o sólo", "dlexa search --source dpd tilde en qué"},
			AgentNotes:   []string{"Leé la salida como router: `- sugerencia:` suele marcar el siguiente paso accionable.", "En JSON, inspeccioná cada candidato antes de automatizar; `--source` ayuda a controlar el alcance."},
			NextSteps:    []string{"Si querés federación completa, omití `--source`.", "Si ya identificaste el destino exacto, ejecutá el comando sugerido correspondiente."},
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
