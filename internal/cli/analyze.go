package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/k1lgor/container-diet/internal/ai"
	"github.com/k1lgor/container-diet/internal/analyzer"
	"github.com/k1lgor/container-diet/internal/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	format         string
	dockerfilePath string
	model          string
	remote         bool
	pullMissing    bool
	autoFix        bool
	providerName   string
	configPath     string
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
		if err := runAnalyze(args); err != nil {
			os.Exit(1)
		}
	},
}

// runAnalyze is the testable core of the analyze command.
// It prints user-facing output directly and returns an error on failure.
func runAnalyze(args []string) error {
	printSection("CONTAINER DIET ANALYZER")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", fail("✖ Error loading config:"), err)
		return err
	}

	if providerName == "" {
		providerName = cfg.DefaultProvider
	}

	if providerName == "" {
		fmt.Fprintf(os.Stderr, "%s No provider configured\n\n", fail("✖ Error:"))
		fmt.Fprintln(os.Stderr, aurora("You must configure a provider in your config file."))
		fmt.Println()
		fmt.Fprintln(os.Stderr, neon("Step 1 - Create Config File:"))
		fmt.Fprintln(os.Stderr, "  container-diet init-config")
		fmt.Println()
		fmt.Fprintln(os.Stderr, neon("Step 2 - Edit the config and set:"))
		fmt.Fprintln(os.Stderr, "  1. default_provider: \"openai\"  (or your preferred provider)")
		fmt.Fprintln(os.Stderr, "  2. Uncomment and configure the provider section")
		fmt.Fprintln(os.Stderr, "  3. Set api_key: \"${YOUR_API_KEY}\"")
		fmt.Fprintln(os.Stderr, "  4. Set default_model: \"model-name\"")
		fmt.Println()
		fmt.Fprintln(os.Stderr, neon("Step 3 - Or use --provider flag:"))
		fmt.Fprintln(os.Stderr, "  container-diet analyze <image> --provider openai")
		fmt.Println()
		return fmt.Errorf("no provider configured")
	}

	providerCfg, err := cfg.GetProvider(providerName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", fail("✖ Error:"), err)
		fmt.Fprintln(os.Stderr, warn("The provider is not configured in your config file."))
		fmt.Fprintln(os.Stderr, aurora("Run 'container-diet init-config' and uncomment the provider section."))
		return err
	}

	if providerCfg.APIKey == "" {
		fmt.Fprintf(os.Stderr, "%s Provider '%s' has no API key configured\n\n", fail("✖ Error:"), providerName)
		fmt.Fprintln(os.Stderr, aurora("Please set the api_key in your config file:"))
		fmt.Fprintln(os.Stderr, fmt.Sprintf("  %s:", providerName))
		fmt.Fprintln(os.Stderr, fmt.Sprintf("    api_key: \"${%s_API_KEY}\"", providerName))
		fmt.Println()
		fmt.Fprintln(os.Stderr, "Or set the environment variable:")
		fmt.Fprintln(os.Stderr, fmt.Sprintf("  export %s_API_KEY=your-key-here", providerName))
		fmt.Println()
		return fmt.Errorf("no API key for provider '%s'", providerName)
	}

	if providerCfg.DefaultModel == "" && model == "" {
		fmt.Fprintf(os.Stderr, "%s No model configured for provider '%s'\n\n", fail("✖ Error:"), providerName)
		fmt.Fprintln(os.Stderr, aurora("Please set the default_model in your config file:"))
		fmt.Fprintln(os.Stderr, fmt.Sprintf("  %s:", providerName))
		fmt.Fprintln(os.Stderr, "    default_model: \"your-preferred-model\"")
		fmt.Println()
		fmt.Fprintln(os.Stderr, "Or use the --model flag:")
		fmt.Fprintln(os.Stderr, "  container-diet analyze <image> --model gpt-4o")
		fmt.Println()
		return fmt.Errorf("no model configured for provider '%s'", providerName)
	}

	var analysis *analyzer.ImageAnalysis

	if len(args) > 0 {
		imageName := args[0]
		fmt.Printf("%s %s\n", neon("🔍 Scanning image:"), aurora(imageName))

		analysis, err = analyzer.AnalyzeImage(imageName, remote, pullMissing)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", fail("✖ Error analyzing image:"), err)
			return err
		}
	}

	// Read Dockerfile if provided
	var dockerfileContent string
	if dockerfilePath != "" {
		fmt.Printf("%s %s\n", neon("📄 Reading Dockerfile:"), aurora(dockerfilePath))
		content, err := os.ReadFile(dockerfilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", fail("✖ Error reading Dockerfile:"), err)
			return err
		}
		dockerfileContent = string(content)
	}

	if analysis == nil && dockerfileContent == "" {
		fmt.Println(warn("⚠ Please provide an image name or a Dockerfile path."))
		return fmt.Errorf("no image or Dockerfile provided")
	}

	// Print Image Analysis Summary
	printAnalysisSummary(analysis)

	// AI Analysis
	fmt.Printf("\n%s\n", neon("🤖 [AI ANALYSIS]"))

	// Create AI provider
	aiProvider, err := ai.NewProvider(providerCfg, providerName, cfg.Analysis.MaxTokens)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", fail("✖ Error initializing AI provider:"), err)
		return err
	}

	// Use provider's default model if none specified
	if model == "" {
		model = aiProvider.DefaultModel()
	}

	fmt.Printf("%s %s (%s)\n", aurora("🚢 Asking the Container Dietician for insights using"), aurora(providerName), aurora(model))
	result, err := aiProvider.AnalyzeReport(analysis, dockerfileContent, model, autoFix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", fail("✖ Error getting AI advice:"), err)
		return err
	}

	// JSON output mode — machine-readable, no colors
	if format == "json" {
		return outputJSON(analysis, result)
	}

	fmt.Printf("\n%s\n%s\n", neon(strings.Repeat("=", 64)), result.Advice)

	if autoFix && result.Fix != "" {
		fixPath := "Dockerfile.diet"
		if dockerfilePath != "" {
			fixPath = dockerfilePath + ".diet"
		}
		// Sanitize the output path to prevent directory traversal
		fixPath = filepath.Join(filepath.Dir(fixPath), filepath.Base(fixPath))

		fmt.Printf("\n%s %s\n", neon("🛠️  AUTO-FIX GENERATED:"), aurora(fixPath))
		err := os.WriteFile(fixPath, []byte(result.Fix), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", fail("✖ Error saving fixed Dockerfile:"), err)
			return err
		}
		fmt.Printf("%s Recommended changes saved to %s. Compare and apply them to slim down that image! 📉\n", success("✓"), success(fixPath))
	}

	fmt.Println(success("\n✓ Analysis complete"))
	return nil
}

// jsonOutput represents the machine-readable analysis result.
type jsonOutput struct {
	ImageName   string      `json:"image_name,omitempty"`
	TotalSize   int64       `json:"total_size_bytes,omitempty"`
	TotalSizeMB float64     `json:"total_size_mb,omitempty"`
	Layers      []layerJSON `json:"layers,omitempty"`
	Advice      string      `json:"advice"`
	Fix         string      `json:"fix,omitempty"`
}

type layerJSON struct {
	Index   int     `json:"index"`
	SizeMB  float64 `json:"size_mb"`
	Command string  `json:"command"`
}

func outputJSON(analysis *analyzer.ImageAnalysis, result *ai.AnalysisResponse) error {
	out := jsonOutput{
		Advice: result.Advice,
		Fix:    result.Fix,
	}

	if analysis != nil {
		out.ImageName = analysis.ImageName
		out.TotalSize = analysis.TotalSize
		out.TotalSizeMB = float64(analysis.TotalSize) / 1024 / 1024
		for i, l := range analysis.Layers {
			out.Layers = append(out.Layers, layerJSON{
				Index:   i + 1,
				SizeMB:  float64(l.Size) / 1024 / 1024,
				Command: l.Command,
			})
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", fail("✖ Error encoding JSON:"), err)
		return err
	}
	return nil
}

func init() {
	analyzeCmd.Flags().StringVarP(&format, "format", "f", "text", "Output format (text or json)")
	analyzeCmd.Flags().StringVar(&dockerfilePath, "dockerfile", "", "Path to Dockerfile for analysis")
	analyzeCmd.Flags().StringVar(&model, "model", "", "AI model to use for analysis (defaults to provider's default)")
	analyzeCmd.Flags().BoolVar(&remote, "remote", false, "Allow pulling image from remote registry if not found locally")
	analyzeCmd.Flags().BoolVar(&pullMissing, "pull-missing", false, "When local image is missing, pull it from remote and continue analysis")
	analyzeCmd.Flags().BoolVar(&autoFix, "auto-fix", false, "Generate an optimized version of the Dockerfile")
	analyzeCmd.Flags().StringVar(&providerName, "provider", "", "AI provider to use (openai, anthropic, openrouter, etc.)")
	analyzeCmd.Flags().StringVar(&configPath, "config", "", "Path to configuration file")
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
