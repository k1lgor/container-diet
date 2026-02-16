package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var version = "0.3.0"

var rootCmd = &cobra.Command{
	Use:     "container-diet",
	Short:   "AI-powered Docker image slimming assistant",
	Long:    "Container Diet helps you inspect Docker images and Dockerfiles, then gives practical AI optimization advice.",
	Example: "  container-diet analyze nginx:latest\n  container-diet analyze my-app:latest --dockerfile Dockerfile\n  container-diet analyze python:3.11-slim --remote\n  container-diet analyze busybox --pull-missing",
	Version: version,
}

func init() {
	rootCmd.SetHelpTemplate(buildHelpTemplate())
}

func buildHelpTemplate() string {
	title := color.New(color.FgHiCyan, color.Bold).Sprint("🐳 Container Diet CLI")
	usage := color.New(color.FgHiMagenta, color.Bold).Sprint("Usage:")
	commands := color.New(color.FgHiMagenta, color.Bold).Sprint("Commands:")
	flags := color.New(color.FgHiMagenta, color.Bold).Sprint("Flags:")
	globalFlags := color.New(color.FgHiMagenta, color.Bold).Sprint("Global Flags:")
	examples := color.New(color.FgHiMagenta, color.Bold).Sprint("Examples:")
	return fmt.Sprintf(`%s

{{with .Long}}{{.}}{{"\n\n"}}{{end}}%s
  {{.UseLine}}

%s
{{range .Commands}}{{if (or .IsAvailableCommand .IsAdditionalHelpTopicCommand)}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}
{{if .HasAvailableLocalFlags}}%s
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
{{if .HasAvailableInheritedFlags}}%s
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
{{if .HasExample}}%s
%s{{end}}
`, title, usage, commands, flags, globalFlags, examples, strings.TrimSpace("{{.Example}}"))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
