package clients

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
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

// GCPVertexAIOpenAIClient wraps OpenAICompatibleClient with OAuth2 token management for Vertex AI
// Uses OpenAI-compatible endpoint for chat/completions (supports tool calling)
// Uses native Vertex AI endpoint for embeddings (OpenAI endpoint doesn't support them)
type GCPVertexAIOpenAIClient struct {
	openaiClient   *OpenAICompatibleClient
	projectID      string
	location       string
	serviceAccount *gcpServiceAccount
	httpClient     *http.Client
	accessToken    string
	tokenExpiry    time.Time
	providerType   domain.ProviderType
	config         *config.ProviderConfig
}

// NewGCPVertexAIOpenAIClient creates a Vertex AI client using OpenAI-compatible endpoint
func NewGCPVertexAIOpenAIClient(providerType domain.ProviderType, cfg *config.ProviderConfig) (domain.LLMProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	projectID := cfg.ProjectID
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required for Vertex AI")
	}

	location := cfg.Location
	if location == "" {
		location = "us-central1"
	}

	credentialsPath := cfg.CredentialsPath
	if credentialsPath == "" {
		return nil, fmt.Errorf("credentials_path is required for Vertex AI")
	}

	model := cfg.DefaultModel
	if model == "" {
		model = "gemini-2.5-flash"
	}

	// OpenAI-compatible endpoint requires "publisher/model" format
	// Convert "gemini-2.5-flash" to "google/gemini-2.5-flash"
	if !strings.Contains(model, "/") {
		model = "google/" + model
		logging.Debug("Converted model name to OpenAI-compatible format: %s", model)
	}

	timeout := 45 * time.Second
	if cfg.TimeoutSeconds > 0 {
		timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}

	// Create wrapper client
	wrapper := &GCPVertexAIOpenAIClient{
		projectID:    projectID,
		location:     location,
		httpClient:   &http.Client{Timeout: timeout},
		providerType: providerType,
		config:       cfg,
	}

	// Load service account
	if err := wrapper.loadServiceAccount(credentialsPath); err != nil {
		return nil, fmt.Errorf("failed to load service account: %w", err)
	}

	// Get initial OAuth2 token
	if err := wrapper.ensureAccessToken(); err != nil {
		return nil, fmt.Errorf("failed to obtain initial access token: %w", err)
	}

	// Construct OpenAI-compatible endpoint
	// Format: https://{location}-aiplatform.googleapis.com/v1beta1/projects/{project}/locations/{location}/endpoints/openapi
	openaiEndpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1beta1/projects/%s/locations/%s/endpoints/openapi",
		location, projectID, location)

	// Create modified config for OpenAI client
	openaiConfig := &config.ProviderConfig{
		APIKey:          wrapper.accessToken, // Will be updated before each request
		APIEndpoint:     openaiEndpoint,
		DefaultModel:    model, // Now in "google/gemini-2.5-flash" format
		TimeoutSeconds:  cfg.TimeoutSeconds,
		MaxRetries:      cfg.MaxRetries,
		EmbeddingModels: cfg.EmbeddingModels,
	}

	// Create OpenAI-compatible client
	openaiClient, err := NewOpenAICompatibleClient(providerType, openaiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI client: %w", err)
	}

	wrapper.openaiClient = openaiClient.(*OpenAICompatibleClient)

	logging.Info("Created Vertex AI OpenAI-compatible client for project %s, location %s, model %s", projectID, location, model)
	logging.Info("Using OpenAI-compatible endpoint: %s", openaiEndpoint)

	return wrapper, nil
}

// CreateCompletion implements domain.LLMProvider
func (c *GCPVertexAIOpenAIClient) CreateCompletion(ctx context.Context, req *domain.CompletionRequest) (*domain.CompletionResponse, error) {
	// Refresh OAuth2 token and update OpenAI client
	if err := c.refreshClientToken(); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Delegate to OpenAI client
	return c.openaiClient.CreateCompletion(ctx, req)
}

// StreamCompletion implements domain.LLMProvider
func (c *GCPVertexAIOpenAIClient) StreamCompletion(ctx context.Context, req *domain.CompletionRequest, writer io.Writer) (*domain.CompletionResponse, error) {
	// Refresh OAuth2 token and update OpenAI client
	if err := c.refreshClientToken(); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Delegate to OpenAI client
	return c.openaiClient.StreamCompletion(ctx, req, writer)
}

// CreateEmbeddings implements domain.LLMProvider
// NOTE: Uses native Vertex AI endpoint because OpenAI-compatible endpoint doesn't support embeddings
func (c *GCPVertexAIOpenAIClient) CreateEmbeddings(ctx context.Context, req *domain.EmbeddingRequest) (*domain.EmbeddingResponse, error) {
	if len(req.Input) == 0 {
		return nil, fmt.Errorf("input is required for embeddings")
	}

	// Ensure we have a valid access token
	if err := c.ensureAccessToken(); err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Use gecko embedding model (native Vertex AI format, not OpenAI format)
	embeddingModel := "text-embedding-004"
	if req.Model != "" {
		// Strip "google/" prefix if present (came from config)
		embeddingModel = strings.TrimPrefix(req.Model, "google/")
	}

	logging.Info("Using native Vertex AI embedding endpoint for model: %s", embeddingModel)

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

	// Native Vertex AI embedding endpoint (not OpenAI-compatible)
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict",
		c.location, c.projectID, c.location, embeddingModel)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Vertex AI embedding API error (%s): %s", resp.Status, string(body))
	}

	var vertexResp vertexEmbeddingResponse
	if err := json.Unmarshal(body, &vertexResp); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
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

// GetSupportedEmbeddingModels implements domain.LLMProvider
// Returns models in native Vertex AI format (no "google/" prefix)
func (c *GCPVertexAIOpenAIClient) GetSupportedEmbeddingModels() []string {
	// Return native Vertex AI embedding model names
	// These work with the native endpoint, not the OpenAI-compatible endpoint
	return []string{
		"text-embedding-004",
		"text-multilingual-embedding-002",
		"textembedding-gecko@003",
		"textembedding-gecko-multilingual@001",
	}
}

// GetMaxEmbeddingTokens implements domain.LLMProvider
func (c *GCPVertexAIOpenAIClient) GetMaxEmbeddingTokens(model string) int {
	return c.openaiClient.GetMaxEmbeddingTokens(model)
}

// GetProviderType implements domain.LLMProvider
func (c *GCPVertexAIOpenAIClient) GetProviderType() domain.ProviderType {
	return c.providerType
}

// GetInterfaceType implements domain.LLMProvider
func (c *GCPVertexAIOpenAIClient) GetInterfaceType() config.InterfaceType {
	return config.OpenAICompatible // Report as OpenAI-compatible
}

// ValidateConfig implements domain.LLMProvider
func (c *GCPVertexAIOpenAIClient) ValidateConfig() error {
	if c.projectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if c.serviceAccount == nil {
		return fmt.Errorf("service account credentials required")
	}
	return nil
}

// Close implements domain.LLMProvider
func (c *GCPVertexAIOpenAIClient) Close() error {
	if c.openaiClient != nil {
		return c.openaiClient.Close()
	}
	return nil
}

// refreshClientToken ensures token is fresh and updates the OpenAI client
func (c *GCPVertexAIOpenAIClient) refreshClientToken() error {
	if err := c.ensureAccessToken(); err != nil {
		return err
	}

	// Update the OpenAI client's API key with fresh token
	c.openaiClient.apiKey = c.accessToken

	return nil
}

// ensureAccessToken ensures we have a valid OAuth2 access token
func (c *GCPVertexAIOpenAIClient) ensureAccessToken() error {
	// Check if token is still valid (with 5 minute buffer)
	if c.accessToken != "" && time.Now().Add(5*time.Minute).Before(c.tokenExpiry) {
		return nil
	}

	logging.Debug("Refreshing Vertex AI OAuth2 token...")

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
func (c *GCPVertexAIOpenAIClient) loadServiceAccount(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read service account file: %w", err)
	}

	var sa gcpServiceAccount
	if err := json.Unmarshal(data, &sa); err != nil {
		return fmt.Errorf("failed to parse service account JSON: %w", err)
	}

	c.serviceAccount = &sa

	logging.Info("Loaded service account: %s", sa.ClientEmail)

	return nil
}

// Note: parsePrivateKey() is defined in gcp_vertex_ai.go at package level
