package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/k1lgor/container-diet/internal/config"
	"github.com/spf13/cobra"
)

var (
	initLocal   bool
	initForce   bool
)

var initConfigCmd = &cobra.Command{
	Use:   "init-config",
	Short: "Create an example configuration file",
	Long: `Creates an example configuration file with sample provider configurations.

By default, creates a global config in ~/.config/container-diet/config.yaml
Use --local to create a project-specific config in ./.container-diet/config.yaml
Use --force to overwrite an existing config file.

Config hierarchy (later overrides earlier):
  1. Global: ~/.config/container-diet/config.yaml
  2. Local:  ./.container-diet/config.yaml (project-specific)`,
	Run: func(cmd *cobra.Command, args []string) {
		var configPath string
		var configType string

		if initLocal {
			configPath = filepath.Join(".container-diet", "config.yaml")
			configType = "local"
		} else {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Failed to get home directory: %v\n", fail("✖ Error:"), err)
				os.Exit(1)
			}
			configPath = filepath.Join(homeDir, ".config", "container-diet", "config.yaml")
			configType = "global"
		}

		// Check if config already exists
		if _, err := os.Stat(configPath); err == nil && !initForce {
			fmt.Printf("%s %s configuration file already exists at:\n   %s\n", warn("⚠"), configType, configPath)
			fmt.Println("Edit the existing file, delete it, or use --force to overwrite.")
			return
		}

		// Create example config
		if err := config.SaveExampleConfig(configPath); err != nil {
			fmt.Fprintf(os.Stderr, "%s Failed to create config: %v\n", fail("✖ Error:"), err)
			os.Exit(1)
		}

		fmt.Printf("%s Created %s configuration file:\n   %s\n\n", success("✓"), configType, configPath)

		if initLocal {
			fmt.Println(neon("Project-specific configuration created!"))
			fmt.Println("This config will override the global config when running from this directory.")
			fmt.Println()
			fmt.Println("Next steps:")
			fmt.Println("1. Edit the configuration file with your project-specific settings")
			fmt.Println("2. Add .container-diet/ to your .gitignore to avoid committing secrets")
			fmt.Println("3. Run 'container-diet analyze <image>' to use this config")
		} else {
			fmt.Println(neon("Global configuration created!"))
			fmt.Println()
			fmt.Println("Next steps:")
			fmt.Println("1. Edit the configuration file with your API keys")
			fmt.Println("2. Set environment variables for sensitive values (e.g., OPENAI_API_KEY)")
			fmt.Println("3. Run 'container-diet analyze <image>' to start using your configured provider")
			fmt.Println()
			fmt.Println("For project-specific settings, run:")
			fmt.Println("  container-diet init-config --local")
		}
	},
}

func init() {
	initConfigCmd.Flags().BoolVar(&initLocal, "local", false, "Create a local/project-specific config instead of global")
	initConfigCmd.Flags().BoolVar(&initForce, "force", false, "Overwrite existing config file")
	rootCmd.AddCommand(initConfigCmd)
}
