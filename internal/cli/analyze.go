package cli

import (
	"fmt"
	"os"
	"strings"
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
	pullMissing    bool
	neon           = color.New(color.FgHiCyan, color.Bold).SprintFunc()
	aurora         = color.New(color.FgHiMagenta, color.Bold).SprintFunc()
	success        = color.New(color.FgHiGreen, color.Bold).SprintFunc()
	warn           = color.New(color.FgHiYellow, color.Bold).SprintFunc()
	fail           = color.New(color.FgHiRed, color.Bold).SprintFunc()
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [image]",
	Short: "Analyze a Docker image and/or Dockerfile",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		printSection("CONTAINER DIET ANALYZER")

		// Enforce API Key
		if os.Getenv("OPENAI_API_KEY") == "" {
			fmt.Fprintln(os.Stderr, fail("✖ Error: OPENAI_API_KEY environment variable is not set."))
			fmt.Fprintln(os.Stderr, warn("This tool requires an OpenAI API key for analysis."))
			os.Exit(1)
		}

		var analysis *analyzer.ImageAnalysis
		var err error

		// Analyze Image if provided
		if len(args) > 0 {
			imageName := args[0]
			fmt.Printf("%s %s\n", neon("🔍 Scanning image:"), aurora(imageName))

			analysis, err = analyzer.AnalyzeImage(imageName, remote, pullMissing)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s %v\n", fail("✖ Error analyzing image:"), err)
				os.Exit(1)
			}
		}

		// Read Dockerfile if provided
		var dockerfileContent string
		if dockerfilePath != "" {
			fmt.Printf("%s %s\n", neon("📄 Reading Dockerfile:"), aurora(dockerfilePath))
			content, err := os.ReadFile(dockerfilePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s %v\n", fail("✖ Error reading Dockerfile:"), err)
				os.Exit(1)
			}
			dockerfileContent = string(content)
		}

		if analysis == nil && dockerfileContent == "" {
			fmt.Println(warn("⚠ Please provide an image name or a Dockerfile path."))
			cmd.Help()
			os.Exit(1)
		}

		// Print Image Analysis Summary
		printAnalysisSummary(analysis)

		// AI Analysis
		fmt.Printf("\n%s\n", neon("🤖 [AI ANALYSIS]"))

		aiClient, err := ai.NewOpenAIClient()
		if err != nil {
			// Should not happen due to check above, but good for safety
			fmt.Fprintf(os.Stderr, "%s %v\n", fail("✖ Error initializing AI client:"), err)
			os.Exit(1)
		}

		fmt.Println(aurora("🚢 Asking the Container Dietician for insights..."))
		advice, err := aiClient.AnalyzeReport(analysis, dockerfileContent, model)
		if err != nil {
			fmt.Printf("%s %v\n", fail("✖ Error getting AI advice:"), err)
		} else {
			fmt.Printf("\n%s\n%s\n", neon(strings.Repeat("=", 64)), advice)
			fmt.Println(success("✓ Analysis complete"))
		}
	},
}

func init() {
	analyzeCmd.Flags().StringVarP(&format, "format", "f", "text", "Output format (text or json)")
	analyzeCmd.Flags().StringVar(&dockerfilePath, "dockerfile", "", "Path to Dockerfile for analysis")
	analyzeCmd.Flags().StringVar(&model, "model", "gpt-4o", "AI model to use for analysis")
	analyzeCmd.Flags().BoolVar(&remote, "remote", false, "Allow pulling image from remote registry if not found locally")
	analyzeCmd.Flags().BoolVar(&pullMissing, "pull-missing", false, "When local image is missing, pull it from remote and continue analysis")
	rootCmd.AddCommand(analyzeCmd)
}

func printAnalysisSummary(analysis *analyzer.ImageAnalysis) {
	if analysis != nil {
		fmt.Printf("\n%s\n", neon("📊 IMAGE SUMMARY"))
		fmt.Printf("\n📦 Image: %s\n", aurora(analysis.ImageName))
		fmt.Printf("📊 Total Size: %s\n", success(fmt.Sprintf("%.2f MB", float64(analysis.TotalSize)/1024/1024)))
		fmt.Printf("🍰 Layers: %s\n", neon(fmt.Sprintf("%d", len(analysis.Layers))))

		fmt.Printf("\n%s\n", aurora("--- Layer Analysis ---"))
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, neon("INDEX")+"\t"+neon("SIZE (MB)")+"\t"+neon("COMMAND"))
		for i, l := range analysis.Layers {
			cmd := l.Command
			if len(cmd) > 50 {
				cmd = cmd[:47] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", aurora(fmt.Sprintf("%d", i+1)), success(fmt.Sprintf("%.2f", float64(l.Size)/1024/1024)), cmd)
		}
		w.Flush()
	}
}

func printSection(title string) {
	line := strings.Repeat("=", len(title)+4)
	fmt.Printf("\n%s\n%s\n%s\n", neon(line), aurora("  "+title+"  "), neon(line))
}
