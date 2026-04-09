package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/spf13/cobra"
)

type runtimeRunner interface {
	RunModule(ctx context.Context, module string, req modules.Request) error
	RenderHelp(ctx context.Context, help model.HelpEnvelope) error
	HandleSyntaxError(ctx context.Context, err error, syntax string) error
	RunDoctor(ctx context.Context) error
	PrintVersion() error
}

func executeRootCommand(ctx context.Context, runtime runtimeRunner, stdout io.Writer, stderr io.Writer, args []string) error {
	root := newRootCommand(ctx, runtime)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs(args)
	if err := root.ExecuteContext(ctx); err != nil {
		return runtime.HandleSyntaxError(ctx, err, "dlexa <query>")
	}
	return nil
}

func newRootCommand(ctx context.Context, runtime runtimeRunner) *cobra.Command {
	var format string
	var noCache bool
	var doctorFlag bool
	var versionFlag bool

	root := &cobra.Command{
		Use:           "dlexa [query]",
		Short:         "Consulta dudas normativas del español para agentes.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args: func(cmd *cobra.Command, args []string) error {
			if doctorFlag || versionFlag {
				return nil
			}
			if len(args) == 0 {
				return nil
			}
			if len(args) > 0 && looksLikeUnknownSyntax(args) {
				return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("help") {
				return rootHelp(ctx, runtime)
			}
			if versionFlag {
				return runtime.PrintVersion()
			}
			if doctorFlag {
				return runtime.RunDoctor(ctx)
			}
			if len(args) == 0 {
				return rootHelp(ctx, runtime)
			}
			return runtime.RunModule(ctx, "dpd", modules.Request{Query: strings.TrimSpace(strings.Join(args, " ")), Format: format, NoCache: noCache, Args: append([]string(nil), args...)})
		},
	}
	root.PersistentFlags().StringVar(&format, "format", "", "render format: markdown|json")
	root.PersistentFlags().BoolVar(&noCache, "no-cache", false, "skip cache reads and writes")
	root.PersistentFlags().BoolVar(&doctorFlag, "doctor", false, "run environment checks")
	root.PersistentFlags().BoolVar(&versionFlag, "version", false, "print version information")
	root.SetHelpFunc(func(*cobra.Command, []string) { _ = rootHelp(ctx, runtime) })
	root.AddCommand(newDPDCommand(ctx, runtime, &format, &noCache))
	root.AddCommand(newSearchCommand(ctx, runtime, &format, &noCache))
	root.AddCommand(newEspanolAlDiaCommand(ctx, runtime, &format, &noCache))
	root.AddCommand(newDudaLinguisticaCommand(ctx, runtime, &format, &noCache))
	root.AddCommand(newNoticiaCommand(ctx, runtime, &format, &noCache))
	return root
}

func rootHelp(ctx context.Context, runtime runtimeRunner) error {
	return runtime.RenderHelp(ctx, model.HelpEnvelope{
		Command:     "dlexa",
		Summary:     "Consulta dudas normativas del español y usá `search` cuando todavía no conocés la ruta exacta.",
		Syntax:      "dlexa <query>",
		Examples:    []string{"dlexa basto", "dlexa dpd solo", "dlexa search solo o sólo", "dlexa duda-linguistica cuando-se-escriben-con-tilde-los-adverbios-en-mente", "dlexa noticia preguntas-frecuentes-tilde-en-las-mayusculas"},
		NextSteps:   []string{"Si no encontrás el contenido exacto, escalá a `dlexa search <consulta>`."},
		RecoveryTip: "Si la forma del comando falla, revisá esta ayuda antes de reintentar.",
	})
}

func looksLikeUnknownSyntax(args []string) bool {
	if len(args) == 0 {
		return false
	}
	first := strings.TrimSpace(args[0])
	if first == "" {
		return false
	}
	return strings.HasPrefix(first, "-") || (len(args) > 1 && strings.HasPrefix(strings.TrimSpace(args[1]), "-")) && first != "search" && first != "dpd" && first != "espanol-al-dia" && first != "duda-linguistica" && first != "noticia"
}
