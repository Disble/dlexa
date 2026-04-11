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
			Command:     "dlexa search",
			Summary:     "Explora contenido normativo y devuelve sugerencias literales para profundizar; algunas rutas ya son ejecutables y otras siguen como guía diferida.",
			Syntax:      "dlexa search [--source <id> ...] <consulta>",
			Examples:    []string{"dlexa search solo o sólo", "dlexa search --source dpd tilde en qué"},
			NextSteps:   []string{"Usá `--source search` o `--source dpd` para acotar proveedores; omitilo para federar todos.", "Leé la salida: `- sugerencia:` indica un siguiente paso ejecutable; los bloques con URL y acceso futuro siguen siendo guía diferida.", "No copies ni ejecutes a ciegas cada `next_command`; verificá primero si la sugerencia se presenta como ejecutable o solo informativa."},
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
