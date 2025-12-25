package rag

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// TermExpansionConfig defines how terms should be expanded
type TermExpansionConfig struct {
	Acronyms         map[string][]string `json:"acronyms,omitempty"`
	Synonyms         map[string][]string `json:"synonyms,omitempty"`
	DomainTerms      map[string][]string `json:"domain_terms,omitempty"`
	CaseSensitive    bool                `json:"case_sensitive,omitempty"`
	MaxExpansions    int                 `json:"max_expansions,omitempty"`
	PreservePhrases  []string            `json:"preserve_phrases,omitempty"`
}

// QueryExpansionConfig defines the overall query expansion strategy
type QueryExpansionConfig struct {
	EnableSynonymExpansion bool                `json:"enable_synonym_expansion"`
	EnableAcronymExpansion bool                `json:"enable_acronym_expansion"`
	EnableDomainExpansion  bool                `json:"enable_domain_expansion"`
	GenerateVariants       int                 `json:"generate_variants,omitempty"`
	PerspectiveAngles      []string            `json:"perspective_angles,omitempty"`
	TermExpansion          TermExpansionConfig `json:"term_expansion"`
	SemanticExpansion      bool                `json:"semantic_expansion,omitempty"`
}

// ExpandedQuery represents a query with its expansions
type ExpandedQuery struct {
	Original          string              `json:"original"`
	ExpandedVariants  []string            `json:"expanded_variants"`
	PerspectiveQueries []PerspectiveQuery `json:"perspective_queries"`
	ExpandedTerms     map[string][]string `json:"expanded_terms"`
	ExpansionMethods  []string            `json:"expansion_methods"`
}

// PerspectiveQuery represents a query from a specific perspective
type PerspectiveQuery struct {
	Perspective string `json:"perspective"`
	Query      string `json:"query"`
	Context    string `json:"context"`
}

// QueryExpander provides intelligent query expansion capabilities
type QueryExpander struct {
	config TermExpansionConfig
}

// NewQueryExpander creates a new query expander with the given configuration
func NewQueryExpander(config TermExpansionConfig) *QueryExpander {
	// Set defaults
	if config.MaxExpansions == 0 {
		config.MaxExpansions = 5
	}
	
	return &QueryExpander{
		config: config,
	}
}

// ExpandQuery expands a query using multiple strategies
func (qe *QueryExpander) ExpandQuery(ctx context.Context, originalQuery string, expansionConfig QueryExpansionConfig) (*ExpandedQuery, error) {
	logging.Info("🔄 Expanding query: %s", originalQuery)
	
	result := &ExpandedQuery{
		Original:         originalQuery,
		ExpandedVariants: []string{originalQuery}, // Always include original
		ExpandedTerms:    make(map[string][]string),
		ExpansionMethods: []string{},
	}
	
	// Synonym expansion
	if expansionConfig.EnableSynonymExpansion {
		synonymExpanded := qe.expandSynonyms(originalQuery)
		if len(synonymExpanded) > 1 { // More than just the original
			result.ExpandedVariants = append(result.ExpandedVariants, synonymExpanded[1:]...)
			result.ExpansionMethods = append(result.ExpansionMethods, "synonym_expansion")
			logging.Debug("✅ Added %d synonym expansions", len(synonymExpanded)-1)
		}
	}
	
	// Acronym expansion
	if expansionConfig.EnableAcronymExpansion {
		acronymExpanded := qe.expandAcronyms(originalQuery)
		if len(acronymExpanded) > 1 {
			result.ExpandedVariants = append(result.ExpandedVariants, acronymExpanded[1:]...)
			result.ExpansionMethods = append(result.ExpansionMethods, "acronym_expansion")
			logging.Debug("✅ Added %d acronym expansions", len(acronymExpanded)-1)
		}
	}
	
	// Domain-specific expansion
	if expansionConfig.EnableDomainExpansion {
		domainExpanded := qe.expandDomainTerms(originalQuery)
		if len(domainExpanded) > 1 {
			result.ExpandedVariants = append(result.ExpandedVariants, domainExpanded[1:]...)
			result.ExpansionMethods = append(result.ExpansionMethods, "domain_expansion")
			logging.Debug("✅ Added %d domain expansions", len(domainExpanded)-1)
		}
	}
	
	// Generate perspective-based queries
	if len(expansionConfig.PerspectiveAngles) > 0 {
		perspectiveQueries := qe.generatePerspectiveQueries(originalQuery, expansionConfig.PerspectiveAngles)
		result.PerspectiveQueries = perspectiveQueries
		result.ExpansionMethods = append(result.ExpansionMethods, "perspective_generation")
		logging.Debug("✅ Generated %d perspective queries", len(perspectiveQueries))
	}
	
	// Generate additional variants if requested
	if expansionConfig.GenerateVariants > 0 {
		additionalVariants := qe.generateQueryVariants(originalQuery, expansionConfig.GenerateVariants)
		result.ExpandedVariants = append(result.ExpandedVariants, additionalVariants...)
		result.ExpansionMethods = append(result.ExpansionMethods, "variant_generation")
		logging.Debug("✅ Generated %d additional variants", len(additionalVariants))
	}
	
	// Remove duplicates and limit results
	result.ExpandedVariants = qe.deduplicateAndLimit(result.ExpandedVariants, qe.config.MaxExpansions)
	
	logging.Info("🎉 Query expansion completed: %d variants, %d perspectives, methods: %v", 
		len(result.ExpandedVariants), len(result.PerspectiveQueries), result.ExpansionMethods)
	
	return result, nil
}

// expandSynonyms expands query using synonym mappings
func (qe *QueryExpander) expandSynonyms(query string) []string {
	expanded := []string{query}
	
	if len(qe.config.Synonyms) == 0 {
		return expanded
	}
	
	words := qe.tokenizeQuery(query)
	
	for _, word := range words {
		searchWord := word
		if !qe.config.CaseSensitive {
			searchWord = strings.ToLower(word)
		}
		
		if synonyms, exists := qe.config.Synonyms[searchWord]; exists {
			for _, synonym := range synonyms {
				expandedQuery := qe.replaceWordInQuery(query, word, synonym)
				if expandedQuery != query {
					expanded = append(expanded, expandedQuery)
				}
			}
		}
	}
	
	return expanded
}

// expandAcronyms expands acronyms in the query
func (qe *QueryExpander) expandAcronyms(query string) []string {
	expanded := []string{query}
	
	if len(qe.config.Acronyms) == 0 {
		return expanded
	}
	
	originalQuery := query
	for acronym, expansions := range qe.config.Acronyms {
		searchAcronym := acronym
		queryToSearch := query
		if !qe.config.CaseSensitive {
			searchAcronym = strings.ToLower(acronym)
			queryToSearch = strings.ToLower(query)
		}
		
		if strings.Contains(queryToSearch, searchAcronym) {
			for _, expansion := range expansions {
				expandedQuery := strings.ReplaceAll(originalQuery, acronym, expansion)
				if expandedQuery != originalQuery {
					expanded = append(expanded, expandedQuery)
				}
			}
		}
	}
	
	return expanded
}

// expandDomainTerms expands`
}
