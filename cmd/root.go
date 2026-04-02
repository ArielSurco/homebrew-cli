package cmd

import (
	"context"
	"image/color"

	"charm.land/fang/v2"
	fanglipgloss "charm.land/lipgloss/v2"
	cmdproject "github.com/ArielSurco/cli/cmd/project"
	"github.com/spf13/cobra"
)

// Version is injected at build time.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:          "arielsurco-cli",
	Short:        "Personal developer CLI",
	Version:      Version,
	SilenceUsage: true,
}

// Execute runs the root command via fang with a customized color scheme
// that removes the codeblock background box.
func Execute() error {
	return fang.Execute(context.Background(), rootCmd,
		fang.WithColorSchemeFunc(func(lightDark fanglipgloss.LightDarkFunc) fang.ColorScheme {
			scheme := fang.DefaultColorScheme(lightDark)
			scheme.Codeblock = color.Transparent
			return scheme
		}),
	)
}

func init() {
	rootCmd.AddCommand(cmdproject.Cmd)
}
