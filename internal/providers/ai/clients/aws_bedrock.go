package clients

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// AWS Bedrock request/response structures for Anthropic Claude Messages API
type bedrockClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type bedrockClaudeRequest struct {
	AnthropicVersion string                 `json:"anthropic_version"`
	MaxTokens        int                    `json:"max_tokens"`
	Messages         []bedrockClaudeMessage `json:"messages"`
	Temperature      float64                `json:"temperature,omitempty"`
	TopP             float64                `json:"top_p,omitempty"`
	System           string                 `json:"system,omitempty"`
}

type bedrockClaudeResponse struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Role         string                 `json:"role"`
	Content      []bedrockClaudeContent `json:"content"`
	Model        string                 `json:"model"`
	StopReason   string                 `json:"stop_reason"`
	StopSequence string                 `json:"stop_sequence,omitempty"`
	Usage        bedrockClaudeUsage     `json:"usage"`
}

type bedrockClaudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type bedrockClaudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type bedrockClaudeStreamChunk struct {
	Type         string                 `json:"type"`
	Index        int                    `json:"index,omitempty"`
	Delta        *bedrockClaudeDelta    `json:"delta,omitempty"`
	ContentBlock *bedrockClaudeContent  `json:"content_block,omitempty"`
	Message      *bedrockClaudeResponse `json:"message,omitempty"`
	Usage        *bedrockClaudeUsage    `json:"usage,omitempty"`
}

type bedrockClaudeDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AWS Bedrock Titan embedding request/response
type bedrockTitanEmbeddingRequest struct {
	InputText string `json:"inputText"`
}

type bedrockTitanEmbeddingResponse struct {
	Embedding           []float32 `json:"embedding"`
	InputTextTokenCount int       `json:"inputTextTokenCount"`
}

// AWS Bedrock Cohere embedding request/response
type bedrockCohereEmbeddingRequest struct {
	Texts     []string `json:"texts"`
	InputType string   `json:"input_type"`
}

type bedrockCohereEmbeddingResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
	ID         string      `json:"id"`
	Texts      []string    `json:"texts"`
}

// AWSBedrockClient implements domain.LLMProvider for AWS Bedrock
type AWSBedrockClient struct {
	httpClient   *http.Client
	region       string
	accessKey    string
	secretKey    string
	sessionToken string
	model        string
	providerType domain.ProviderType
	config       *config.ProviderConfig
	timeout      time.Duration
	maxRetries   int
}

// NewAWSBedrockClient creates a new AWS Bedrock provider
func NewAWSBedrockClient(providerType domain.ProviderType, cfg *config.ProviderConfig) (domain.LLMProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	// Get AWS credentials from config
	region := cfg.AWSRegion
	if region == "" {
		region = "us-east-1" // Default region
	}

	accessKey := cfg.AWSAccessKeyID
	if accessKey == "" {
		return nil, fmt.Errorf("AWS access key ID is required")
	}

	secretKey := cfg.AWSSecretAccessKey
	if secretKey == "" {
		return nil, fmt.Errorf("AWS secret access key is required")
	}

	sessionToken := cfg.AWSSessionToken // Optional

	model := cfg.DefaultModel
	if model == "" {
		return nil, fmt.Errorf("model ID is required for Bedrock")
	}

	timeout := 45 * time.Second
	if cfg.TimeoutSeconds > 0 {
		timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}

	maxRetries := 3
	if cfg.MaxRetries >= 0 {
		maxRetries = cfg.MaxRetries
	}

	logging.Info("Creating AWS Bedrock client for region %s, model %s", region, model)

	return &AWSBedrockClient{
		httpClient:   &http.Client{Timeout: timeout},
		region:       region,
		accessKey:    accessKey,
		secretKey:    secretKey,
		sessionToken: sessionToken,
		model:        model,
		providerType: providerType,
		config:       cfg,
		timeout:      timeout,
		maxRetries:   maxRetries,
	}, nil
}

// CreateCompletion implements domain.LLMProvider
func (c *AWSBedrockClient) CreateCompletion(ctx context.Context, req *domain.CompletionRequest) (*domain.CompletionResponse, error) {
	// Convert messages to Claude Messages API format
	messages := c.convertToClaudeMessages(req.Messages)

	// Create Bedrock request (Claude Messages API format)
	bedrockReq := bedrockClaudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        2048,
		Messages:         messages,
		Temperature:      0.7,
	}

	// Add system prompt if provided
	if req.SystemPrompt != "" {
		bedrockReq.System = req.SystemPrompt
	}

	payloadBytes, err := json.Marshal(bedrockReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/invoke", c.region, c.model)

	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			logging.Warn("Retrying Bedrock request (attempt %d/%d)", retry, c.maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payloadBytes))
		if err != nil {
			lastErr = err
			continue
		}

		// Sign request with AWS SigV4
		if err := c.signRequest(httpReq, payloadBytes); err != nil {
			lastErr = fmt.Errorf("failed to sign request: %w", err)
			continue
		}

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("Bedrock API error (%s): %s", resp.Status, string(body))
			continue
		}

		var bedrockResp bedrockClaudeResponse
		if err := json.Unmarshal(body, &bedrockResp); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %w", err)
			continue
		}

		// Extract text from content blocks
		var responseText string
		for _, content := range bedrockResp.Content {
			if content.Type == "text" {
				responseText += content.Text
			}
		}

		return &domain.CompletionResponse{
			Response:  responseText,
			ToolCalls: nil, // Tool calling requires different format
		}, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// StreamCompletion implements domain.LLMProvider
func (c *AWSBedrockClient) StreamCompletion(ctx context.Context, req *domain.CompletionRequest, writer io.Writer) (*domain.CompletionResponse, error) {
	// Convert messages to Claude Messages API format
	messages := c.convertToClaudeMessages(req.Messages)

	bedrockReq := bedrockClaudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        2048,
		Messages:         messages,
		Temperature:      0.7,
	}

	// Add system prompt if provided
	if req.SystemPrompt != "" {
		bedrockReq.System = req.SystemPrompt
	}

	payloadBytes, err := json.Marshal(bedrockReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Streaming endpoint
	endpoint := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/invoke-with-response-stream", c.region, c.model)

	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			logging.Warn("Retrying Bedrock streaming request (attempt %d/%d)", retry, c.maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payloadBytes))
		if err != nil {
			lastErr = err
			continue
		}

		if err := c.signRequest(httpReq, payloadBytes); err != nil {
			lastErr = fmt.Errorf("failed to sign request: %w", err)
			continue
		}

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("Bedrock API error (%s): %s", resp.Status, string(body))
			continue
		}

		// Process streaming response
		fullContent, err := c.processBedrockStream(resp, writer)
		if err != nil {
			lastErr = err
			continue
		}

		return &domain.CompletionResponse{
			Response:  fullContent,
			ToolCalls: nil,
		}, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// processBedrockStream processes AWS Bedrock event stream (Messages API format)
func (c *AWSBedrockClient) processBedrockStream(resp *http.Response, writer io.Writer) (string, error) {
	defer resp.Body.Close()

	var fullContent string
	decoder := json.NewDecoder(resp.Body)

	for {
		var chunk bedrockClaudeStreamChunk
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			logging.Warn("Stream decode error: %v", err)
			continue
		}

		// Handle different chunk types
		if chunk.Type == "content_block_delta" && chunk.Delta != nil && chunk.Delta.Text != "" {
			fullContent += chunk.Delta.Text
			if writer != nil {
				writer.Write([]byte(chunk.Delta.Text))
			}
		}

		// Check for message_stop
		if chunk.Type == "message_stop" {
			break
		}
	}

	return fullContent, nil
}

// CreateEmbeddings implements domain.LLMProvider
func (c *AWSBedrockClient) CreateEmbeddings(ctx context.Context, req *domain.EmbeddingRequest) (*domain.EmbeddingResponse, error) {
	if len(req.Input) == 0 {
		return nil, fmt.Errorf("input is required for embeddings")
	}

	// Determine embedding model from config's default_model
	embeddingModel := ""
	if c.config != nil && c.config.DefaultModel != "" {
		embeddingModel = c.config.DefaultModel
		logging.Debug("Using default_model from config: %s", embeddingModel)
	}

	// Fallback to Cohere if no default configured
	if embeddingModel == "" {
		embeddingModel = "cohere.embed-english-v3"
		logging.Debug("No default model in config, using fallback: %s", embeddingModel)
	}

	// Override with requested model if specified
	if req.Model != "" {
		embeddingModel = req.Model
		logging.Debug("Using model from request: %s", embeddingModel)
	}

	logging.Info("Creating embeddings with model: %s for %d inputs", embeddingModel, len(req.Input))

	// Route to appropriate implementation based on model
	if strings.Contains(embeddingModel, "cohere") {
		return c.createCohereEmbeddings(ctx, req.Input, embeddingModel)
	} else if strings.Contains(embeddingModel, "titan") {
		return c.createTitanEmbeddings(ctx, req.Input, embeddingModel)
	}

	return nil, fmt.Errorf("unsupported embedding model: %s", embeddingModel)
}

// createCohereEmbeddings creates embeddings using Cohere models (batch API)
func (c *AWSBedrockClient) createCohereEmbeddings(ctx context.Context, inputs []string, model string) (*domain.EmbeddingResponse, error) {
	// Cohere supports batch processing
	cohereReq := bedrockCohereEmbeddingRequest{
		Texts:     inputs,
		InputType: "search_document",
	}

	payloadBytes, err := json.Marshal(cohereReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	endpoint := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/invoke", c.region, model)
	logging.Debug("Cohere embedding endpoint: %s", endpoint)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	if err := c.signRequest(httpReq, payloadBytes); err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bedrock embedding API error (%s): %s", resp.Status, string(body))
	}

	var cohereResp bedrockCohereEmbeddingResponse
	if err := json.Unmarshal(body, &cohereResp); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	// Convert to standard format
	embeddings := make([]domain.Embedding, len(cohereResp.Embeddings))
	for i, embedding := range cohereResp.Embeddings {
		embeddings[i] = domain.Embedding{
			Object:    "embedding",
			Index:     i,
			Embedding: embedding,
		}
	}

	logging.Info("Successfully created %d Cohere embeddings", len(embeddings))

	return &domain.EmbeddingResponse{
		Object: "list",
		Data:   embeddings,
		Model:  model,
		Usage: domain.Usage{
			PromptTokens: 0,
			TotalTokens:  0,
		},
	}, nil
}

// createTitanEmbeddings creates embeddings using Titan models (single-input API)
func (c *AWSBedrockClient) createTitanEmbeddings(ctx context.Context, inputs []string, model string) (*domain.EmbeddingResponse, error) {
	var embeddings []domain.Embedding

	// Titan processes one input at a time
	for i, text := range inputs {
		titanReq := bedrockTitanEmbeddingRequest{
			InputText: text,
		}

		payloadBytes, err := json.Marshal(titanReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
		}

		endpoint := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/invoke", c.region, model)
		logging.Debug("Titan embedding request %d/%d: Endpoint: %s", i+1, len(inputs), endpoint)

		httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payloadBytes))
		if err != nil {
			return nil, err
		}

		if err := c.signRequest(httpReq, payloadBytes); err != nil {
			return nil, fmt.Errorf("failed to sign request: %w", err)
		}

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Bedrock embedding API error (%s): %s", resp.Status, string(body))
		}

		var titanResp bedrockTitanEmbeddingResponse
		if err := json.Unmarshal(body, &titanResp); err != nil {
			return nil, fmt.Errorf("failed to parse embedding response: %w", err)
		}

		embeddings = append(embeddings, domain.Embedding{
			Object:    "embedding",
			Index:     i,
			Embedding: titanResp.Embedding,
		})
	}

	logging.Info("Successfully created %d Titan embeddings", len(embeddings))

	return &domain.EmbeddingResponse{
		Object: "list",
		Data:   embeddings,
		Model:  model,
		Usage: domain.Usage{
			PromptTokens: 0,
			TotalTokens:  0,
		},
	}, nil
}

// GetSupportedEmbeddingModels implements domain.LLMProvider
func (c *AWSBedrockClient) GetSupportedEmbeddingModels() []string {
	// Return models from config if available
	if c.config != nil && len(c.config.EmbeddingModels) > 0 {
		models := make([]string, 0, len(c.config.EmbeddingModels))
		for modelName := range c.config.EmbeddingModels {
			models = append(models, modelName)
		}
		return models
	}

	// Fallback to default models with correct IDs
	return []string{
		"cohere.embed-english-v3",
		"cohere.embed-multilingual-v3",
		"amazon.titan-embed-text-v2:0",
		"amazon.titan-embed-g1-text-02",
	}
}

// GetMaxEmbeddingTokens implements domain.LLMProvider
func (c *AWSBedrockClient) GetMaxEmbeddingTokens(model string) int {
	// Try to get from config first
	if c.config != nil && c.config.EmbeddingModels != nil {
		if modelConfig, exists := c.config.EmbeddingModels[model]; exists && modelConfig.MaxTokens > 0 {
			return modelConfig.MaxTokens
		}
	}

	// Fallback defaults by model type
	switch {
	case strings.Contains(model, "titan"):
		return 8192
	case strings.Contains(model, "cohere"):
		return 512
	default:
		return 8192
	}
}

// GetProviderType implements domain.LLMProvider
func (c *AWSBedrockClient) GetProviderType() domain.ProviderType {
	return c.providerType
}

// GetInterfaceType implements domain.LLMProvider
func (c *AWSBedrockClient) GetInterfaceType() config.InterfaceType {
	return config.AWSBedrock
}

// ValidateConfig implements domain.LLMProvider
func (c *AWSBedrockClient) ValidateConfig() error {
	if c.accessKey == "" {
		return fmt.Errorf("AWS access key is required")
	}
	if c.secretKey == "" {
		return fmt.Errorf("AWS secret key is required")
	}
	if c.model == "" {
		return fmt.Errorf("model ID is required")
	}
	return nil
}

// Close implements domain.LLMProvider
func (c *AWSBedrockClient) Close() error {
	return nil
}

// convertToClaudeMessages converts domain messages to Claude Messages API format
func (c *AWSBedrockClient) convertToClaudeMessages(messages []domain.Message) []bedrockClaudeMessage {
	var claudeMessages []bedrockClaudeMessage

	for _, msg := range messages {
		// Skip system messages (they go in the system field)
		if msg.Role == "system" {
			continue
		}

		claudeMessages = append(claudeMessages, bedrockClaudeMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	return claudeMessages
}

// signRequest signs AWS request with SigV4 (lightweight implementation)
func (c *AWSBedrockClient) signRequest(req *http.Request, payload []byte) error {
	now := time.Now().UTC()
	dateStamp := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")

	service := "bedrock"

	// Set required headers BEFORE using them
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Amz-Date", amzDate)

	// Include session token if present
	if c.sessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", c.sessionToken)
	}

	// Build canonical headers and signed headers list (must be in alphabetical order)
	var canonicalHeadersList []string
	var signedHeadersList []string

	canonicalHeadersList = append(canonicalHeadersList, fmt.Sprintf("content-type:%s", req.Header.Get("Content-Type")))
	signedHeadersList = append(signedHeadersList, "content-type")

	canonicalHeadersList = append(canonicalHeadersList, fmt.Sprintf("host:%s", req.Host))
	signedHeadersList = append(signedHeadersList, "host")

	canonicalHeadersList = append(canonicalHeadersList, fmt.Sprintf("x-amz-date:%s", amzDate))
	signedHeadersList = append(signedHeadersList, "x-amz-date")

	// Include session token in canonical headers if present (alphabetically after x-amz-date)
	if c.sessionToken != "" {
		canonicalHeadersList = append(canonicalHeadersList, fmt.Sprintf("x-amz-security-token:%s", c.sessionToken))
		signedHeadersList = append(signedHeadersList, "x-amz-security-token")
	}

	// Join canonical headers WITHOUT trailing newline (we'll add it in the canonical request)
	canonicalHeaders := strings.Join(canonicalHeadersList, "\n")
	signedHeaders := strings.Join(signedHeadersList, ";")

	// Create canonical request components
	// AWS SigV4 requires RFC 3986 URI encoding (which encodes colons)
	canonicalURI := c.uriEncode(req.URL.Path)
	canonicalQueryString := "" // Empty for this request
	if req.URL.RawQuery != "" {
		canonicalQueryString = req.URL.RawQuery
	}

	payloadHash := hashSHA256(payload)

	// Build canonical request with exact format AWS expects
	// Format: METHOD\nURI\nQUERY_STRING\nHEADERS\n\nSIGNED_HEADERS\nPAYLOAD_HASH
	canonicalRequest := req.Method + "\n" +
		canonicalURI + "\n" +
		canonicalQueryString + "\n" +
		canonicalHeaders + "\n" +
		"\n" +
		signedHeaders + "\n" +
		payloadHash

	// Debug log the canonical request
	logging.Debug("Canonical Request:\n%s", canonicalRequest)
	logging.Debug("Canonical Request Hash: %s", hashSHA256([]byte(canonicalRequest)))

	// Create string to sign
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, c.region, service)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s",
		amzDate,
		credentialScope,
		hashSHA256([]byte(canonicalRequest)))

	logging.Debug("String to Sign:\n%s", stringToSign)

	// Calculate signature
	signature := c.calculateSignature(dateStamp, service, stringToSign)

	logging.Debug("Signature: %s", signature)

	// Add authorization header
	authorization := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		c.accessKey,
		credentialScope,
		signedHeaders,
		signature)

	req.Header.Set("Authorization", authorization)

	return nil
}

// calculateSignature calculates AWS SigV4 signature
func (c *AWSBedrockClient) calculateSignature(dateStamp, service, stringToSign string) string {
	kDate := hmacSHA256([]byte("AWS4"+c.secretKey), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(c.region))
	kService := hmacSHA256(kRegion, []byte(service))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	signature := hmacSHA256(kSigning, []byte(stringToSign))
	return hex.EncodeToString(signature)
}

// hashSHA256 calculates SHA256 hash
func hashSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// hmacSHA256 calculates HMAC-SHA256
func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// uriEncode encodes a URI path according to RFC 3986 (required for AWS SigV4)
// Unlike Go's url.PathEscape, this encodes colons which AWS requires
func (c *AWSBedrockClient) uriEncode(path string) string {
	var encoded strings.Builder
	for i := 0; i < len(path); i++ {
		ch := path[i]
		// Unreserved characters per RFC 3986: A-Z a-z 0-9 - _ . ~
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '_' || ch == '.' || ch == '~' || ch == '/' {
			encoded.WriteByte(ch)
		} else {
			// Percent-encode everything else
			encoded.WriteString(fmt.Sprintf("%%%02X", ch))
		}
	}
	return encoded.String()
}
