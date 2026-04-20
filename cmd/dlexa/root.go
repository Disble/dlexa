package main

import (
	"context"
	"fmt"
	"io"

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
		return runtime.HandleSyntaxError(ctx, err, syntaxRoot)
	}
	return nil
}

func newRootCommand(ctx context.Context, runtime runtimeRunner) *cobra.Command {
	var format string
	var noCache bool
	var doctorFlag bool
	var versionFlag bool

	root := &cobra.Command{
		Use:           "dlexa <command>",
		Short:         "Consulta dudas normativas del español para agentes.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          newRootArgsValidator(&doctorFlag, &versionFlag),
		RunE:          newRootRunE(ctx, runtime, &doctorFlag, &versionFlag),
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

func newRootArgsValidator(doctorFlag, versionFlag *bool) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if rootSkipsQueryValidation(doctorFlag, versionFlag, args) {
			return nil
		}
		return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
	}
}

func rootSkipsQueryValidation(doctorFlag, versionFlag *bool, args []string) bool {
	return flagValue(doctorFlag) || flagValue(versionFlag) || len(args) == 0
}

func newRootRunE(
	ctx context.Context,
	runtime runtimeRunner,
	doctorFlag *bool,
	versionFlag *bool,
) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if ok, action := rootImmediateAction(ctx, runtime, cmd, args, doctorFlag, versionFlag); ok {
			return action
		}
		return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
	}
}

func rootImmediateAction(
	ctx context.Context,
	runtime runtimeRunner,
	cmd *cobra.Command,
	args []string,
	doctorFlag *bool,
	versionFlag *bool,
) (bool, error) {
	switch {
	case cmd.Flags().Changed("help"):
		return true, rootHelp(ctx, runtime)
	case flagValue(versionFlag):
		return true, runtime.PrintVersion()
	case flagValue(doctorFlag):
		return true, runtime.RunDoctor(ctx)
	case len(args) == 0:
		return true, rootHelp(ctx, runtime)
	default:
		return false, nil
	}
}

func flagValue(flag *bool) bool {
	return flag != nil && *flag
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func rootHelp(ctx context.Context, runtime runtimeRunner) error {
	return runtime.RenderHelp(ctx, model.HelpEnvelope{
		Command:      helpCommandRoot,
		Summary:      "Elegí la superficie correcta para resolver la duda: `dpd` abre entradas directas, `search` descubre rutas y los comandos por slug abren contenido ya identificado.",
		Syntax:       syntaxRoot,
		Capabilities: []string{"Consultar una entrada DPD exacta con `dlexa dpd <termino>`.", "Descubrir qué superficie conviene con `dlexa search <consulta>`.", "Abrir contenido ya identificado con `dlexa espanol-al-dia <slug>`, `dlexa duda-linguistica <slug>` o `dlexa noticia <slug>`."},
		InputHints:   []string{"Una duda en lenguaje natural suele arrancar mejor en `search`.", "Un término puntual del DPD va directo en `dpd`.", "Un slug público exacto va en su comando específico."},
		Examples:     []string{"dlexa dpd basto", "dlexa dpd solo", "dlexa search solo o sólo", "dlexa duda-linguistica cuando-se-escriben-con-tilde-los-adverbios-en-mente", "dlexa noticia preguntas-frecuentes-tilde-en-las-mayusculas"},
		AgentNotes:   []string{"Leé esta ayuda como mapa de superficies: primero elegí comando, después copiá la sintaxis mínima del bloque correspondiente.", "Los ejemplos entre backticks son literales copiables.", "`--format json` sirve para automatización estructurada; Markdown prioriza lectura humana y LLM."},
		NextSteps:    []string{"Si ya conocés la entrada exacta, seguí con `dlexa dpd <termino>`.", "Si partís de una duda en lenguaje natural, seguí con `dlexa search <consulta>`."},
	})
}
