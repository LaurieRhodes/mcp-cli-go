// File: internal/services/workflow/embedding_utils.go
package workflow

import (
	"fmt"
	"math"
	"strings"
)

// truncateString truncates a string to the specified length with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// vectorToPostgresArray converts a vector to PostgreSQL array format
func vectorToPostgresArray(vector []float32) string {
	parts := make([]string, len(vector))
	for i, v := range vector {
		parts[i] = fmt.Sprintf("%.6f", v)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// buildFiltersClause builds a SQL WHERE clause from filters map
func buildFiltersClause(filters map[string]interface{}) string {
	if len(filters) == 0 {
		return ""
	}

	var conditions []string
	for key, value := range filters {
		switch v := value.(type) {
		case string:
			conditions = append(conditions, fmt.Sprintf("%s = '%s'", key, strings.ReplaceAll(v, "'", "''")))
		case bool:
			conditions = append(conditions, fmt.Sprintf("%s = %t", key, v))
		case int, int64:
			conditions = append(conditions, fmt.Sprintf("%s = %v", key, v))
		case float64, float32:
			conditions = append(conditions, fmt.Sprintf("%s = %v", key, v))
		default:
			conditions = append(conditions, fmt.Sprintf("%s = '%v'", key, v))
		}
	}

	return strings.Join(conditions, " AND ")
}

// Vector similarity calculation functions

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(vec1, vec2 []float32) float64 {
	if len(vec1) != len(vec2) {
		return 0.0
	}

	var dotProduct, normA, normB float64

	for i := 0; i < len(vec1); i++ {
		dotProduct += float64(vec1[i]) * float64(vec2[i])
		normA += float64(vec1[i]) * float64(vec1[i])
		normB += float64(vec2[i]) * float64(vec2[i])
	}

	normA = math.Sqrt(normA)
	normB = math.Sqrt(normB)

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (normA * normB)
}

// euclideanDistance calculates the Euclidean distance between two vectors
func euclideanDistance(vec1, vec2 []float32) float64 {
	if len(vec1) != len(vec2) {
		return -1
	}

	var sum float64
	for i := 0; i < len(vec1); i++ {
		diff := float64(vec1[i] - vec2[i])
		sum += diff * diff
	}

	return math.Sqrt(sum)
}

// dotProduct calculates the dot product of two vectors
func dotProduct(vec1, vec2 []float32) float64 {
	if len(vec1) != len(vec2) {
		return 0.0
	}

	var result float64
	for i := 0; i < len(vec1); i++ {
		result += float64(vec1[i]) * float64(vec2[i])
	}

	return result
}

// getParamKeys returns all keys from a map
func getParamKeys(params map[string]interface{}) []string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	return keys
}

// isParameterFormatError checks if an error is related to parameter formatting
func isParameterFormatError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "parameter") ||
		strings.Contains(errStr, "argument") ||
		strings.Contains(errStr, "missing") ||
		strings.Contains(errStr, "required")
}
