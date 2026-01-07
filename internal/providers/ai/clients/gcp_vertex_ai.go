package clients

import (
	"bufio"
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// GCP Service Account structure
type gcpServiceAccount struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

// JWT claims for OAuth2
type jwtClaims struct {
	Iss   string `json:"iss"`
	Scope string `json:"scope"`
	Aud   string `json:"aud"`
	Exp   int64  `json:"exp"`
	Iat   int64  `json:"iat"`
}

// OAuth2 token response
type oauth2TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Vertex AI Gemini request/response structures
type vertexGeminiRequest struct {
	Contents         []vertexContent       `json:"contents"`
	GenerationConfig *vertexGenConfig      `json:"generationConfig,omitempty"`
	SafetySettings   []vertexSafetySetting `json:"safetySettings,omitempty"`
}

type vertexContent struct {
	Role  string       `json:"role"`
	Parts []vertexPart `json:"parts"`
}

type vertexPart struct {
	Text string `json:"text"`
}

type vertexGenConfig struct {
	Temperature     float64  `json:"temperature,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	TopK            int      `json:"topK,omitempty"`
	ResponseMimeType string  `json:"responseMimeType,omitempty"` // text/plain disables code execution
}

type vertexSafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type vertexGeminiResponse struct {
	Candidates []vertexCandidate `json:"candidates"`
}

type vertexCandidate struct {
	Content       vertexContent `json:"content"`
	FinishReason  string        `json:"finishReason"`
	SafetyRatings []interface{} `json:"safetyRatings"`
}

// Vertex AI embedding structures
type vertexEmbeddingRequest struct {
	Instances []vertexEmbeddingInstance `json:"instances"`
}

type vertexEmbeddingInstance struct {
	Content string `json:"content"`
}

type vertexEmbeddingResponse struct {
	Predictions []vertexEmbeddingPrediction `json:"predictions"`
}

type vertexEmbeddingPrediction struct {
	Embeddings struct {
		Values []float32 `json:"values"`
	} `json:"embeddings"`
}

// GCPVertexAIClient implements domain.LLMProvider for Google Cloud Vertex AI
type GCPVertexAIClient struct {
	httpClient      *http.Client
	projectID       string
	location        string
	model           string
	accessToken     string
	tokenExpiry     time.Time
	serviceAccount  *gcpServiceAccount
	providerType    domain.ProviderType
	config          *config.ProviderConfig
	timeout         time.Duration
	maxRetries      int
}

// NewGCPVertexAIClient creates a new GCP Vertex AI provider
func NewGCPVertexAIClient(providerType domain.ProviderType, cfg *config.ProviderConfig) (domain.LLMProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	// Get GCP-specific config
	projectID := cfg.ProjectID
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	location := cfg.Location
	if location == "" {
		location = "us-central1"
	}

	credentialsPath := cfg.CredentialsPath // Optional - can use ADC

	model := cfg.DefaultModel
	if model == "" {
		model = "gemini-pro"
	}

	timeout := 45 * time.Second
	if cfg.TimeoutSeconds > 0 {
		timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}

	maxRetries := 3
	if cfg.MaxRetries >= 0 {
		maxRetries = cfg.MaxRetries
	}

	client := &GCPVertexAIClient{
		httpClient:   &http.Client{Timeout: timeout},
		projectID:    projectID,
		location:     location,
		model:        model,
		providerType: providerType,
		config:       cfg,
		timeout:      timeout,
		maxRetries:   maxRetries,
	}

	// Load service account if credentials path provided
	if credentialsPath != "" {
		if err := client.loadServiceAccount(credentialsPath); err != nil {
			return nil, fmt.Errorf("failed to load service account: %w", err)
		}
		logging.Info("Loaded service account for GCP Vertex AI")
	} else {
		logging.Warn("No service account provided - authentication will fail. Set credentials_path in config")
	}

	logging.Info("Creating GCP Vertex AI client for project %s, location %s, model %s", projectID, location, model)

	return client, nil
}

// CreateCompletion implements domain.LLMProvider
func (c *GCPVertexAIClient) CreateCompletion(ctx context.Context, req *domain.CompletionRequest) (*domain.CompletionResponse, error) {
	// Convert messages to Vertex AI format
	contents := c.convertToVertexContents(req.Messages, req.SystemPrompt)
	
	vertexReq := vertexGeminiRequest{
		Contents: contents,
		GenerationConfig: &vertexGenConfig{
			Temperature:      0.7,
			MaxOutputTokens:  2048,
			ResponseMimeType: "text/plain", // Disable code execution
		},
	}

	payloadBytes, err := json.Marshal(vertexReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Endpoint: https://{location}-aiplatform.googleapis.com/v1/projects/{project}/locations/{location}/publishers/google/models/{model}:generateContent
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
		c.location, c.projectID, c.location, c.model)

	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			logging.Warn("Retrying Vertex AI request (attempt %d/%d)", retry, c.maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		// Ensure we have a valid access token
		if err := c.ensureAccessToken(); err != nil {
			lastErr = fmt.Errorf("failed to get access token: %w", err)
			continue
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payloadBytes))
		if err != nil {
			lastErr = err
			continue
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)

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
			lastErr = fmt.Errorf("Vertex AI API error (%s): %s", resp.Status, string(body))
			continue
		}

		var vertexResp vertexGeminiResponse
		if err := json.Unmarshal(body, &vertexResp); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %w", err)
			continue
		}

		if len(vertexResp.Candidates) == 0 || len(vertexResp.Candidates[0].Content.Parts) == 0 {
			lastErr = fmt.Errorf("no response content")
			continue
		}

		text := vertexResp.Candidates[0].Content.Parts[0].Text

		return &domain.CompletionResponse{
			Response:  text,
			ToolCalls: nil, // Function calling requires different format
		}, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// StreamCompletion implements domain.LLMProvider  
func (c *GCPVertexAIClient) StreamCompletion(ctx context.Context, req *domain.CompletionRequest, writer io.Writer) (*domain.CompletionResponse, error) {
	contents := c.convertToVertexContents(req.Messages, req.SystemPrompt)
	
	vertexReq := vertexGeminiRequest{
		Contents: contents,
		GenerationConfig: &vertexGenConfig{
			Temperature:      0.7,
			MaxOutputTokens:  2048,
			ResponseMimeType: "text/plain", // Disable code execution
		},
	}

	payloadBytes, err := json.Marshal(vertexReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Streaming endpoint
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:streamGenerateContent",
		c.location, c.projectID, c.location, c.model)

	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			logging.Warn("Retrying Vertex AI streaming (attempt %d/%d)", retry, c.maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		if err := c.ensureAccessToken(); err != nil {
			lastErr = err
			continue
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payloadBytes))
		if err != nil {
			lastErr = err
			continue
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("Vertex AI API error (%s): %s", resp.Status, string(body))
			continue
		}

		fullContent, err := c.processVertexStream(resp, writer)
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

// processVertexStream processes Vertex AI streaming response
func (c *GCPVertexAIClient) processVertexStream(resp *http.Response, writer io.Writer) (string, error) {
	defer resp.Body.Close()

	var fullContent string
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()
		
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Vertex AI uses newline-delimited JSON
		var streamResp vertexGeminiResponse
		if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
			logging.Warn("Failed to parse stream chunk: %v", err)
			continue
		}

		if len(streamResp.Candidates) > 0 && len(streamResp.Candidates[0].Content.Parts) > 0 {
			text := streamResp.Candidates[0].Content.Parts[0].Text
			if text != "" {
				fullContent += text
				if writer != nil {
					writer.Write([]byte(text))
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fullContent, fmt.Errorf("streaming error: %w", err)
	}

	return fullContent, nil
}

// CreateEmbeddings implements domain.LLMProvider
func (c *GCPVertexAIClient) CreateEmbeddings(ctx context.Context, req *domain.EmbeddingRequest) (*domain.EmbeddingResponse, error) {
	if len(req.Input) == 0 {
		return nil, fmt.Errorf("input is required for embeddings")
	}

	// Use gecko embedding model
	embeddingModel := "textembedding-gecko@003"
	if req.Model != "" {
		embeddingModel = req.Model
	}

	// Create instances for each input
	instances := make([]vertexEmbeddingInstance, len(req.Input))
	for i, text := range req.Input {
		instances[i] = vertexEmbeddingInstance{
			Content: text,
		}
	}

	vertexReq := vertexEmbeddingRequest{
		Instances: instances,
	}

	payloadBytes, err := json.Marshal(vertexReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	// Embedding endpoint
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict",
		c.location, c.projectID, c.location, embeddingModel)

	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		if err := c.ensureAccessToken(); err != nil {
			lastErr = err
			continue
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payloadBytes))
		if err != nil {
			lastErr = err
			continue
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)

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
			lastErr = fmt.Errorf("Vertex AI embedding API error (%s): %s", resp.Status, string(body))
			continue
		}

		var vertexResp vertexEmbeddingResponse
		if err := json.Unmarshal(body, &vertexResp); err != nil {
			lastErr = fmt.Errorf("failed to parse embedding response: %w", err)
			continue
		}

		// Convert to domain format
		embeddings := make([]domain.Embedding, len(vertexResp.Predictions))
		for i, pred := range vertexResp.Predictions {
			embeddings[i] = domain.Embedding{
				Object:    "embedding",
				Index:     i,
				Embedding: pred.Embeddings.Values,
			}
		}

		return &domain.EmbeddingResponse{
			Object: "list",
			Data:   embeddings,
			Model:  embeddingModel,
			Usage: domain.Usage{
				PromptTokens: 0, // Vertex AI doesn't return token counts
				TotalTokens:  0,
			},
		}, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// GetSupportedEmbeddingModels implements domain.LLMProvider
func (c *GCPVertexAIClient) GetSupportedEmbeddingModels() []string {
	return []string{"textembedding-gecko@003"}
}

// GetMaxEmbeddingTokens implements domain.LLMProvider
func (c *GCPVertexAIClient) GetMaxEmbeddingTokens(model string) int {
	return 3072
}

// GetProviderType implements domain.LLMProvider
func (c *GCPVertexAIClient) GetProviderType() domain.ProviderType {
	return c.providerType
}

// GetInterfaceType implements domain.LLMProvider
func (c *GCPVertexAIClient) GetInterfaceType() config.InterfaceType {
	return config.GCPVertexAI
}

// ValidateConfig implements domain.LLMProvider
func (c *GCPVertexAIClient) ValidateConfig() error {
	if c.projectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if c.model == "" {
		return fmt.Errorf("model is required")
	}
	return nil
}

// Close implements domain.LLMProvider
func (c *GCPVertexAIClient) Close() error {
	return nil
}

// convertToVertexContents converts domain messages to Vertex AI format
func (c *GCPVertexAIClient) convertToVertexContents(messages []domain.Message, systemPrompt string) []vertexContent {
	var contents []vertexContent
	
	// Add system prompt as first user message if present
	if systemPrompt != "" {
		contents = append(contents, vertexContent{
			Role: "user",
			Parts: []vertexPart{
				{Text: systemPrompt},
			},
		})
	}
	
	for _, msg := range messages {
		role := msg.Role
		if role == "system" {
			role = "user" // Vertex AI doesn't have system role
		}
		
		contents = append(contents, vertexContent{
			Role: role,
			Parts: []vertexPart{
				{Text: msg.Content},
			},
		})
	}
	
	return contents
}

// ensureAccessToken ensures we have a valid OAuth2 access token
func (c *GCPVertexAIClient) ensureAccessToken() error {
	// Check if token is still valid (with 5 minute buffer)
	if c.accessToken != "" && time.Now().Add(5*time.Minute).Before(c.tokenExpiry) {
		return nil
	}

	if c.serviceAccount == nil {
		return fmt.Errorf("service account credentials required - set credentials_path in config")
	}

	// Create JWT
	now := time.Now()
	claims := jwtClaims{
		Iss:   c.serviceAccount.ClientEmail,
		Scope: "https://www.googleapis.com/auth/cloud-platform",
		Aud:   c.serviceAccount.TokenURI,
		Exp:   now.Add(time.Hour).Unix(),
		Iat:   now.Unix(),
	}

	// Create JWT header and payload
	header := map[string]string{
		"alg": "RS256",
		"typ": "JWT",
	}

	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signInput := headerB64 + "." + claimsB64

	// Sign with private key
	privateKey, err := parsePrivateKey(c.serviceAccount.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	hash := sha256.Sum256([]byte(signInput))
	signature, err := rsa.SignPKCS1v15(nil, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return fmt.Errorf("failed to sign JWT: %w", err)
	}

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)
	jwt := signInput + "." + signatureB64

	// Exchange JWT for access token
	tokenReq := fmt.Sprintf("grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer&assertion=%s", jwt)

	httpReq, err := http.NewRequest("POST", c.serviceAccount.TokenURI, strings.NewReader(tokenReq))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OAuth2 token exchange failed (%s): %s", resp.Status, string(body))
	}

	var tokenResp oauth2TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	logging.Debug("Successfully obtained OAuth2 access token, expires at %v", c.tokenExpiry)

	return nil
}

// loadServiceAccount loads service account credentials from file
func (c *GCPVertexAIClient) loadServiceAccount(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read service account file: %w", err)
	}

	var sa gcpServiceAccount
	if err := json.Unmarshal(data, &sa); err != nil {
		return fmt.Errorf("failed to parse service account JSON: %w", err)
	}

	c.serviceAccount = &sa
	c.projectID = sa.ProjectID

	return nil
}

// parsePrivateKey parses RSA private key from PEM format
func parsePrivateKey(pemKey string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}

	return rsaKey, nil
}
