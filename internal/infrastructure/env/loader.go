package env

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Store holds environment variables loaded from .env file
type Store struct {
	mu   sync.RWMutex
	vars map[string]string
}

var (
	// globalStore is the singleton env store
	globalStore *Store
	once        sync.Once
)

// GetStore returns the global environment store
func GetStore() *Store {
	once.Do(func() {
		globalStore = &Store{
			vars: make(map[string]string),
		}
	})
	return globalStore
}

// LoadFromFile loads environment variables from a .env file
// File format: KEY=VALUE (one per line)
// Comments start with #
// Returns number of variables loaded and any error
func (s *Store) LoadFromFile(filepath string) (int, error) {
	file, err := os.Open(filepath)
	if err != nil {
		// If file doesn't exist, that's ok - just return 0 loaded
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	defer file.Close()

	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip malformed lines
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove surrounding quotes if present
		value = strings.Trim(value, `"'`)
		
		if key != "" {
			s.vars[key] = value
			count++
		}
	}
	
	if err := scanner.Err(); err != nil {
		return count, err
	}
	
	return count, nil
}

// Get retrieves a value from the store
// Returns empty string if not found
func (s *Store) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.vars[key]
}

// Has checks if a key exists in the store
func (s *Store) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.vars[key]
	return exists
}

// GetWithFallback tries to get from store first, then falls back to os.Getenv
func (s *Store) GetWithFallback(key string) string {
	// Try store first
	if value := s.Get(key); value != "" {
		return value
	}
	// Fall back to system environment
	return os.Getenv(key)
}

// LoadDotEnv loads .env file from the same directory as the executable
// This is called automatically during init
func LoadDotEnv() error {
	store := GetStore()
	
	// Get executable path
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	
	// Get directory containing executable
	dir := filepath.Dir(exe)
	
	// Try to load .env file
	envPath := filepath.Join(dir, ".env")
	count, err := store.LoadFromFile(envPath)
	
	// We don't return error if file doesn't exist
	// Only return error for actual read failures
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	
	// Only log count if we actually loaded something
	// Never log the actual values for security
	if count > 0 {
		// Note: We log count only, never the actual keys or values
		_ = count // Loaded successfully
	}
	
	return nil
}

// ExpandEnv expands environment variables in a string
// Checks .env store first, then system environment
// Supports ${VAR} and $VAR formats
func ExpandEnv(s string) string {
	if s == "" {
		return s
	}
	
	store := GetStore()
	
	// Custom expand function that checks our store first
	mapper := func(key string) string {
		return store.GetWithFallback(key)
	}
	
	return os.Expand(s, mapper)
}

// Keys returns all keys in the store (for testing/debugging)
// WARNING: This should never be logged in production
func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	keys := make([]string, 0, len(s.vars))
	for k := range s.vars {
		keys = append(keys, k)
	}
	return keys
}

// Count returns the number of variables in the store
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.vars)
}

// Clear removes all variables from the store (for testing)
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vars = make(map[string]string)
}
