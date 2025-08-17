package web

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"j5.nz/gw2/internal/cache"
)

const (
	DefaultCacheSize = 1000 // Maximum number of cached items
	DefaultTTL       = 5 * time.Minute // Cache items for 5 minutes
	CleanupInterval  = 1 * time.Minute // Clean up expired items every minute
)

// WebCache wraps the generic cache for web-specific functionality
type WebCache struct {
	cache cache.Cache
}

// NewWebCache creates a new web cache instance
func NewWebCache() *WebCache {
	lruCache := cache.NewLRUCache(DefaultCacheSize)
	lruCache.StartCleanupRoutine(CleanupInterval)
	
	return &WebCache{
		cache: lruCache,
	}
}

// CacheKey generates a cache key for HTTP requests
func (wc *WebCache) CacheKey(r *http.Request) string {
	// Create a unique key based on path and query parameters
	key := r.URL.Path + "?" + r.URL.RawQuery
	
	// Hash the key if it's too long to avoid memory issues
	if len(key) > 250 {
		hash := md5.Sum([]byte(key))
		return fmt.Sprintf("%x", hash)
	}
	
	return key
}

// Get retrieves data from cache
func (wc *WebCache) Get(key string) ([]byte, bool) {
	if value, found := wc.cache.Get(key); found {
		if data, ok := value.([]byte); ok {
			return data, true
		}
	}
	return nil, false
}

// Set stores data in cache
func (wc *WebCache) Set(key string, data []byte) {
	wc.cache.Set(key, data, DefaultTTL)
}

// SetWithTTL stores data in cache with custom TTL
func (wc *WebCache) SetWithTTL(key string, data []byte, ttl time.Duration) {
	wc.cache.Set(key, data, ttl)
}

// Delete removes an item from cache
func (wc *WebCache) Delete(key string) {
	wc.cache.Delete(key)
}

// Clear removes all items from cache
func (wc *WebCache) Clear() {
	wc.cache.Clear()
}

// Stats returns cache statistics
func (wc *WebCache) Stats() cache.Stats {
	return wc.cache.Stats()
}

// CacheResponse represents a cached HTTP response
type CacheResponse struct {
	Body        []byte            `json:"body"`
	Headers     map[string]string `json:"headers"`
	StatusCode  int               `json:"status_code"`
	ContentType string            `json:"content_type"`
	CachedAt    time.Time         `json:"cached_at"`
}

// SetResponse caches an HTTP response
func (wc *WebCache) SetResponse(key string, statusCode int, headers http.Header, body []byte) {
	response := CacheResponse{
		Body:       body,
		Headers:    make(map[string]string),
		StatusCode: statusCode,
		CachedAt:   time.Now(),
	}
	
	// Store important headers
	for name, values := range headers {
		if len(values) > 0 {
			response.Headers[name] = values[0]
		}
	}
	
	if contentType := headers.Get("Content-Type"); contentType != "" {
		response.ContentType = contentType
	}
	
	if data, err := json.Marshal(response); err == nil {
		wc.Set(key, data)
	}
}

// GetResponse retrieves a cached HTTP response
func (wc *WebCache) GetResponse(key string) (*CacheResponse, bool) {
	if data, found := wc.Get(key); found {
		var response CacheResponse
		if err := json.Unmarshal(data, &response); err == nil {
			return &response, true
		}
	}
	return nil, false
}

// CacheMiddleware creates middleware for caching HTTP responses
func (wc *WebCache) CacheMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only cache GET requests
		if r.Method != http.MethodGet {
			next(w, r)
			return
		}
		
		// Generate cache key
		key := wc.CacheKey(r)
		
		// Try to get from cache
		if cachedResponse, found := wc.GetResponse(key); found {
			// Restore headers
			for name, value := range cachedResponse.Headers {
				w.Header().Set(name, value)
			}
			
			// Set content type
			if cachedResponse.ContentType != "" {
				w.Header().Set("Content-Type", cachedResponse.ContentType)
			}
			
			// Add cache headers
			w.Header().Set("X-Cache", "HIT")
			w.Header().Set("X-Cache-Date", cachedResponse.CachedAt.Format(time.RFC3339))
			
			// Write response
			w.WriteHeader(cachedResponse.StatusCode)
			w.Write(cachedResponse.Body)
			return
		}
		
		// Create a response writer that captures the response
		capture := &responseCapture{
			ResponseWriter: w,
			statusCode:     200,
		}
		
		// Add miss header
		w.Header().Set("X-Cache", "MISS")
		
		// Call the next handler
		next(capture, r)
		
		// Cache the response if it was successful
		if capture.statusCode >= 200 && capture.statusCode < 300 && len(capture.body) > 0 {
			wc.SetResponse(key, capture.statusCode, capture.Header(), capture.body)
		}
	}
}

// responseCapture captures HTTP response data for caching
type responseCapture struct {
	http.ResponseWriter
	body       []byte
	statusCode int
}

func (rc *responseCapture) Write(data []byte) (int, error) {
	rc.body = append(rc.body, data...)
	return rc.ResponseWriter.Write(data)
}

func (rc *responseCapture) WriteHeader(statusCode int) {
	rc.statusCode = statusCode
	rc.ResponseWriter.WriteHeader(statusCode)
}

// InvalidatePattern removes all cache entries matching a pattern
func (wc *WebCache) InvalidatePattern(pattern string) {
	// This is a simplified implementation
	// In a production system, you might want a more sophisticated pattern matching
	wc.cache.Clear()
}