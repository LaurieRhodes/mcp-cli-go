package chunking

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/core/tokens"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// SentenceChunker implements sentence-based chunking strategy
type SentenceChunker struct {
	tokenManager *tokens.TokenManager
	overlap      int
}

// NewSentenceChunker creates a new sentence-based chunker
func NewSentenceChunker(tokenManager *tokens.TokenManager, overlap int) *SentenceChunker {
	return &SentenceChunker{
		tokenManager: tokenManager,
		overlap:      overlap,
	}
}

// ChunkText splits text into chunks based on sentence boundaries
func (sc *SentenceChunker) ChunkText(text string, maxTokens int) ([]domain.TextChunk, error) {
	if text == "" {
		return []domain.TextChunk{}, nil
	}

	// Split text into sentences using regex that handles various sentence endings
	sentenceRegex := regexp.MustCompile(`(?:[.!?]+\s+|\n+)`)
	sentences := sentenceRegex.Split(text, -1)
	
	// Clean up sentences and track positions
	var cleanSentences []string
	var sentencePositions []int
	currentPos := 0
	
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence != "" {
			cleanSentences = append(cleanSentences, sentence)
			sentencePositions = append(sentencePositions, currentPos)
			currentPos += len(sentence)
		}
	}
	
	if len(cleanSentences) == 0 {
		return []domain.TextChunk{}, nil
	}

	var chunks []domain.TextChunk
	var currentChunk strings.Builder
	var currentStartPos int
	var currentSentenceIndices []int
	chunkIndex := 0

	for i, sentence := range cleanSentences {
		// Check if adding this sentence would exceed maxTokens
		testText := currentChunk.String()
		if testText != "" {
			testText += " "
		}
		testText += sentence
		
		tokenCount := sc.tokenManager.CountTokensInString(testText)
		
		if tokenCount > maxTokens && currentChunk.Len() > 0 {
			// Current chunk would be too large, finalize current chunk
			chunkText := currentChunk.String()
			if chunkText != "" {
				endPos := sentencePositions[currentSentenceIndices[len(currentSentenceIndices)-1]] + len(cleanSentences[currentSentenceIndices[len(currentSentenceIndices)-1]])
				
				chunk := domain.TextChunk{
					Text:       chunkText,
					Index:      chunkIndex,
					StartPos:   currentStartPos,
					EndPos:     endPos,
					TokenCount: sc.tokenManager.CountTokensInString(chunkText),
				}
				chunks = append(chunks, chunk)
				chunkIndex++
			}
			
			// Start new chunk with overlap
			overlapSentences := sc.getOverlapSentences(currentSentenceIndices, cleanSentences, sc.overlap)
			currentChunk.Reset()
			currentSentenceIndices = []int{}
			
			if len(overlapSentences) > 0 {
				currentChunk.WriteString(strings.Join(overlapSentences, " "))
				currentStartPos = sentencePositions[i-len(overlapSentences)]
				for j := i - len(overlapSentences); j < i; j++ {
					currentSentenceIndices = append(currentSentenceIndices, j)
				}
			} else {
				currentStartPos = sentencePositions[i]
			}
		}
		
		// Add current sentence
		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(sentence)
		currentSentenceIndices = append(currentSentenceIndices, i)
	}
	
	// Add final chunk if there's content
	if currentChunk.Len() > 0 {
		chunkText := currentChunk.String()
		endPos := sentencePositions[currentSentenceIndices[len(currentSentenceIndices)-1]] + len(cleanSentences[currentSentenceIndices[len(currentSentenceIndices)-1]])
		
		chunk := domain.TextChunk{
			Text:       chunkText,
			Index:      chunkIndex,
			StartPos:   currentStartPos,
			EndPos:     endPos,
			TokenCount: sc.tokenManager.CountTokensInString(chunkText),
		}
		chunks = append(chunks, chunk)
	}

	logging.Debug("Sentence chunking complete: %d chunks created from %d sentences", len(chunks), len(cleanSentences))
	return chunks, nil
}

// getOverlapSentences returns sentences for overlap between chunks
func (sc *SentenceChunker) getOverlapSentences(currentIndices []int, sentences []string, overlap int) []string {
	if overlap <= 0 || len(currentIndices) == 0 {
		return []string{}
	}
	
	// Take the last 'overlap' sentences from current chunk
	startIdx := len(currentIndices) - overlap
	if startIdx < 0 {
		startIdx = 0
	}
	
	var overlapSentences []string
	for i := startIdx; i < len(currentIndices); i++ {
		sentenceIdx := currentIndices[i]
		if sentenceIdx < len(sentences) {
			overlapSentences = append(overlapSentences, sentences[sentenceIdx])
		}
	}
	
	return overlapSentences
}

// GetName returns the name of this chunking strategy
func (sc *SentenceChunker) GetName() string {
	return "sentence"
}

// GetDescription returns a description of this chunking strategy
func (sc *SentenceChunker) GetDescription() string {
	return "Splits text at sentence boundaries while preserving semantic meaning"
}

// ParagraphChunker implements paragraph-based chunking strategy
type ParagraphChunker struct {
	tokenManager *tokens.TokenManager
	overlap      int
}

// NewParagraphChunker creates a new paragraph-based chunker
func NewParagraphChunker(tokenManager *tokens.TokenManager, overlap int) *ParagraphChunker {
	return &ParagraphChunker{
		tokenManager: tokenManager,
		overlap:      overlap,
	}
}

// ChunkText splits text into chunks based on paragraph boundaries
func (pc *ParagraphChunker) ChunkText(text string, maxTokens int) ([]domain.TextChunk, error) {
	if text == "" {
		return []domain.TextChunk{}, nil
	}

	// Split text into paragraphs (double newlines or more)
	paragraphRegex := regexp.MustCompile(`\s*\
`)
	paragraphs := paragraphRegex.Split(text, -1)
	
	// Clean up paragraphs and track positions
	var cleanParagraphs []string
	var paragraphPositions []int
	currentPos := 0
	
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph != "" {
			cleanParagraphs = append(cleanParagraphs, paragraph)
			paragraphPositions = append(paragraphPositions, currentPos)
			currentPos += len(paragraph)
		}
	}
	
	if len(cleanParagraphs) == 0 {
		return []domain.TextChunk{}, nil
	}

	var chunks []domain.TextChunk
	var currentChunk strings.Builder
	var currentStartPos int
	var currentParagraphIndices []int
	chunkIndex := 0

	for i, paragraph := range cleanParagraphs {
		// Check if adding this paragraph would exceed maxTokens
		testText := currentChunk.String()
		if testText != "" {
			testText += ""
		}
		testText += paragraph
		
		tokenCount := pc.tokenManager.CountTokensInString(testText)
		
		if tokenCount > maxTokens && currentChunk.Len() > 0 {
			// Current chunk would be too large, finalize current chunk
			chunkText := currentChunk.String()
			if chunkText != "" {
				endPos := paragraphPositions[currentParagraphIndices[len(currentParagraphIndices)-1]] + len(cleanParagraphs[currentParagraphIndices[len(currentParagraphIndices)-1]])
				
				chunk := domain.TextChunk{
					Text:       chunkText,
					Index:      chunkIndex,
					StartPos:   currentStartPos,
					EndPos:     endPos,
					TokenCount: pc.tokenManager.CountTokensInString(chunkText),
				}
				chunks = append(chunks, chunk)
				chunkIndex++
			}
			
			// Start new chunk with overlap
			overlapParagraphs := pc.getOverlapParagraphs(currentParagraphIndices, cleanParagraphs, pc.overlap)
			currentChunk.Reset()
			currentParagraphIndices = []int{}
			
			if len(overlapParagraphs) > 0 {
				currentChunk.WriteString(strings.Join(overlapParagraphs, ""))
				currentStartPos = paragraphPositions[i-len(overlapParagraphs)]
				for j := i - len(overlapParagraphs); j < i; j++ {
					currentParagraphIndices = append(currentParagraphIndices, j)
				}
			} else {
				currentStartPos = paragraphPositions[i]
			}
		}
		
		// Add current paragraph
		if currentChunk.Len() > 0 {
			currentChunk.WriteString("")
		}
		currentChunk.WriteString(paragraph)
		currentParagraphIndices = append(currentParagraphIndices, i)
	}
	
	// Add final chunk if there's content
	if currentChunk.Len() > 0 {
		chunkText := currentChunk.String()
		endPos := paragraphPositions[currentParagraphIndices[len(currentParagraphIndices)-1]] + len(cleanParagraphs[currentParagraphIndices[len(currentParagraphIndices)-1]])
		
		chunk := domain.TextChunk{
			Text:       chunkText,
			Index:      chunkIndex,
			StartPos:   currentStartPos,
			EndPos:     endPos,
			TokenCount: pc.tokenManager.CountTokensInString(chunkText),
		}
		chunks = append(chunks, chunk)
	}

	logging.Debug("Paragraph chunking complete: %d chunks created from %d paragraphs", len(chunks), len(cleanParagraphs))
	return chunks, nil
}

// getOverlapParagraphs returns paragraphs for overlap between chunks
func (pc *ParagraphChunker) getOverlapParagraphs(currentIndices []int, paragraphs []string, overlap int) []string {
	if overlap <= 0 || len(currentIndices) == 0 {
		return []string{}
	}
	
	// Take the last 'overlap' paragraphs from current chunk
	startIdx := len(currentIndices) - overlap
	if startIdx < 0 {
		startIdx = 0
	}
	
	var overlapParagraphs []string
	for i := startIdx; i < len(currentIndices); i++ {
		paragraphIdx := currentIndices[i]
		if paragraphIdx < len(paragraphs) {
			overlapParagraphs = append(overlapParagraphs, paragraphs[paragraphIdx])
		}
	}
	
	return overlapParagraphs
}

// GetName returns the name of this chunking strategy
func (pc *ParagraphChunker) GetName() string {
	return "paragraph"
}

// GetDescription returns a description of this chunking strategy
func (pc *ParagraphChunker) GetDescription() string {
	return "Splits text at paragraph boundaries to preserve document structure"
}

// FixedChunker implements fixed-size chunking strategy
type FixedChunker struct {
	tokenManager *tokens.TokenManager
	overlap      int
}

// NewFixedChunker creates a new fixed-size chunker
func NewFixedChunker(tokenManager *tokens.TokenManager, overlap int) *FixedChunker {
	return &FixedChunker{
		tokenManager: tokenManager,
		overlap:      overlap,
	}
}

// ChunkText splits text into fixed-size chunks with overlap
func (fc *FixedChunker) ChunkText(text string, maxTokens int) ([]domain.TextChunk, error) {
	if text == "" {
		return []domain.TextChunk{}, nil
	}

	// Split text into words to have finer control
	words := strings.Fields(text)
	if len(words) == 0 {
		return []domain.TextChunk{}, nil
	}

	var chunks []domain.TextChunk
	chunkIndex := 0
	overlapWords := 0
	
	if fc.overlap > 0 {
		// Calculate overlap in words (approximate)
		totalTokens := fc.tokenManager.CountTokensInString(text)
		wordsPerToken := float64(len(words)) / float64(totalTokens)
		overlapWords = int(float64(fc.overlap) * wordsPerToken)
	}

	for i := 0; i < len(words); {
		var chunkWords []string
		var currentTokens int
		startWordIndex := i
		
		// Add overlap from previous chunk
		if chunkIndex > 0 && overlapWords > 0 {
			overlapStart := i - overlapWords
			if overlapStart < 0 {
				overlapStart = 0
			}
			for j := overlapStart; j < i; j++ {
				chunkWords = append(chunkWords, words[j])
			}
		}
		
		// Add words until we reach maxTokens
		for i < len(words) {
			testChunk := append(chunkWords, words[i])
			testText := strings.Join(testChunk, " ")
			testTokens := fc.tokenManager.CountTokensInString(testText)
			
			if testTokens > maxTokens && len(chunkWords) > 0 {
				// Don't add this word, chunk is full
				break
			}
			
			chunkWords = append(chunkWords, words[i])
			currentTokens = testTokens
			i++
		}
		
		// Create chunk
		if len(chunkWords) > 0 {
			chunkText := strings.Join(chunkWords, " ")
			
			// Calculate positions (approximate)
			startPos := 0
			if startWordIndex > 0 {
				beforeText := strings.Join(words[:startWordIndex], " ")
				startPos = len(beforeText) + 1 // +1 for space
			}
			endPos := startPos + len(chunkText)
			
			chunk := domain.TextChunk{
				Text:       chunkText,
				Index:      chunkIndex,
				StartPos:   startPos,
				EndPos:     endPos,
				TokenCount: currentTokens,
			}
			chunks = append(chunks, chunk)
			chunkIndex++
		}
		
		// If we haven't moved forward, advance by one word to avoid infinite loop
		if i == startWordIndex {
			i++
		}
	}

	logging.Debug("Fixed chunking complete: %d chunks created with max %d tokens each", len(chunks), maxTokens)
	return chunks, nil
}

// GetName returns the name of this chunking strategy
func (fc *FixedChunker) GetName() string {
	return "fixed"
}

// GetDescription returns a description of this chunking strategy
func (fc *FixedChunker) GetDescription() string {
	return "Splits text into fixed-size chunks with configurable overlap"
}

// ChunkingManager manages different chunking strategies
type ChunkingManager struct {
	strategies map[domain.ChunkingType]func(*tokens.TokenManager, int) domain.ChunkingStrategy
}

// NewChunkingManager creates a new chunking manager
func NewChunkingManager() *ChunkingManager {
	return &ChunkingManager{
		strategies: map[domain.ChunkingType]func(*tokens.TokenManager, int) domain.ChunkingStrategy{
			domain.ChunkingSentence:  func(tm *tokens.TokenManager, overlap int) domain.ChunkingStrategy { return NewSentenceChunker(tm, overlap) },
			domain.ChunkingParagraph: func(tm *tokens.TokenManager, overlap int) domain.ChunkingStrategy { return NewParagraphChunker(tm, overlap) },
			domain.ChunkingFixed:     func(tm *tokens.TokenManager, overlap int) domain.ChunkingStrategy { return NewFixedChunker(tm, overlap) },
		},
	}
}

// GetStrategy returns a chunking strategy instance
func (cm *ChunkingManager) GetStrategy(strategyType domain.ChunkingType, tokenManager *tokens.TokenManager, overlap int) (domain.ChunkingStrategy, error) {
	factory, exists := cm.strategies[strategyType]
	if !exists {
		return nil, fmt.Errorf("unsupported chunking strategy: %s", strategyType)
	}
	
	return factory(tokenManager, overlap), nil
}

// GetAvailableStrategies returns all available chunking strategies
func (cm *ChunkingManager) GetAvailableStrategies() []domain.ChunkingType {
	var strategies []domain.ChunkingType
	for strategy := range cm.strategies {
		strategies = append(strategies, strategy)
	}
	return strategies
}

// GetStrategyDescription returns description for a strategy
func (cm *ChunkingManager) GetStrategyDescription(strategyType domain.ChunkingType) string {
	// Create a temporary instance to get description
	dummyTokenManager, err := tokens.NewTokenManagerFallback("gpt-4")
	if err != nil {
		return "Description unavailable"
	}
	
	strategy, err := cm.GetStrategy(strategyType, dummyTokenManager, 0)
	if err != nil {
		return "Description unavailable"
	}
	
	return strategy.GetDescription()
}
