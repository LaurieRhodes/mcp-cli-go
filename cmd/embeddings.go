package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai"
	"github.com/LaurieRhodes/mcp-cli-go/internal/services/embeddings"
	"github.com/spf13/cobra"
)

// Embeddings command flags
var (
	embeddingProvider     string
	embeddingModel        string
	embeddingInputFile    string // Renamed to avoid conflict
	embeddingOutputFile   string // Renamed to avoid conflict
	embeddingOutputFormat string
	chunkStrategy         string
	maxChunkSize          int
	chunkOverlap          int
	includeMeta           bool
	encodingFormat        string
	dimensions            int
	showModels            bool
	showStrategies        bool
)

// EmbeddingsCmd represents the embeddings command
var EmbeddingsCmd = &cobra.Command{
	Use:   "embeddings [text]",
	Short: "Generate vector embeddings from text input",
	Long: `Generate vector embeddings from text input using various embedding models.

Supports multiple input sources:
- Direct text argument
- Standard input (stdin)
- File input

Text is automatically chunked using configurable strategies to handle
large inputs and optimize embedding quality.

Examples:
  # Basic usage with stdin
  echo "Your text here" | mcp-cli embeddings
  
  # File input with specific model
  mcp-cli embeddings --input-file document.txt --model text-embedding-3-large
  
  # Advanced chunking and output
  mcp-cli embeddings --chunk-strategy sentence --max-chunk-size 512 --output-format json --overlap 50
  
  # Direct text input
  mcp-cli embeddings "Analyze this specific text"
  
  # Show available models and strategies
  mcp-cli embeddings --show-models
  mcp-cli embeddings --show-strategies`,
	RunE: executeEmbeddings,
}

func init() {
	// Provider and model flags
	EmbeddingsCmd.Flags().StringVar(&embeddingProvider, "provider", "", "AI provider to use (openai, deepseek, openrouter)")
	EmbeddingsCmd.Flags().StringVar(&embeddingModel, "model", "", "Embedding model to use")

	// Input/output flags
	EmbeddingsCmd.Flags().StringVar(&embeddingInputFile, "input-file", "", "Input file path")
	EmbeddingsCmd.Flags().StringVar(&embeddingOutputFile, "output-file", "", "Output file path (default: stdout)")
	EmbeddingsCmd.Flags().StringVar(&embeddingOutputFormat, "output-format", "json", "Output format (json, csv, compact)")

	// Chunking flags
	EmbeddingsCmd.Flags().StringVar(&chunkStrategy, "chunk-strategy", "sentence", "Chunking strategy (sentence, paragraph, fixed)")
	EmbeddingsCmd.Flags().IntVar(&maxChunkSize, "max-chunk-size", 512, "Maximum chunk size in tokens")
	EmbeddingsCmd.Flags().IntVar(&chunkOverlap, "overlap", 0, "Overlap between chunks in tokens")

	// Embedding flags
	EmbeddingsCmd.Flags().StringVar(&encodingFormat, "encoding-format", "float", "Encoding format (float, base64)")
	EmbeddingsCmd.Flags().IntVar(&dimensions, "dimensions", 0, "Number of dimensions (for supported models)")
	EmbeddingsCmd.Flags().BoolVar(&includeMeta, "include-metadata", true, "Include chunk and model metadata")

	// Info flags
	EmbeddingsCmd.Flags().BoolVar(&showModels, "show-models", false, "Show available embedding models")
	EmbeddingsCmd.Flags().BoolVar(&showStrategies, "show-strategies", false, "Show available chunking strategies")
}

func executeEmbeddings(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Initialize services
	configService := config.NewService()
	providerFactory := ai.NewProviderFactory()
	embeddingService := embeddings.NewService(configService, providerFactory)

	// Load configuration using configFile from root command
	_, err := configService.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Handle info flags
	if showModels {
		return showAvailableModels(configService, providerFactory)
	}

	if showStrategies {
		return showAvailableStrategies(embeddingService)
	}

	// Get input text
	var inputText string

	if embeddingInputFile != "" {
		// Read from file
		data, err := os.ReadFile(embeddingInputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
		inputText = string(data)
		logging.Info("Read %d characters from file: %s", len(inputText), embeddingInputFile)
	} else if len(args) > 0 {
		// Use command line argument
		inputText = strings.Join(args, " ")
	} else {
		// Read from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Data is being piped
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
			inputText = string(data)
			logging.Info("Read %d characters from stdin", len(inputText))
		} else {
			return fmt.Errorf("no input provided. Use --input-file, provide text as argument, or pipe to stdin")
		}
	}

	if strings.TrimSpace(inputText) == "" {
		return fmt.Errorf("input text is empty")
	}

	// Create embedding request
	req := &domain.EmbeddingJobRequest{
		Input:          inputText,
		Provider:       embeddingProvider,
		Model:          embeddingModel,
		ChunkStrategy:  domain.ChunkingType(chunkStrategy),
		MaxChunkSize:   maxChunkSize,
		ChunkOverlap:   chunkOverlap,
		EncodingFormat: encodingFormat,
		Dimensions:     dimensions,
		Metadata: map[string]interface{}{
			"cli_version": "1.0.0",
			"source":      getInputSource(),
		},
	}

	// Generate embeddings
	logging.Info("Generating embeddings...")
	job, err := embeddingService.GenerateEmbeddings(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Format and output results
	output, err := formatOutput(job, embeddingOutputFormat, includeMeta)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	// Write output
	if embeddingOutputFile != "" {
		err = os.WriteFile(embeddingOutputFile, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		logging.Info("Output written to: %s", embeddingOutputFile)
	} else {
		fmt.Print(output)
	}

	// Log summary
	logging.Info("Embedding generation completed: %d chunks, %d embeddings",
		len(job.Chunks), len(job.Embeddings))

	return nil
}

func showAvailableModels(configService domain.ConfigurationService, providerFactory domain.ProviderFactory) error {
	fmt.Println("Available Embedding Models:")
	fmt.Println("==========================")

	// Get default provider
	defaultProviderName, defaultConfig, interfaceType, err := configService.GetDefaultProvider()
	if err != nil {
		logging.Warn("Could not get default provider: %v", err)
		return nil
	}

	// Create provider instance to get supported models
	providerType := domain.ProviderType(defaultProviderName)

	provider, err := providerFactory.CreateProvider(providerType, defaultConfig, interfaceType)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}
	defer provider.Close()

	models := provider.GetSupportedEmbeddingModels()
	for _, model := range models {
		maxTokens := provider.GetMaxEmbeddingTokens(model)
		fmt.Printf("  %-30s (max tokens: %d)\n", model, maxTokens)
	}

	return nil
}

func showAvailableStrategies(embeddingService domain.EmbeddingService) error {
	fmt.Println("Available Chunking Strategies:")
	fmt.Println("=============================")

	strategies := embeddingService.GetAvailableChunkingStrategies()
	for _, strategy := range strategies {
		fmt.Printf("  %-15s - %s", strategy, getStrategyDescription(strategy))
	}
	return nil
}

func getStrategyDescription(strategy domain.ChunkingType) string {
	switch strategy {
	case domain.ChunkingSentence:
		return "Splits text at sentence boundaries while preserving semantic meaning"
	case domain.ChunkingParagraph:
		return "Splits text at paragraph boundaries to preserve document structure"
	case domain.ChunkingFixed:
		return "Splits text into fixed-size chunks with configurable overlap"
	default:
		return "Unknown strategy"
	}
}

func getInputSource() string {
	if embeddingInputFile != "" {
		return "file"
	}
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return "stdin"
	}
	return "argument"
}

func formatOutput(job *domain.EmbeddingJob, format string, includeMeta bool) (string, error) {
	switch format {
	case "json":
		return formatJSON(job, includeMeta)
	case "csv":
		return formatCSV(job, includeMeta)
	case "compact":
		return formatCompact(job)
	default:
		return "", fmt.Errorf("unsupported output format: %s", format)
	}
}

func formatJSON(job *domain.EmbeddingJob, includeMeta bool) (string, error) {
	if includeMeta {
		// Full job with metadata
		data, err := json.MarshalIndent(job, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	} else {
		// Minimal format - just vectors
		vectors := make([][]float32, len(job.Embeddings))
		for i, embedding := range job.Embeddings {
			vectors[i] = embedding.Vector
		}

		minimal := map[string]interface{}{
			"model":   job.Model,
			"vectors": vectors,
		}

		data, err := json.MarshalIndent(minimal, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
}

func formatCSV(job *domain.EmbeddingJob, includeMeta bool) (string, error) {
	var lines []string

	if includeMeta {
		// CSV with metadata
		lines = append(lines, "chunk_index,text,vector_json,start_pos,end_pos,token_count")

		for _, embedding := range job.Embeddings {
			vectorJSON, _ := json.Marshal(embedding.Vector)
			text := strings.ReplaceAll(embedding.Chunk.Text, "\"", "\"\"") // Escape quotes
			line := fmt.Sprintf("%d,\"%s\",\"%s\",%d,%d,%d",
				embedding.Chunk.Index,
				text,
				string(vectorJSON),
				embedding.Chunk.StartPos,
				embedding.Chunk.EndPos,
				embedding.Chunk.TokenCount,
			)
			lines = append(lines, line)
		}
	} else {
		// CSV with just vectors
		lines = append(lines, "chunk_index,vector_json")

		for i, embedding := range job.Embeddings {
			vectorJSON, _ := json.Marshal(embedding.Vector)
			line := fmt.Sprintf("%d,\"%s\"", i, string(vectorJSON))
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "") + "", nil
}

func formatCompact(job *domain.EmbeddingJob) (string, error) {
	// Very minimal JSON format
	vectors := make([][]float32, len(job.Embeddings))
	for i, embedding := range job.Embeddings {
		vectors[i] = embedding.Vector
	}

	compact := map[string]interface{}{
		"model":   job.Model,
		"vectors": vectors,
	}

	data, err := json.Marshal(compact)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
