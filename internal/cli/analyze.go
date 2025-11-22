package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/k1lgor/container-diet/internal/ai"
	"github.com/k1lgor/container-diet/internal/analyzer"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	format         string
	dockerfilePath string
	model          string
	remote         bool
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [image]",
	Short: "Analyze a Docker image and/or Dockerfile",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Enforce API Key
		if os.Getenv("OPENAI_API_KEY") == "" {
			fmt.Fprintln(os.Stderr, "Error: OPENAI_API_KEY environment variable is not set.")
			fmt.Fprintln(os.Stderr, "This tool requires an OpenAI API key for analysis.")
			os.Exit(1)
		}

		var analysis *analyzer.ImageAnalysis
		var err error

		// Analyze Image if provided
		if len(args) > 0 {
			imageName := args[0]
			fmt.Printf("Analyzing image: %s...\n", imageName)

			analysis, err = analyzer.AnalyzeImage(imageName, remote)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error analyzing image: %v\n", err)
				os.Exit(1)
			}
		}

		// Read Dockerfile if provided
		var dockerfileContent string
		if dockerfilePath != "" {
			fmt.Printf("Reading Dockerfile: %s...\n", dockerfilePath)
			content, err := os.ReadFile(dockerfilePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading Dockerfile: %v\n", err)
				os.Exit(1)
			}
			dockerfileContent = string(content)
		}

		if analysis == nil && dockerfileContent == "" {
			fmt.Println("Please provide an image name or a Dockerfile path.")
			cmd.Help()
			os.Exit(1)
		}

		// Print Image Analysis Summary
		printAnalysisSummary(analysis)

		// AI Analysis
		blue := color.New(color.FgHiBlue, color.Bold).SprintFunc()
		fmt.Printf("\n%s\n", blue("🐳 [AI ANALYSIS COMPLETE]"))

		aiClient, err := ai.NewOpenAIClient()
		if err != nil {
			// Should not happen due to check above, but good for safety
			fmt.Fprintf(os.Stderr, "Error initializing AI client: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Asking the Container Dietician for insights... 🚢")
		advice, err := aiClient.AnalyzeReport(analysis, dockerfileContent, model)
		if err != nil {
			fmt.Printf("Error getting AI advice: %v\n", err)
		} else {
			fmt.Println(advice)
		}
	},
}

func init() {
	analyzeCmd.Flags().StringVarP(&format, "format", "f", "text", "Output format (text or json)")
	analyzeCmd.Flags().StringVar(&dockerfilePath, "dockerfile", "", "Path to Dockerfile for analysis")
	analyzeCmd.Flags().StringVar(&model, "model", "gpt-4o", "AI model to use for analysis")
	analyzeCmd.Flags().BoolVar(&remote, "remote", false, "Allow pulling image from remote registry if not found locally")
	rootCmd.AddCommand(analyzeCmd)
}

func printAnalysisSummary(analysis *analyzer.ImageAnalysis) {
	blue := color.New(color.FgHiBlue, color.Bold).SprintFunc()

	if analysis != nil {
		fmt.Printf("\n📦 Image: %s\n", analysis.ImageName)
		fmt.Printf("📊 Total Size: %.2f MB\n", float64(analysis.TotalSize)/1024/1024)
		fmt.Printf("🍰 Layers: %d\n", len(analysis.Layers))

		fmt.Printf("\n%s\n", blue("--- Layer Analysis ---"))
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "INDEX\tSIZE (MB)\tCOMMAND")
		for i, l := range analysis.Layers {
			cmd := l.Command
			if len(cmd) > 50 {
				cmd = cmd[:47] + "..."
			}
			fmt.Fprintf(w, "%d\t%.2f\t%s\n", i+1, float64(l.Size)/1024/1024, cmd)
		}
		w.Flush()
	}
}
