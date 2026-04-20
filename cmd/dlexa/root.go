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
		Command:     helpCommandRoot,
		Summary:     "Consulta dudas normativas del español y usá `search` cuando todavía no conocés la ruta exacta.",
		Syntax:      syntaxRoot,
		Examples:    []string{"dlexa dpd basto", "dlexa dpd solo", "dlexa search solo o sólo", "dlexa duda-linguistica cuando-se-escriben-con-tilde-los-adverbios-en-mente", "dlexa noticia preguntas-frecuentes-tilde-en-las-mayusculas"},
		NextSteps:   []string{"Usá `dlexa dpd <termino>` para consultas directas al DPD.", "Si no encontrás el contenido exacto, escalá a `dlexa search <consulta>`."},
		RecoveryTip: "Si la forma del comando falla, revisá esta ayuda antes de reintentar.",
	})
}
