// Package gw2api provides a fully typed client for the Guild Wars 2 API v2
package gw2api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

const (
	BaseURL     = "https://api.guildwars2.com"
	UserAgent   = "gw2api-go/1.0"
	DefaultLang = LanguageEnglish
)

// RetryConfig configures retry behavior for API requests
type RetryConfig struct {
	MaxRetries      int
	BaseDelay       time.Duration
	MaxDelay        time.Duration
	BackoffMultiple float64
}

// Client provides access to the Guild Wars 2 API
type Client struct {
	baseURL     string
	httpClient  *http.Client
	apiKey      string
	language    Language
	userAgent   string
	dataCache   *DataCache
	rateLimiter *rate.Limiter
	retryConfig *RetryConfig
	verbose     bool
}

// ClientOption configures a Client
type ClientOption func(*Client)

// WithAPIKey sets the API key for authenticated endpoints
func WithAPIKey(key string) ClientOption {
	return func(c *Client) {
		c.apiKey = key
	}
}

// WithLanguage sets the default language for localized content
func WithLanguage(lang Language) ClientOption {
	return func(c *Client) {
		c.language = lang
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithUserAgent sets a custom user agent
func WithUserAgent(ua string) ClientOption {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// WithVerboseLogging enables verbose request/response logging
func WithVerboseLogging() ClientOption {
	return func(c *Client) {
		c.verbose = true
	}
}

// WithRateLimit sets a custom rate limit (requests per second)
func WithRateLimit(requestsPerSecond float64) ClientOption {
	return func(c *Client) {
		c.rateLimiter = rate.NewLimiter(rate.Limit(requestsPerSecond), 1)
	}
}

// WithRetryConfig sets custom retry configuration for handling server issues
func WithRetryConfig(config *RetryConfig) ClientOption {
	return func(c *Client) {
		c.retryConfig = config
	}
}

// WithRetries sets basic retry configuration with default exponential backoff
func WithRetries(maxRetries int) ClientOption {
	return func(c *Client) {
		c.retryConfig = &RetryConfig{
			MaxRetries:      maxRetries,
			BaseDelay:       500 * time.Millisecond,
			MaxDelay:        30 * time.Second,
			BackoffMultiple: 2.0,
		}
	}
}

// WithDataCache enables comprehensive data caching and loads data from the specified directory
func WithDataCache(dataDir string) ClientOption {
	return func(c *Client) {
		c.dataCache = NewDataCache()
		if err := c.dataCache.LoadFromDirectory(dataDir); err != nil {
			// Log error but don't fail client creation
			fmt.Printf("Warning: Failed to load data cache from %s: %v\n", dataDir, err)
		}
	}
}

// WithItemCache enables item caching and loads items from the specified file (deprecated - use WithDataCache)
func WithItemCache(filePath string) ClientOption {
	return func(c *Client) {
		c.dataCache = NewDataCache()
		if err := c.dataCache.GetItemCache().LoadFromFile(filePath); err != nil {
			// Log error but don't fail client creation
			fmt.Printf("Warning: Failed to load item cache from %s: %v\n", filePath, err)
		}
	}
}

// DataCache returns the client's data cache (if available)
func (c *Client) DataCache() *DataCache {
	return c.dataCache
}

// NewClient creates a new GW2 API client
func NewClient(options ...ClientOption) *Client {
	c := &Client{
		baseURL:     BaseURL,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		language:    DefaultLang,
		userAgent:   UserAgent,
		rateLimiter: rate.NewLimiter(5, 1), // Default: 5 requests per second
		retryConfig: &RetryConfig{ // Default retry config for server downtime
			MaxRetries:      3,
			BaseDelay:       1 * time.Second,
			MaxDelay:        30 * time.Second,
			BackoffMultiple: 2.0,
		},
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

// HTTPError represents an HTTP error with status code
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if httpErr, ok := err.(HTTPError); ok {
		// Retry server errors (5xx) and rate limiting (429)
		return httpErr.StatusCode >= 500 || httpErr.StatusCode == 429
	}
	
	// Retry network errors, timeouts, etc.
	return true
}

// calculateBackoffDelay calculates the delay for exponential backoff
func (c *Client) calculateBackoffDelay(attempt int) time.Duration {
	if c.retryConfig == nil {
		return 0
	}
	
	delay := time.Duration(float64(c.retryConfig.BaseDelay) * math.Pow(c.retryConfig.BackoffMultiple, float64(attempt)))
	if delay > c.retryConfig.MaxDelay {
		delay = c.retryConfig.MaxDelay
	}
	return delay
}

// RequestOptions configures individual API requests
type RequestOptions struct {
	Language      Language
	IDs           []int
	Page          int
	PageSize      int
	All           bool
	SchemaVersion string
}

// RequestOption configures a request
type RequestOption func(*RequestOptions)

// WithLang overrides the default language for this request
func WithLang(lang Language) RequestOption {
	return func(o *RequestOptions) {
		o.Language = lang
	}
}

// WithIDs requests specific IDs
func WithIDs(ids ...int) RequestOption {
	return func(o *RequestOptions) {
		o.IDs = ids
	}
}

// WithPage requests a specific page
func WithPage(page int) RequestOption {
	return func(o *RequestOptions) {
		o.Page = page
	}
}

// WithPageSize sets the page size
func WithPageSize(size int) RequestOption {
	return func(o *RequestOptions) {
		o.PageSize = size
	}
}

// WithAll requests all available items
func WithAll() RequestOption {
	return func(o *RequestOptions) {
		o.All = true
	}
}

// WithSchemaVersion sets the schema version
func WithSchemaVersion(version string) RequestOption {
	return func(o *RequestOptions) {
		o.SchemaVersion = version
	}
}

// get performs a GET request to the API
func (c *Client) get(ctx context.Context, endpoint string, opts *RequestOptions) ([]byte, *PaginationResponse, error) {
	if c.retryConfig == nil || c.retryConfig.MaxRetries == 0 {
		return c.makeRequest(ctx, endpoint, opts)
	}

	var lastErr error
	
	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := c.calculateBackoffDelay(attempt - 1)
			select {
			case <-ctx.Done():
				return nil, nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		body, pagination, err := c.makeRequest(ctx, endpoint, opts)
		
		// Success case
		if err == nil {
			return body, pagination, nil
		}

		lastErr = err
		
		// Check if error is retryable
		if !isRetryableError(err) {
			return nil, nil, err
		}
	}

	return nil, nil, fmt.Errorf("request failed after %d retries: %w", c.retryConfig.MaxRetries, lastErr)
}

func (c *Client) makeRequest(ctx context.Context, endpoint string, opts *RequestOptions) ([]byte, *PaginationResponse, error) {
	u, err := url.Parse(c.baseURL + endpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid endpoint: %w", err)
	}

	q := u.Query()

	// Add language parameter
	lang := c.language
	if opts != nil && opts.Language != "" {
		lang = opts.Language
	}
	if lang != "" {
		q.Set("lang", string(lang))
	}

	// Add authentication
	if c.apiKey != "" {
		q.Set("access_token", c.apiKey)
	}

	// Add schema version
	if opts != nil && opts.SchemaVersion != "" {
		q.Set("v", opts.SchemaVersion)
	}

	// Add bulk expansion parameters
	if opts != nil {
		if opts.All {
			q.Set("ids", "all")
		} else if len(opts.IDs) > 0 {
			ids := make([]string, len(opts.IDs))
			for i, id := range opts.IDs {
				ids[i] = strconv.Itoa(id)
			}
			q.Set("ids", strings.Join(ids, ","))
		}

		// Add pagination parameters
		if opts.Page > 0 {
			q.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			q.Set("page_size", strconv.Itoa(opts.PageSize))
		}
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	// Verbose logging
	if c.verbose {
		log.Printf("[API] GET %s", u.String())
	}

	// Apply rate limiting before making the request
	if c.rateLimiter != nil {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, nil, fmt.Errorf("rate limiting failed: %w", err)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Verbose response logging
	if c.verbose {
		log.Printf("[API] Response %d: %d bytes", resp.StatusCode, len(body))
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Text != "" {
			// For known API errors, wrap with status code for retry logic
			return nil, nil, HTTPError{
				StatusCode: resp.StatusCode,
				Message:    apiErr.Text,
			}
		}
		return nil, nil, HTTPError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	// Parse pagination headers if present
	var pagination *PaginationResponse
	if page := resp.Header.Get("X-Page"); page != "" {
		pagination = &PaginationResponse{}
		if p, err := strconv.Atoi(page); err == nil {
			pagination.Page = p
		}
		if ps := resp.Header.Get("X-Page-Size"); ps != "" {
			if p, err := strconv.Atoi(ps); err == nil {
				pagination.PageSize = p
			}
		}
		if pt := resp.Header.Get("X-Page-Total"); pt != "" {
			if p, err := strconv.Atoi(pt); err == nil {
				pagination.PageTotal = p
			}
		}
		if t := resp.Header.Get("X-Result-Total"); t != "" {
			if p, err := strconv.Atoi(t); err == nil {
				pagination.Total = p
			}
		}
	}

	return body, pagination, nil
}

// Generic helper functions

// GetIDs is a generic function to get a list of IDs from an endpoint
func GetIDs[T ~int](ctx context.Context, c *Client, endpoint string, options ...RequestOption) ([]T, error) {
	opts := &RequestOptions{}
	for _, opt := range options {
		opt(opts)
	}

	data, _, err := c.get(ctx, endpoint, opts)
	if err != nil {
		return nil, err
	}

	var ids []T
	if err := json.Unmarshal(data, &ids); err != nil {
		return nil, fmt.Errorf("failed to parse IDs: %w", err)
	}

	return ids, nil
}

// GetByID is a generic function to get a single item by ID
func GetByID[T any](ctx context.Context, c *Client, endpoint string, id int, options ...RequestOption) (*T, error) {
	opts := &RequestOptions{IDs: []int{id}}
	for _, opt := range options {
		opt(opts)
	}

	data, _, err := c.get(ctx, endpoint, opts)
	if err != nil {
		return nil, err
	}

	// Try to unmarshal as array first (bulk expansion response)
	var results []T
	if err := json.Unmarshal(data, &results); err == nil && len(results) > 0 {
		return &results[0], nil
	}

	// If that fails, try to unmarshal as single object
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetByIDs is a generic function to get multiple items by IDs
func GetByIDs[T any](ctx context.Context, c *Client, endpoint string, ids []int, options ...RequestOption) ([]T, error) {
	opts := &RequestOptions{IDs: ids}
	for _, opt := range options {
		opt(opts)
	}

	data, _, err := c.get(ctx, endpoint, opts)
	if err != nil {
		return nil, err
	}

	var results []T
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return results, nil
}

// GetAll is a generic function to get all items from an endpoint
func GetAll[T any](ctx context.Context, c *Client, endpoint string, options ...RequestOption) ([]T, error) {
	opts := &RequestOptions{All: true}
	for _, opt := range options {
		opt(opts)
	}

	data, _, err := c.get(ctx, endpoint, opts)
	if err != nil {
		return nil, err
	}

	var results []T
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return results, nil
}

// GetPaged is a generic function to get a page of items
func GetPaged[T any](ctx context.Context, c *Client, endpoint string, options ...RequestOption) ([]T, *PaginationResponse, error) {
	opts := &RequestOptions{}
	for _, opt := range options {
		opt(opts)
	}

	data, pagination, err := c.get(ctx, endpoint, opts)
	if err != nil {
		return nil, nil, err
	}

	var results []T
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return results, pagination, nil
}

// GetSingle is a generic function to get a single item (not by ID)
func GetSingle[T any](ctx context.Context, c *Client, endpoint string, options ...RequestOption) (*T, error) {
	opts := &RequestOptions{}
	for _, opt := range options {
		opt(opts)
	}

	data, _, err := c.get(ctx, endpoint, opts)
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// API Methods

// GetBuild returns the current game build ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/build
// Scopes: None (public endpoint)
func (c *Client) GetBuild(ctx context.Context, options ...RequestOption) (*Build, error) {
	return GetSingle[Build](ctx, c, "/v2/build", options...)
}

// GetBackstoryAnswerIDs returns all available backstory answer IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/backstory/answers
// Scopes: None (public endpoint)
func (c *Client) GetBackstoryAnswerIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/backstory/answers", options...)
}

// GetBackstoryAnswer returns details for a specific backstory answer.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/backstory/answers
// Scopes: None (public endpoint)
func (c *Client) GetBackstoryAnswer(ctx context.Context, id string, options ...RequestOption) (*BackstoryAnswer, error) {
	return GetSingle[BackstoryAnswer](ctx, c, "/v2/backstory/answers/"+id, options...)
}

// GetBackstoryQuestionIDs returns all available backstory question IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/backstory/questions
// Scopes: None (public endpoint)
func (c *Client) GetBackstoryQuestionIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/backstory/questions", options...)
}

// GetBackstoryQuestion returns details for a specific backstory question.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/backstory/questions
// Scopes: None (public endpoint)
func (c *Client) GetBackstoryQuestion(ctx context.Context, id string, options ...RequestOption) (*BackstoryQuestion, error) {
	return GetSingle[BackstoryQuestion](ctx, c, "/v2/backstory/questions/"+id, options...)
}

// GetBackstoryAnswers returns multiple backstory answers by IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/backstory/answers
// Scopes: None (public endpoint)
func (c *Client) GetBackstoryAnswers(ctx context.Context, ids []string, options ...RequestOption) ([]*BackstoryAnswer, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no IDs provided")
	}

	idsStr := strings.Join(ids, ",")
	endpoint := "/v2/backstory/answers?ids=" + idsStr

	results, err := GetAll[BackstoryAnswer](ctx, c, endpoint, options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*BackstoryAnswer, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// GetBackstoryQuestions returns multiple backstory questions by IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/backstory/questions
// Scopes: None (public endpoint)
func (c *Client) GetBackstoryQuestions(ctx context.Context, ids []string, options ...RequestOption) ([]*BackstoryQuestion, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no IDs provided")
	}

	idsStr := strings.Join(ids, ",")
	endpoint := "/v2/backstory/questions?ids=" + idsStr

	results, err := GetAll[BackstoryQuestion](ctx, c, endpoint, options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*BackstoryQuestion, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// GetAchievementIDs returns all available achievement IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/achievements
// Scopes: None (public endpoint)
func (c *Client) GetAchievementIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/achievements", options...)
}

// GetAchievement returns details for a specific achievement.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/achievements
// Scopes: None (public endpoint)
func (c *Client) GetAchievement(ctx context.Context, id int, options ...RequestOption) (*Achievement, error) {
	return GetByID[Achievement](ctx, c, "/v2/achievements", id, options...)
}

// GetAchievements returns multiple achievements by IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/achievements
// Scopes: None (public endpoint)
func (c *Client) GetAchievements(ctx context.Context, ids []int, options ...RequestOption) ([]*Achievement, error) {
	// Try cache first if available
	if c.dataCache != nil && c.dataCache.GetAchievementCache().IsLoaded() {
		cachedAchievements := c.dataCache.GetAchievementCache().GetByIDs(ids)
		if len(cachedAchievements) == len(ids) {
			// All achievements found in cache
			return cachedAchievements, nil
		}

		// Some achievements found in cache, determine which ones to fetch from API
		cachedMap := make(map[int]*Achievement)
		for _, achievement := range cachedAchievements {
			cachedMap[achievement.ID] = achievement
		}

		// Find missing IDs
		var missingIDs []int
		for _, id := range ids {
			if _, found := cachedMap[id]; !found {
				missingIDs = append(missingIDs, id)
			}
		}

		if len(missingIDs) == 0 {
			// All achievements were in cache
			result := make([]*Achievement, len(ids))
			for i, id := range ids {
				result[i] = cachedMap[id]
			}
			return result, nil
		}

		// Fetch missing achievements from API
		apiResults, err := GetByIDs[Achievement](ctx, c, "/v2/achievements", missingIDs, options...)
		if err != nil {
			// Return cached achievements even if API fails
			return cachedAchievements, nil
		}

		// Combine cached and API results
		for i := range apiResults {
			cachedMap[apiResults[i].ID] = &apiResults[i]
		}

		// Build result in original order
		result := make([]*Achievement, len(ids))
		for i, id := range ids {
			if achievement, found := cachedMap[id]; found {
				result[i] = achievement
			}
		}

		return result, nil
	}

	// No cache available, fetch directly from API
	results, err := GetByIDs[Achievement](ctx, c, "/v2/achievements", ids, options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*Achievement, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// GetCurrencyIDs returns all available currency IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/currencies
// Scopes: None (public endpoint)
func (c *Client) GetCurrencyIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/currencies", options...)
}

// GetCurrency returns details for a specific currency.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/currencies
// Scopes: None (public endpoint)
func (c *Client) GetCurrency(ctx context.Context, id int, options ...RequestOption) (*Currency, error) {
	return GetByID[Currency](ctx, c, "/v2/currencies", id, options...)
}

// GetCurrencies returns multiple currencies by IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/currencies
// Scopes: None (public endpoint)
func (c *Client) GetCurrencies(ctx context.Context, ids []int, options ...RequestOption) ([]*Currency, error) {
	results, err := GetByIDs[Currency](ctx, c, "/v2/currencies", ids, options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*Currency, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// GetItemIDs returns all available item IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/items
// Scopes: None (public endpoint)
func (c *Client) GetItemIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/items", options...)
}

// GetItem returns details for a specific item.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/items
// Scopes: None (public endpoint)
func (c *Client) GetItem(ctx context.Context, id int, options ...RequestOption) (*Item, error) {
	return GetByID[Item](ctx, c, "/v2/items", id, options...)
}

// GetItems returns multiple items by IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/items
// Scopes: None (public endpoint)
func (c *Client) GetItems(ctx context.Context, ids []int, options ...RequestOption) ([]*Item, error) {
	// Try cache first if available
	if c.dataCache != nil && c.dataCache.GetItemCache().IsLoaded() {
		cachedItems := c.dataCache.GetItemCache().GetByIDs(ids)
		if len(cachedItems) == len(ids) {
			// All items found in cache
			return cachedItems, nil
		}

		// Some items found in cache, determine which ones to fetch from API
		cachedMap := make(map[int]*Item)
		for _, item := range cachedItems {
			cachedMap[item.ID] = item
		}

		// Find missing IDs
		var missingIDs []int
		for _, id := range ids {
			if _, found := cachedMap[id]; !found {
				missingIDs = append(missingIDs, id)
			}
		}

		if len(missingIDs) == 0 {
			// All items were in cache
			result := make([]*Item, len(ids))
			for i, id := range ids {
				result[i] = cachedMap[id]
			}
			return result, nil
		}

		// Fetch missing items from API
		apiResults, err := GetByIDs[Item](ctx, c, "/v2/items", missingIDs, options...)
		if err != nil {
			// Return cached items even if API fails
			return cachedItems, nil
		}

		// Combine cached and API results
		for i := range apiResults {
			cachedMap[apiResults[i].ID] = &apiResults[i]
		}

		// Build result in original order
		result := make([]*Item, len(ids))
		for i, id := range ids {
			if item, found := cachedMap[id]; found {
				result[i] = item
			}
		}

		return result, nil
	}

	// Fallback to API only
	results, err := GetByIDs[Item](ctx, c, "/v2/items", ids, options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*Item, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// GetWorldIDs returns all available world IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/worlds
// Scopes: None (public endpoint)
func (c *Client) GetWorldIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/worlds", options...)
}

// GetWorld returns a specific world by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/worlds
// Scopes: None (public endpoint)
func (c *Client) GetWorld(ctx context.Context, id int, options ...RequestOption) (*World, error) {
	return GetByID[World](ctx, c, "/v2/worlds", id, options...)
}

// GetWorlds returns multiple worlds by IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/worlds
// Scopes: None (public endpoint)
func (c *Client) GetWorlds(ctx context.Context, ids []int, options ...RequestOption) ([]*World, error) {
	results, err := GetByIDs[World](ctx, c, "/v2/worlds", ids, options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*World, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// GetWorldsPage returns a page of worlds.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/worlds
// Scopes: None (public endpoint)
func (c *Client) GetWorldsPage(ctx context.Context, options ...RequestOption) ([]*World, *PaginationResponse, error) {
	results, pagination, err := GetPaged[World](ctx, c, "/v2/worlds", options...)
	if err != nil {
		return nil, nil, err
	}

	ptrs := make([]*World, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, pagination, nil
}

// GetCommercePriceIDs returns all available item IDs with trading post prices.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/commerce/prices
// Scopes: None (public endpoint)
func (c *Client) GetCommercePriceIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/commerce/prices", options...)
}

// GetCommercePrice returns trading post price information for a specific item.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/commerce/prices
// Scopes: None (public endpoint)
func (c *Client) GetCommercePrice(ctx context.Context, itemID int, options ...RequestOption) (*Price, error) {
	return GetByID[Price](ctx, c, "/v2/commerce/prices", itemID, options...)
}

// GetCommercePrices returns trading post price information for multiple items.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/commerce/prices
// Scopes: None (public endpoint)
func (c *Client) GetCommercePrices(ctx context.Context, itemIDs []int, options ...RequestOption) ([]*Price, error) {
	results, err := GetByIDs[Price](ctx, c, "/v2/commerce/prices", itemIDs, options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*Price, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// GetSkillIDs returns all available skill IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/skills
// Scopes: None (public endpoint)
func (c *Client) GetSkillIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/skills", options...)
}

// GetSkill returns a specific skill by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/skills
// Scopes: None (public endpoint)
func (c *Client) GetSkill(ctx context.Context, id int, options ...RequestOption) (*Skill, error) {
	return GetByID[Skill](ctx, c, "/v2/skills", id, options...)
}

// GetSkills returns multiple skills by IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/skills
// Scopes: None (public endpoint)
func (c *Client) GetSkills(ctx context.Context, ids []int, options ...RequestOption) ([]*Skill, error) {
	// Try cache first if available
	if c.dataCache != nil && c.dataCache.GetSkillCache().IsLoaded() {
		cachedSkills := c.dataCache.GetSkillCache().GetByIDs(ids)
		if len(cachedSkills) == len(ids) {
			// All skills found in cache
			return cachedSkills, nil
		}

		// Some skills found in cache, determine which ones to fetch from API
		cachedMap := make(map[int]*Skill)
		for _, skill := range cachedSkills {
			cachedMap[skill.ID] = skill
		}

		// Find missing IDs
		var missingIDs []int
		for _, id := range ids {
			if _, found := cachedMap[id]; !found {
				missingIDs = append(missingIDs, id)
			}
		}

		if len(missingIDs) == 0 {
			// All skills were in cache
			result := make([]*Skill, len(ids))
			for i, id := range ids {
				result[i] = cachedMap[id]
			}
			return result, nil
		}

		// Fetch missing skills from API
		apiResults, err := GetByIDs[Skill](ctx, c, "/v2/skills", missingIDs, options...)
		if err != nil {
			// Return cached skills even if API fails
			return cachedSkills, nil
		}

		// Combine cached and API results
		for i := range apiResults {
			cachedMap[apiResults[i].ID] = &apiResults[i]
		}

		// Build result in original order
		result := make([]*Skill, len(ids))
		for i, id := range ids {
			if skill, found := cachedMap[id]; found {
				result[i] = skill
			}
		}

		return result, nil
	}

	// Fallback to API only
	results, err := GetByIDs[Skill](ctx, c, "/v2/skills", ids, options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*Skill, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// GetAllCurrencies returns all currencies.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/currencies
// Scopes: None (public endpoint)
func (c *Client) GetAllCurrencies(ctx context.Context, options ...RequestOption) ([]*Currency, error) {
	results, err := GetAll[Currency](ctx, c, "/v2/currencies", options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*Currency, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// GetAllWorlds returns all worlds.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/worlds
// Scopes: None (public endpoint)
func (c *Client) GetAllWorlds(ctx context.Context, options ...RequestOption) ([]*World, error) {
	results, err := GetAll[World](ctx, c, "/v2/worlds", options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*World, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// ========================
// Account Endpoints
// ========================

// GetAccount returns basic account information.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account
// Scopes: account
// Optional Scopes: guilds, progression
func (c *Client) GetAccount(ctx context.Context, options ...RequestOption) (*Account, error) {
	return GetSingle[Account](ctx, c, "/v2/account", options...)
}

// GetAccountAchievements returns account's progress towards all achievements.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/achievements
// Scopes: account, progression
func (c *Client) GetAccountAchievements(ctx context.Context, options ...RequestOption) ([]AccountAchievement, error) {
	return GetAll[AccountAchievement](ctx, c, "/v2/account/achievements", options...)
}

// GetAccountBank returns items stored in the account vault.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/bank
// Scopes: account, inventories
func (c *Client) GetAccountBank(ctx context.Context, options ...RequestOption) ([]BankSlot, error) {
	return GetAll[BankSlot](ctx, c, "/v2/account/bank", options...)
}

// GetAccountBuildStorage returns build templates stored in the account.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/buildstorage
// Scopes: account
func (c *Client) GetAccountBuildStorage(ctx context.Context, options ...RequestOption) ([]BuildStorage, error) {
	return GetAll[BuildStorage](ctx, c, "/v2/account/buildstorage", options...)
}

// GetAccountDailyCrafting returns daily crafting progress.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/dailycrafting
// Scopes: account, progression
func (c *Client) GetAccountDailyCrafting(ctx context.Context, options ...RequestOption) ([]DailyCrafting, error) {
	return GetAll[DailyCrafting](ctx, c, "/v2/account/dailycrafting", options...)
}

// GetAccountDungeons returns dungeons completed since daily reset.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/dungeons
// Scopes: account, progression
func (c *Client) GetAccountDungeons(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/account/dungeons", options...)
}

// GetAccountDyes returns unlocked dyes.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/dyes
// Scopes: account, unlocks
func (c *Client) GetAccountDyes(ctx context.Context, options ...RequestOption) ([]Dye, error) {
	return GetAll[Dye](ctx, c, "/v2/account/dyes", options...)
}

// GetAccountEmotes returns unlocked emotes.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/emotes
// Scopes: account, unlocks
func (c *Client) GetAccountEmotes(ctx context.Context, options ...RequestOption) ([]Emote, error) {
	return GetAll[Emote](ctx, c, "/v2/account/emotes", options...)
}

// GetAccountFinishers returns unlocked finishers.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/finishers
// Scopes: account, unlocks
func (c *Client) GetAccountFinishers(ctx context.Context, options ...RequestOption) ([]Finisher, error) {
	return GetAll[Finisher](ctx, c, "/v2/account/finishers", options...)
}

// GetAccountGliders returns unlocked gliders.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/gliders
// Scopes: account, unlocks
func (c *Client) GetAccountGliders(ctx context.Context, options ...RequestOption) ([]Glider, error) {
	return GetAll[Glider](ctx, c, "/v2/account/gliders", options...)
}

// GetAccountHome returns home instance information.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/home
// Scopes: None (public endpoint)
func (c *Client) GetAccountHome(ctx context.Context, options ...RequestOption) (*HomeInfo, error) {
	return GetSingle[HomeInfo](ctx, c, "/v2/account/home", options...)
}

// GetAccountHomeCats returns unlocked home instance cats.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/home/cats
// Scopes: account, progression, unlocks
func (c *Client) GetAccountHomeCats(ctx context.Context, options ...RequestOption) ([]HomeCat, error) {
	return GetAll[HomeCat](ctx, c, "/v2/account/home/cats", options...)
}

// GetAccountHomeNodes returns unlocked home instance nodes.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/home/nodes
// Scopes: account, progression, unlocks
func (c *Client) GetAccountHomeNodes(ctx context.Context, options ...RequestOption) ([]HomeNode, error) {
	return GetAll[HomeNode](ctx, c, "/v2/account/home/nodes", options...)
}

// GetAccountHomestead returns homestead information.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/homestead
// Scopes: None (public endpoint)
func (c *Client) GetAccountHomestead(ctx context.Context, options ...RequestOption) (*Homestead, error) {
	return GetSingle[Homestead](ctx, c, "/v2/account/homestead", options...)
}

// GetAccountHomesteadDecorations returns homestead decorations used by the account.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/homestead/decorations
// Scopes: account, unlocks
func (c *Client) GetAccountHomesteadDecorations(ctx context.Context, options ...RequestOption) ([]HomesteadDecoration, error) {
	return GetAll[HomesteadDecoration](ctx, c, "/v2/account/homestead/decorations", options...)
}

// GetAccountHomesteadGlyphs returns glyphs stored in homestead collection boxes.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/homestead/glyphs
// Scopes: account, unlocks
func (c *Client) GetAccountHomesteadGlyphs(ctx context.Context, options ...RequestOption) ([]HomesteadGlyph, error) {
	return GetAll[HomesteadGlyph](ctx, c, "/v2/account/homestead/glyphs", options...)
}

// GetAccountInventory returns the shared inventory slots.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/inventory
// Scopes: account, inventories
func (c *Client) GetAccountInventory(ctx context.Context, options ...RequestOption) ([]InventorySlot, error) {
	return GetAll[InventorySlot](ctx, c, "/v2/account/inventory", options...)
}

// GetAccountJadeBots returns unlocked jade bot skins.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/jadebots
// Scopes: account, unlocks
func (c *Client) GetAccountJadeBots(ctx context.Context, options ...RequestOption) ([]JadeBot, error) {
	return GetAll[JadeBot](ctx, c, "/v2/account/jadebots", options...)
}

// GetAccountLegendaryArmory returns legendary armory items unlocked for the account.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/legendaryarmory
// Scopes: account, unlocks, inventories
func (c *Client) GetAccountLegendaryArmory(ctx context.Context, options ...RequestOption) ([]LegendaryArmory, error) {
	return GetAll[LegendaryArmory](ctx, c, "/v2/account/legendaryarmory", options...)
}

// GetAccountLuck returns the total amount of luck consumed on the account.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/luck
// Scopes: account, progression, unlocks
func (c *Client) GetAccountLuck(ctx context.Context, options ...RequestOption) ([]Luck, error) {
	return GetAll[Luck](ctx, c, "/v2/account/luck", options...)
}

// GetAccountMail returns account mail.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/mail
// Scopes: account
func (c *Client) GetAccountMail(ctx context.Context, options ...RequestOption) ([]Mail, error) {
	return GetAll[Mail](ctx, c, "/v2/account/mail", options...)
}

// GetAccountMailCarriers returns unlocked mail carriers.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/mailcarriers
// Scopes: account, unlocks
func (c *Client) GetAccountMailCarriers(ctx context.Context, options ...RequestOption) ([]MailCarrier, error) {
	return GetAll[MailCarrier](ctx, c, "/v2/account/mailcarriers", options...)
}

// GetAccountMapChests returns Hero's Choice Chests acquired since daily reset.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/mapchests
// Scopes: account, progression
func (c *Client) GetAccountMapChests(ctx context.Context, options ...RequestOption) ([]MapChest, error) {
	return GetAll[MapChest](ctx, c, "/v2/account/mapchests", options...)
}

// GetAccountMasteries returns unlocked masteries for the account.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/masteries
// Scopes: account, progression
func (c *Client) GetAccountMasteries(ctx context.Context, options ...RequestOption) ([]AccountMastery, error) {
	return GetAll[AccountMastery](ctx, c, "/v2/account/masteries", options...)
}

// GetAccountMasteryPoints returns the total amount of mastery points unlocked.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/mastery/points
// Scopes: account, progression
func (c *Client) GetAccountMasteryPoints(ctx context.Context, options ...RequestOption) ([]MasteryPoint, error) {
	return GetAll[MasteryPoint](ctx, c, "/v2/account/mastery/points", options...)
}

// GetAccountMaterials returns materials stored in the account vault.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/materials
// Scopes: account, inventories
func (c *Client) GetAccountMaterials(ctx context.Context, options ...RequestOption) ([]MaterialSlot, error) {
	return GetAll[MaterialSlot](ctx, c, "/v2/account/materials", options...)
}

// GetAccountMinis returns unlocked miniatures.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/minis
// Scopes: account, unlocks
func (c *Client) GetAccountMinis(ctx context.Context, options ...RequestOption) ([]Mini, error) {
	return GetAll[Mini](ctx, c, "/v2/account/minis", options...)
}

// GetAccountMounts returns mount information.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/mounts
// Scopes: None (public endpoint)
func (c *Client) GetAccountMounts(ctx context.Context, options ...RequestOption) (*MountInfo, error) {
	return GetSingle[MountInfo](ctx, c, "/v2/account/mounts", options...)
}

// GetAccountMountSkins returns unlocked mount skins.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/mounts/skins
// Scopes: account, unlocks
func (c *Client) GetAccountMountSkins(ctx context.Context, options ...RequestOption) ([]MountSkin, error) {
	return GetAll[MountSkin](ctx, c, "/v2/account/mounts/skins", options...)
}

// GetAccountMountTypes returns unlocked mount types.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/mounts/types
// Scopes: account, unlocks
func (c *Client) GetAccountMountTypes(ctx context.Context, options ...RequestOption) ([]MountType, error) {
	return GetAll[MountType](ctx, c, "/v2/account/mounts/types", options...)
}

// GetAccountNovelties returns unlocked novelties.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/novelties
// Scopes: account, unlocks
func (c *Client) GetAccountNovelties(ctx context.Context, options ...RequestOption) ([]Novelty, error) {
	return GetAll[Novelty](ctx, c, "/v2/account/novelties", options...)
}

// GetAccountOutfits returns unlocked outfits.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/outfits
// Scopes: account, unlocks
func (c *Client) GetAccountOutfits(ctx context.Context, options ...RequestOption) ([]Outfit, error) {
	return GetAll[Outfit](ctx, c, "/v2/account/outfits", options...)
}

// GetAccountProgression returns account-wide progression for Fractals and Luck.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/progression
// Scopes: progression, unlocks
func (c *Client) GetAccountProgression(ctx context.Context, options ...RequestOption) ([]Progression, error) {
	return GetAll[Progression](ctx, c, "/v2/account/progression", options...)
}

// GetAccountPvPHeroes returns unlocked PvP heroes.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/pvp/heroes
// Scopes: account, unlocks
func (c *Client) GetAccountPvPHeroes(ctx context.Context, options ...RequestOption) ([]AccountPvPHero, error) {
	return GetAll[AccountPvPHero](ctx, c, "/v2/account/pvp/heroes", options...)
}

// GetAccountRaids returns completed raid encounters since weekly reset.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/raids
// Scopes: account, progression
func (c *Client) GetAccountRaids(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/account/raids", options...)
}

// GetAccountRecipes returns unlocked recipes.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/recipes
// Scopes: account, unlocks
func (c *Client) GetAccountRecipes(ctx context.Context, options ...RequestOption) ([]Recipe, error) {
	return GetAll[Recipe](ctx, c, "/v2/account/recipes", options...)
}

// GetAccountSkiffs returns unlocked skiff skins.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/skiffs
// Scopes: account, unlocks
func (c *Client) GetAccountSkiffs(ctx context.Context, options ...RequestOption) ([]Skiff, error) {
	return GetAll[Skiff](ctx, c, "/v2/account/skiffs", options...)
}

// GetAccountSkins returns unlocked skins.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/skins
// Scopes: account, unlocks
func (c *Client) GetAccountSkins(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/account/skins", options...)
}

// GetAccountTitles returns unlocked titles.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/titles
// Scopes: account, unlocks
func (c *Client) GetAccountTitles(ctx context.Context, options ...RequestOption) ([]UnlockedTitle, error) {
	return GetAll[UnlockedTitle](ctx, c, "/v2/account/titles", options...)
}

// GetAccountWallet returns the account's currencies.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/wallet
// Scopes: account, wallet
func (c *Client) GetAccountWallet(ctx context.Context, options ...RequestOption) ([]WalletCurrency, error) {
	return GetAll[WalletCurrency](ctx, c, "/v2/account/wallet", options...)
}

// GetAccountWizardsVaultDaily returns daily Wizard's Vault objectives.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/wizardsvault/daily
// Scopes: account, progression
func (c *Client) GetAccountWizardsVaultDaily(ctx context.Context, options ...RequestOption) ([]WizardsVaultDaily, error) {
	return GetAll[WizardsVaultDaily](ctx, c, "/v2/account/wizardsvault/daily", options...)
}

// GetAccountWizardsVaultListings returns Wizard's Vault reward listings.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/wizardsvault/listings
// Scopes: account, progression
func (c *Client) GetAccountWizardsVaultListings(ctx context.Context, options ...RequestOption) ([]WizardsVaultListing, error) {
	return GetAll[WizardsVaultListing](ctx, c, "/v2/account/wizardsvault/listings", options...)
}

// GetAccountWizardsVaultSpecial returns special Wizard's Vault objectives.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/wizardsvault/special
// Scopes: account, progression
func (c *Client) GetAccountWizardsVaultSpecial(ctx context.Context, options ...RequestOption) ([]WizardsVaultSpecial, error) {
	return GetAll[WizardsVaultSpecial](ctx, c, "/v2/account/wizardsvault/special", options...)
}

// GetAccountWizardsVaultWeekly returns weekly Wizard's Vault objectives.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/wizardsvault/weekly
// Scopes: account, progression
func (c *Client) GetAccountWizardsVaultWeekly(ctx context.Context, options ...RequestOption) ([]WizardsVaultWeekly, error) {
	return GetAll[WizardsVaultWeekly](ctx, c, "/v2/account/wizardsvault/weekly", options...)
}

// GetAccountWorldBosses returns defeated world bosses since daily reset.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/worldbosses
// Scopes: account, progression
func (c *Client) GetAccountWorldBosses(ctx context.Context, options ...RequestOption) ([]WorldBoss, error) {
	return GetAll[WorldBoss](ctx, c, "/v2/account/worldbosses", options...)
}

// GetAccountWvW returns WvW account information.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/account/wvw
// Scopes: account
func (c *Client) GetAccountWvW(ctx context.Context, options ...RequestOption) (*WvWInfo, error) {
	return GetSingle[WvWInfo](ctx, c, "/v2/account/wvw", options...)
}

// ========================
// Achievement Categories & Groups
// ========================

// GetAchievementCategoryIDs returns all achievement category IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/achievements/categories
// Scopes: None (public endpoint)
func (c *Client) GetAchievementCategoryIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/achievements/categories", options...)
}

// GetAchievementCategory returns details for a specific achievement category.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/achievements/categories
// Scopes: None (public endpoint)
func (c *Client) GetAchievementCategory(ctx context.Context, id int, options ...RequestOption) (*AchievementCategory, error) {
	return GetByID[AchievementCategory](ctx, c, "/v2/achievements/categories", id, options...)
}

// GetAchievementCategories returns multiple achievement categories by IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/achievements/categories
// Scopes: None (public endpoint)
func (c *Client) GetAchievementCategories(ctx context.Context, ids []int, options ...RequestOption) ([]*AchievementCategory, error) {
	results, err := GetByIDs[AchievementCategory](ctx, c, "/v2/achievements/categories", ids, options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*AchievementCategory, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// GetAchievementGroupIDs returns all achievement group IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/achievements/groups
// Scopes: None (public endpoint)
func (c *Client) GetAchievementGroupIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/achievements/groups", options...)
}

// GetAchievementGroup returns details for a specific achievement group.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/achievements/groups
// Scopes: None (public endpoint)
func (c *Client) GetAchievementGroup(ctx context.Context, id string, options ...RequestOption) (*AchievementGroup, error) {
	// Note: This would need a string-based GetByID variant
	return GetSingle[AchievementGroup](ctx, c, "/v2/achievements/groups/"+id, options...)
}

// GetDailyAchievements returns today's daily achievements.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/achievements/daily
// Scopes: None (public endpoint)
func (c *Client) GetDailyAchievements(ctx context.Context, options ...RequestOption) (*DailyAchievements, error) {
	return GetSingle[DailyAchievements](ctx, c, "/v2/achievements/daily", options...)
}

// GetDailyAchievementsTomorrow returns tomorrow's daily achievements.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/achievements/daily/tomorrow
// Scopes: None (public endpoint)
func (c *Client) GetDailyAchievementsTomorrow(ctx context.Context, options ...RequestOption) (*DailyAchievements, error) {
	return GetSingle[DailyAchievements](ctx, c, "/v2/achievements/daily/tomorrow", options...)
}

// ========================
// Missing General Endpoints
// ========================

// GetColorIDs returns all available color IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/colors
// Scopes: None (public endpoint)
func (c *Client) GetColorIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/colors", options...)
}

// GetColor returns details for a specific color.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/colors
// Scopes: None (public endpoint)
func (c *Client) GetColor(ctx context.Context, id int, options ...RequestOption) (*Color, error) {
	return GetByID[Color](ctx, c, "/v2/colors", id, options...)
}

// GetColors returns multiple colors by IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/colors
// Scopes: None (public endpoint)
func (c *Client) GetColors(ctx context.Context, ids []int, options ...RequestOption) ([]*Color, error) {
	results, err := GetByIDs[Color](ctx, c, "/v2/colors", ids, options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*Color, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// GetCommerceListings returns all current trading post listings.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/commerce/listings
// Scopes: None (public endpoint)
func (c *Client) GetCommerceListings(ctx context.Context, options ...RequestOption) ([]Listing, error) {
	return GetAll[Listing](ctx, c, "/v2/commerce/listings", options...)
}

// GetCommerceExchange returns gem to gold exchange rates.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/commerce/exchange
// Scopes: None (public endpoint)
func (c *Client) GetCommerceExchange(ctx context.Context, options ...RequestOption) (*ExchangeRate, error) {
	return GetSingle[ExchangeRate](ctx, c, "/v2/commerce/exchange", options...)
}

// GetCommerceDelivery returns items available for pickup from trading post.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/commerce/delivery
// Scopes: account, tradingpost
func (c *Client) GetCommerceDelivery(ctx context.Context, options ...RequestOption) (*DeliveryItem, error) {
	return GetSingle[DeliveryItem](ctx, c, "/v2/commerce/delivery", options...)
}

// GetCommerceTransactions returns trading post transaction history.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/commerce/transactions
// Scopes: account, tradingpost
func (c *Client) GetCommerceTransactions(ctx context.Context, options ...RequestOption) ([]Transaction, error) {
	return GetAll[Transaction](ctx, c, "/v2/commerce/transactions", options...)
}

// GetContinentIDs returns all continent IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/continents
// Scopes: None (public endpoint)
func (c *Client) GetContinentIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/continents", options...)
}

// GetContinent returns a specific continent by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/continents
// Scopes: None (public endpoint)
func (c *Client) GetContinent(ctx context.Context, id int, options ...RequestOption) (*Continent, error) {
	return GetByID[Continent](ctx, c, "/v2/continents", id, options...)
}

// GetCreateSubtoken creates a subtoken.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/createsubtoken
// Scopes: account
func (c *Client) GetCreateSubtoken(ctx context.Context, options ...RequestOption) (*CreateSubtoken, error) {
	return GetSingle[CreateSubtoken](ctx, c, "/v2/createsubtoken", options...)
}

// GetDailyCrafting returns daily crafting items.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/dailycrafting
// Scopes: None (public endpoint)
func (c *Client) GetDailyCrafting(ctx context.Context, options ...RequestOption) ([]DailyCraftingItem, error) {
	return GetAll[DailyCraftingItem](ctx, c, "/v2/dailycrafting", options...)
}

// GetDungeonIDs returns all dungeon IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/dungeons
// Scopes: None (public endpoint)
func (c *Client) GetDungeonIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/dungeons", options...)
}

// GetDungeon returns a specific dungeon by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/dungeons
// Scopes: None (public endpoint)
func (c *Client) GetDungeon(ctx context.Context, id string, options ...RequestOption) (*Dungeon, error) {
	return GetSingle[Dungeon](ctx, c, "/v2/dungeons/"+id, options...)
}

// GetEmblem returns emblem information.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/emblem
// Scopes: None (public endpoint)
func (c *Client) GetEmblem(ctx context.Context, options ...RequestOption) (*Emblem, error) {
	return GetSingle[Emblem](ctx, c, "/v2/emblem", options...)
}

// GetEmoteIDs returns all emote IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/emotes
// Scopes: None (public endpoint)
func (c *Client) GetEmoteIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/emotes", options...)
}

// GetEmoteDetail returns a specific emote by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/emotes
// Scopes: None (public endpoint)
func (c *Client) GetEmoteDetail(ctx context.Context, id string, options ...RequestOption) (*EmoteDetail, error) {
	return GetSingle[EmoteDetail](ctx, c, "/v2/emotes/"+id, options...)
}

// GetEventIDs returns all event IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/events
// Scopes: None (public endpoint)
func (c *Client) GetEventIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/events", options...)
}

// GetEvent returns a specific event by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/events
// Scopes: None (public endpoint)
func (c *Client) GetEvent(ctx context.Context, id string, options ...RequestOption) (*Event, error) {
	return GetSingle[Event](ctx, c, "/v2/events/"+id, options...)
}

// GetFileIDs returns all file IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/files
// Scopes: None (public endpoint)
func (c *Client) GetFileIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/files", options...)
}

// GetFileDetail returns a specific file by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/files
// Scopes: None (public endpoint)
func (c *Client) GetFileDetail(ctx context.Context, id string, options ...RequestOption) (*FileDetail, error) {
	return GetSingle[FileDetail](ctx, c, "/v2/files/"+id, options...)
}

// GetFinisherIDs returns all finisher IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/finishers
// Scopes: None (public endpoint)
func (c *Client) GetFinisherIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/finishers", options...)
}

// GetFinisher returns a specific finisher by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/finishers
// Scopes: None (public endpoint)
func (c *Client) GetFinisher(ctx context.Context, id int, options ...RequestOption) (*Finisher, error) {
	return GetByID[Finisher](ctx, c, "/v2/finishers", id, options...)
}

// GetFinishers returns multiple finishers by IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/finishers
// Scopes: None (public endpoint)
func (c *Client) GetFinishers(ctx context.Context, ids []int, options ...RequestOption) ([]*Finisher, error) {
	results, err := GetByIDs[Finisher](ctx, c, "/v2/finishers", ids, options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*Finisher, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// GetGliderIDs returns all glider IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/gliders
// Scopes: None (public endpoint)
func (c *Client) GetGliderIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/gliders", options...)
}

// GetGlider returns a specific glider by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/gliders
// Scopes: None (public endpoint)
func (c *Client) GetGlider(ctx context.Context, id int, options ...RequestOption) (*Glider, error) {
	return GetByID[Glider](ctx, c, "/v2/gliders", id, options...)
}

// GetGliders returns multiple gliders by IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/gliders
// Scopes: None (public endpoint)
func (c *Client) GetGliders(ctx context.Context, ids []int, options ...RequestOption) ([]*Glider, error) {
	results, err := GetByIDs[Glider](ctx, c, "/v2/gliders", ids, options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*Glider, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// GetHome returns home instance information.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/home
// Scopes: None (public endpoint)
func (c *Client) GetHome(ctx context.Context, options ...RequestOption) (*HomeInfo, error) {
	return GetSingle[HomeInfo](ctx, c, "/v2/home", options...)
}

// GetHomeCats returns home instance cats.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/home/cats
// Scopes: None (public endpoint)
func (c *Client) GetHomeCats(ctx context.Context, options ...RequestOption) ([]HomeCat, error) {
	return GetAll[HomeCat](ctx, c, "/v2/home/cats", options...)
}

// GetHomeNodes returns home instance nodes.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/home/nodes
// Scopes: None (public endpoint)
func (c *Client) GetHomeNodes(ctx context.Context, options ...RequestOption) ([]HomeNode, error) {
	return GetAll[HomeNode](ctx, c, "/v2/home/nodes", options...)
}

// GetHomestead returns homestead information.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/homestead
// Scopes: None (public endpoint)
func (c *Client) GetHomestead(ctx context.Context, options ...RequestOption) (*HomesteadInfo, error) {
	return GetSingle[HomesteadInfo](ctx, c, "/v2/homestead", options...)
}

// GetHomesteadDecorationIDs returns all homestead decoration IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/homestead/decorations
// Scopes: None (public endpoint)
func (c *Client) GetHomesteadDecorationIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/homestead/decorations", options...)
}

// GetHomesteadDecoration returns a specific homestead decoration by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/homestead/decorations
// Scopes: None (public endpoint)
func (c *Client) GetHomesteadDecoration(ctx context.Context, id int, options ...RequestOption) (*HomesteadDecorationDetail, error) {
	return GetByID[HomesteadDecorationDetail](ctx, c, "/v2/homestead/decorations", id, options...)
}

// GetHomesteadDecorationCategoryIDs returns all decoration category IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/homestead/decorations/categories
// Scopes: None (public endpoint)
func (c *Client) GetHomesteadDecorationCategoryIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/homestead/decorations/categories", options...)
}

// GetHomesteadDecorationCategory returns a specific decoration category by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/homestead/decorations/categories
// Scopes: None (public endpoint)
func (c *Client) GetHomesteadDecorationCategory(ctx context.Context, id int, options ...RequestOption) (*HomesteadDecorationCategory, error) {
	return GetByID[HomesteadDecorationCategory](ctx, c, "/v2/homestead/decorations/categories", id, options...)
}

// GetHomesteadGlyphIDs returns all homestead glyph IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/homestead/glyphs
// Scopes: None (public endpoint)
func (c *Client) GetHomesteadGlyphIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/homestead/glyphs", options...)
}

// GetHomesteadGlyph returns a specific homestead glyph by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/homestead/glyphs
// Scopes: None (public endpoint)
func (c *Client) GetHomesteadGlyph(ctx context.Context, id int, options ...RequestOption) (*HomesteadGlyphDetail, error) {
	return GetByID[HomesteadGlyphDetail](ctx, c, "/v2/homestead/glyphs", id, options...)
}

// GetItemStatIDs returns all item stat IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/itemstats
// Scopes: None (public endpoint)
func (c *Client) GetItemStatIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/itemstats", options...)
}

// GetItemStat returns a specific item stat by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/itemstats
// Scopes: None (public endpoint)
func (c *Client) GetItemStat(ctx context.Context, id int, options ...RequestOption) (*ItemStat, error) {
	return GetByID[ItemStat](ctx, c, "/v2/itemstats", id, options...)
}

// GetJadeBotIDs returns all jade bot IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/jadebots
// Scopes: None (public endpoint)
func (c *Client) GetJadeBotIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/jadebots", options...)
}

// GetJadeBot returns a specific jade bot by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/jadebots
// Scopes: None (public endpoint)
func (c *Client) GetJadeBot(ctx context.Context, id int, options ...RequestOption) (*JadeBot, error) {
	return GetByID[JadeBot](ctx, c, "/v2/jadebots", id, options...)
}

// GetLegendaryArmory returns legendary armory information.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/legendaryarmory
// Scopes: None (public endpoint)
func (c *Client) GetLegendaryArmory(ctx context.Context, options ...RequestOption) (*LegendaryArmoryDetail, error) {
	return GetSingle[LegendaryArmoryDetail](ctx, c, "/v2/legendaryarmory", options...)
}

// GetLegendIDs returns all legend IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/legends
// Scopes: None (public endpoint)
func (c *Client) GetLegendIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/legends", options...)
}

// GetLegend returns a specific legend by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/legends
// Scopes: None (public endpoint)
func (c *Client) GetLegend(ctx context.Context, id string, options ...RequestOption) (*Legend, error) {
	return GetSingle[Legend](ctx, c, "/v2/legends/"+id, options...)
}

// GetLogos returns logo information.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/logos
// Scopes: None (public endpoint)
func (c *Client) GetLogos(ctx context.Context, options ...RequestOption) (*LogoDetail, error) {
	return GetSingle[LogoDetail](ctx, c, "/v2/logos", options...)
}

// GetMailCarrierIDs returns all mail carrier IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/mailcarriers
// Scopes: None (public endpoint)
func (c *Client) GetMailCarrierIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/mailcarriers", options...)
}

// GetMailCarrier returns a specific mail carrier by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/mailcarriers
// Scopes: None (public endpoint)
func (c *Client) GetMailCarrier(ctx context.Context, id int, options ...RequestOption) (*MailCarrier, error) {
	return GetByID[MailCarrier](ctx, c, "/v2/mailcarriers", id, options...)
}

// GetMapChests returns map chest information.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/mapchests
// Scopes: None (public endpoint)
func (c *Client) GetMapChests(ctx context.Context, options ...RequestOption) ([]MapChest, error) {
	return GetAll[MapChest](ctx, c, "/v2/mapchests", options...)
}

// GetMapIDs returns all map IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/maps
// Scopes: None (public endpoint)
func (c *Client) GetMapIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/maps", options...)
}

// GetMap returns a specific map by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/maps
// Scopes: None (public endpoint)
func (c *Client) GetMap(ctx context.Context, id int, options ...RequestOption) (*MapDetail, error) {
	return GetByID[MapDetail](ctx, c, "/v2/maps", id, options...)
}

// GetMaps returns multiple maps by IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/maps
// Scopes: None (public endpoint)
func (c *Client) GetMaps(ctx context.Context, ids []int, options ...RequestOption) ([]*MapDetail, error) {
	results, err := GetByIDs[MapDetail](ctx, c, "/v2/maps", ids, options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*MapDetail, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// GetMasteryIDs returns all mastery IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/masteries
// Scopes: None (public endpoint)
func (c *Client) GetMasteryIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/masteries", options...)
}

// GetMastery returns a specific mastery by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/masteries
// Scopes: None (public endpoint)
func (c *Client) GetMastery(ctx context.Context, id int, options ...RequestOption) (*Mastery, error) {
	return GetByID[Mastery](ctx, c, "/v2/masteries", id, options...)
}

// GetMaterialIDs returns all material category IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/materials
// Scopes: None (public endpoint)
func (c *Client) GetMaterialIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/materials", options...)
}

// GetMaterial returns a specific material category by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/materials
// Scopes: None (public endpoint)
func (c *Client) GetMaterial(ctx context.Context, id int, options ...RequestOption) (*Material, error) {
	return GetByID[Material](ctx, c, "/v2/materials", id, options...)
}

// GetMiniIDs returns all mini IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/minis
// Scopes: None (public endpoint)
func (c *Client) GetMiniIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/minis", options...)
}

// GetMini returns a specific mini by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/minis
// Scopes: None (public endpoint)
func (c *Client) GetMini(ctx context.Context, id int, options ...RequestOption) (*Mini, error) {
	return GetByID[Mini](ctx, c, "/v2/minis", id, options...)
}

// GetMounts returns mount information.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/mounts
// Scopes: None (public endpoint)
func (c *Client) GetMounts(ctx context.Context, options ...RequestOption) (*MountInfo, error) {
	return GetSingle[MountInfo](ctx, c, "/v2/mounts", options...)
}

// GetMountSkinIDs returns all mount skin IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/mounts/skins
// Scopes: None (public endpoint)
func (c *Client) GetMountSkinIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/mounts/skins", options...)
}

// GetMountSkin returns a specific mount skin by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/mounts/skins
// Scopes: None (public endpoint)
func (c *Client) GetMountSkin(ctx context.Context, id int, options ...RequestOption) (*MountSkinDetail, error) {
	return GetByID[MountSkinDetail](ctx, c, "/v2/mounts/skins", id, options...)
}

// GetMountTypeIDs returns all mount type IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/mounts/types
// Scopes: None (public endpoint)
func (c *Client) GetMountTypeIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/mounts/types", options...)
}

// GetMountType returns a specific mount type by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/mounts/types
// Scopes: None (public endpoint)
func (c *Client) GetMountType(ctx context.Context, id string, options ...RequestOption) (*MountTypeDetail, error) {
	return GetSingle[MountTypeDetail](ctx, c, "/v2/mounts/types/"+id, options...)
}

// GetNoveltyIDs returns all novelty IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/novelties
// Scopes: None (public endpoint)
func (c *Client) GetNoveltyIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/novelties", options...)
}

// GetNovelty returns a specific novelty by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/novelties
// Scopes: None (public endpoint)
func (c *Client) GetNovelty(ctx context.Context, id int, options ...RequestOption) (*NoveltyDetail, error) {
	return GetByID[NoveltyDetail](ctx, c, "/v2/novelties", id, options...)
}

// GetOutfitIDs returns all outfit IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/outfits
// Scopes: None (public endpoint)
func (c *Client) GetOutfitIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/outfits", options...)
}

// GetOutfit returns a specific outfit by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/outfits
// Scopes: None (public endpoint)
func (c *Client) GetOutfit(ctx context.Context, id int, options ...RequestOption) (*OutfitDetail, error) {
	return GetByID[OutfitDetail](ctx, c, "/v2/outfits", id, options...)
}

// GetPetIDs returns all pet IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pets
// Scopes: None (public endpoint)
func (c *Client) GetPetIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/pets", options...)
}

// GetPet returns a specific pet by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pets
// Scopes: None (public endpoint)
func (c *Client) GetPet(ctx context.Context, id int, options ...RequestOption) (*Pet, error) {
	return GetByID[Pet](ctx, c, "/v2/pets", id, options...)
}

// GetPets returns multiple pets by IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pets
// Scopes: None (public endpoint)
func (c *Client) GetPets(ctx context.Context, ids []int, options ...RequestOption) ([]*Pet, error) {
	results, err := GetByIDs[Pet](ctx, c, "/v2/pets", ids, options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*Pet, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// GetProfessionIDs returns all profession IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/professions
// Scopes: None (public endpoint)
func (c *Client) GetProfessionIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/professions", options...)
}

// GetProfession returns a specific profession by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/professions
// Scopes: None (public endpoint)
func (c *Client) GetProfession(ctx context.Context, id string, options ...RequestOption) (*Profession, error) {
	return GetSingle[Profession](ctx, c, "/v2/professions/"+id, options...)
}

// GetQuagganIDs returns all quaggan IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/quaggans
// Scopes: None (public endpoint)
func (c *Client) GetQuagganIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/quaggans", options...)
}

// GetQuaggan returns a specific quaggan by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/quaggans
// Scopes: None (public endpoint)
func (c *Client) GetQuaggan(ctx context.Context, id string, options ...RequestOption) (*Quaggan, error) {
	return GetSingle[Quaggan](ctx, c, "/v2/quaggans/"+id, options...)
}

// GetQuestIDs returns all quest IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/quests
// Scopes: None (public endpoint)
func (c *Client) GetQuestIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/quests", options...)
}

// GetQuest returns a specific quest by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/quests
// Scopes: None (public endpoint)
func (c *Client) GetQuest(ctx context.Context, id int, options ...RequestOption) (*Quest, error) {
	return GetByID[Quest](ctx, c, "/v2/quests", id, options...)
}

// GetRaceIDs returns all race IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/races
// Scopes: None (public endpoint)
func (c *Client) GetRaceIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/races", options...)
}

// GetRace returns a specific race by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/races
// Scopes: None (public endpoint)
func (c *Client) GetRace(ctx context.Context, id string, options ...RequestOption) (*Race, error) {
	return GetSingle[Race](ctx, c, "/v2/races/"+id, options...)
}

// GetRaidIDs returns all raid IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/raids
// Scopes: None (public endpoint)
func (c *Client) GetRaidIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/raids", options...)
}

// GetRaid returns a specific raid by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/raids
// Scopes: None (public endpoint)
func (c *Client) GetRaid(ctx context.Context, id string, options ...RequestOption) (*Raid, error) {
	return GetSingle[Raid](ctx, c, "/v2/raids/"+id, options...)
}

// GetRecipeIDs returns all recipe IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/recipes
// Scopes: None (public endpoint)
func (c *Client) GetRecipeIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/recipes", options...)
}

// GetRecipes returns a specific recipe by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/recipes
// Scopes: None (public endpoint)
func (c *Client) GetRecipes(ctx context.Context, ids []int, options ...RequestOption) ([]*RecipeDetail, error) {
	// Try cache first if available
	if c.dataCache != nil && c.dataCache.GetRecipeCache().IsLoaded() {
		cachedRecipes := c.dataCache.GetRecipeCache().GetByIDs(ids)
		if len(cachedRecipes) == len(ids) {
			// All recipes found in cache
			return cachedRecipes, nil
		}

		// Some recipes found in cache, determine which ones to fetch from API
		cachedMap := make(map[int]*RecipeDetail)
		for _, recipe := range cachedRecipes {
			cachedMap[recipe.ID] = recipe
		}

		// Find missing IDs
		var missingIDs []int
		for _, id := range ids {
			if _, found := cachedMap[id]; !found {
				missingIDs = append(missingIDs, id)
			}
		}

		if len(missingIDs) == 0 {
			// All recipes were in cache
			result := make([]*RecipeDetail, len(ids))
			for i, id := range ids {
				result[i] = cachedMap[id]
			}
			return result, nil
		}

		// Fetch missing recipes from API
		apiResults, err := GetByIDs[RecipeDetail](ctx, c, "/v2/recipes", missingIDs, options...)
		if err != nil {
			// Return cached recipes even if API fails
			return cachedRecipes, nil
		}

		// Combine cached and API results
		for i := range apiResults {
			cachedMap[apiResults[i].ID] = &apiResults[i]
		}

		// Build result in original order
		result := make([]*RecipeDetail, len(ids))
		for i, id := range ids {
			if recipe, found := cachedMap[id]; found {
				result[i] = recipe
			}
		}

		return result, nil
	}

	// Fallback to API only
	results, err := GetByIDs[RecipeDetail](ctx, c, "/v2/recipes", ids, options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*RecipeDetail, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// GetRecipeSearch returns recipe search functionality.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/recipes/search
// Scopes: None (public endpoint)
func (c *Client) GetRecipeSearch(ctx context.Context, options ...RequestOption) (*RecipeSearch, error) {
	return GetSingle[RecipeSearch](ctx, c, "/v2/recipes/search", options...)
}

// GetSkiffIDs returns all skiff IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/skiffs
// Scopes: None (public endpoint)
func (c *Client) GetSkiffIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/skiffs", options...)
}

// GetSkiff returns a specific skiff by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/skiffs
// Scopes: None (public endpoint)
func (c *Client) GetSkiff(ctx context.Context, id int, options ...RequestOption) (*SkiffDetail, error) {
	return GetByID[SkiffDetail](ctx, c, "/v2/skiffs", id, options...)
}

// GetSkinIDs returns all skin IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/skins
// Scopes: None (public endpoint)
func (c *Client) GetSkinIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/skins", options...)
}

// GetSkins returns multiple skins by IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/skins
// Scopes: None (public endpoint)
func (c *Client) GetSkins(ctx context.Context, ids []int, options ...RequestOption) ([]*SkinDetail, error) {
	results, err := GetByIDs[SkinDetail](ctx, c, "/v2/skins", ids, options...)
	if err != nil {
		return nil, err
	}

	ptrs := make([]*SkinDetail, len(results))
	for i := range results {
		ptrs[i] = &results[i]
	}
	return ptrs, nil
}

// GetSkin returns a specific skin by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/skins
// Scopes: None (public endpoint)
func (c *Client) GetSkin(ctx context.Context, id int, options ...RequestOption) (*SkinDetail, error) {
	return GetByID[SkinDetail](ctx, c, "/v2/skins", id, options...)
}

// GetSpecializationIDs returns all specialization IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/specializations
// Scopes: None (public endpoint)
func (c *Client) GetSpecializationIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/specializations", options...)
}

// GetSpecialization returns a specific specialization by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/specializations
// Scopes: None (public endpoint)
func (c *Client) GetSpecialization(ctx context.Context, id int, options ...RequestOption) (*Specialization, error) {
	return GetByID[Specialization](ctx, c, "/v2/specializations", id, options...)
}

// GetStoryIDs returns all story IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/stories
// Scopes: None (public endpoint)
func (c *Client) GetStoryIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/stories", options...)
}

// GetStory returns a specific story by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/stories
// Scopes: None (public endpoint)
func (c *Client) GetStory(ctx context.Context, id int, options ...RequestOption) (*Story, error) {
	return GetByID[Story](ctx, c, "/v2/stories", id, options...)
}

// GetStorySeasonIDs returns all story season IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/stories/seasons
// Scopes: None (public endpoint)
func (c *Client) GetStorySeasonIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/stories/seasons", options...)
}

// GetStorySeason returns a specific story season by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/stories/seasons
// Scopes: None (public endpoint)
func (c *Client) GetStorySeason(ctx context.Context, id string, options ...RequestOption) (*StorySeason, error) {
	return GetSingle[StorySeason](ctx, c, "/v2/stories/seasons/"+id, options...)
}

// GetTitleIDs returns all title IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/titles
// Scopes: None (public endpoint)
func (c *Client) GetTitleIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/titles", options...)
}

// GetTitle returns a specific title by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/titles
// Scopes: None (public endpoint)
func (c *Client) GetTitle(ctx context.Context, id int, options ...RequestOption) (*Title, error) {
	return GetByID[Title](ctx, c, "/v2/titles", id, options...)
}

// GetTokenInfo returns API token information.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/tokeninfo
// Scopes: account
func (c *Client) GetTokenInfo(ctx context.Context, options ...RequestOption) (*TokenInfo, error) {
	return GetSingle[TokenInfo](ctx, c, "/v2/tokeninfo", options...)
}

// GetTraitIDs returns all trait IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/traits
// Scopes: None (public endpoint)
func (c *Client) GetTraitIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/traits", options...)
}

// GetTrait returns a specific trait by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/traits
// Scopes: None (public endpoint)
func (c *Client) GetTrait(ctx context.Context, id int, options ...RequestOption) (*Trait, error) {
	return GetByID[Trait](ctx, c, "/v2/traits", id, options...)
}

// GetVendorIDs returns all vendor IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/vendors
// Scopes: None (public endpoint)
func (c *Client) GetVendorIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/vendors", options...)
}

// GetVendor returns a specific vendor by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/vendors
// Scopes: None (public endpoint)
func (c *Client) GetVendor(ctx context.Context, id int, options ...RequestOption) (*Vendor, error) {
	return GetByID[Vendor](ctx, c, "/v2/vendors", id, options...)
}

// GetWizardsVaultListingIDs returns all wizard's vault listing IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wizardsvault/listings
// Scopes: None (public endpoint)
func (c *Client) GetWizardsVaultListingIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/wizardsvault/listings", options...)
}

// GetWizardsVaultListing returns a specific wizard's vault listing by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wizardsvault/listings
// Scopes: None (public endpoint)
func (c *Client) GetWizardsVaultListing(ctx context.Context, id int, options ...RequestOption) (*WizardsVaultListingDetail, error) {
	return GetByID[WizardsVaultListingDetail](ctx, c, "/v2/wizardsvault/listings", id, options...)
}

// GetWizardsVaultObjectiveIDs returns all wizard's vault objective IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wizardsvault/objectives
// Scopes: None (public endpoint)
func (c *Client) GetWizardsVaultObjectiveIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/wizardsvault/objectives", options...)
}

// GetWizardsVaultObjective returns a specific wizard's vault objective by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wizardsvault/objectives
// Scopes: None (public endpoint)
func (c *Client) GetWizardsVaultObjective(ctx context.Context, id int, options ...RequestOption) (*WizardsVaultObjective, error) {
	return GetByID[WizardsVaultObjective](ctx, c, "/v2/wizardsvault/objectives", id, options...)
}

// GetWorldBosses returns world boss information.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/worldbosses
// Scopes: None (public endpoint)
func (c *Client) GetWorldBosses(ctx context.Context, options ...RequestOption) (*WorldBossDetail, error) {
	return GetSingle[WorldBossDetail](ctx, c, "/v2/worldbosses", options...)
}

// ========================
// Character Endpoints
// ========================

// GetCharacterNames returns all character names.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters
// Scopes: characters
func (c *Client) GetCharacterNames(ctx context.Context, options ...RequestOption) ([]string, error) {
	// Custom implementation that doesn't add ids=all
	opts := &RequestOptions{}
	for _, opt := range options {
		opt(opts)
	}

	data, _, err := c.get(ctx, "/v2/characters", opts)
	if err != nil {
		return nil, err
	}

	var results []string
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return results, nil
}

// GetCharacters returns all character details.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters
// Scopes: characters
func (c *Client) GetCharacters(ctx context.Context, options ...RequestOption) ([]Character, error) {
	return GetAll[Character](ctx, c, "/v2/characters", options...)
}

// GetCharacterBackstory returns character backstory.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters/(name)/backstory
// Scopes: characters
func (c *Client) GetCharacterBackstory(ctx context.Context, name string, options ...RequestOption) (*CharacterBackstory, error) {
	return GetSingle[CharacterBackstory](ctx, c, "/v2/characters/"+name+"/backstory", options...)
}

// GetCharacterBuildTabs returns character build tabs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters/(name)/buildtabs
// Scopes: characters, builds
func (c *Client) GetCharacterBuildTabs(ctx context.Context, name string, options ...RequestOption) ([]CharacterBuildTab, error) {
	return GetAll[CharacterBuildTab](ctx, c, "/v2/characters/"+name+"/buildtabs", options...)
}

// GetCharacterBuildTabActive returns active build tab.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters/(name)/buildtabs/active
// Scopes: characters, builds
func (c *Client) GetCharacterBuildTabActive(ctx context.Context, name string, options ...RequestOption) (*CharacterBuildTabActive, error) {
	return GetSingle[CharacterBuildTabActive](ctx, c, "/v2/characters/"+name+"/buildtabs/active", options...)
}

// GetCharacterCore returns core character information.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters/(name)/core
// Scopes: characters
func (c *Client) GetCharacterCore(ctx context.Context, name string, options ...RequestOption) (*CharacterCore, error) {
	return GetSingle[CharacterCore](ctx, c, "/v2/characters/"+name+"/core", options...)
}

// GetCharacterCrafting returns character crafting disciplines.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters/(name)/crafting
// Scopes: characters
func (c *Client) GetCharacterCrafting(ctx context.Context, name string, options ...RequestOption) ([]CharacterCrafting, error) {
	return GetAll[CharacterCrafting](ctx, c, "/v2/characters/"+name+"/crafting", options...)
}

// GetCharacterDungeons returns character dungeon progress.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters/(name)/dungeons
// Scopes: characters, progression
func (c *Client) GetCharacterDungeons(ctx context.Context, name string, options ...RequestOption) ([]CharacterDungeon, error) {
	return GetAll[CharacterDungeon](ctx, c, "/v2/characters/"+name+"/dungeons", options...)
}

// GetCharacterEquipment returns character equipment.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters/(name)/equipment
// Scopes: characters, inventories
func (c *Client) GetCharacterEquipment(ctx context.Context, name string, options ...RequestOption) ([]CharacterEquipment, error) {
	return GetAll[CharacterEquipment](ctx, c, "/v2/characters/"+name+"/equipment", options...)
}

// GetCharacterEquipmentTabs returns character equipment tabs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters/(name)/equipmenttabs
// Scopes: characters, inventories
func (c *Client) GetCharacterEquipmentTabs(ctx context.Context, name string, options ...RequestOption) ([]CharacterEquipmentTab, error) {
	return GetAll[CharacterEquipmentTab](ctx, c, "/v2/characters/"+name+"/equipmenttabs", options...)
}

// GetCharacterEquipmentTabActive returns active equipment tab.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters/(name)/equipmenttabs/active
// Scopes: characters, inventories
func (c *Client) GetCharacterEquipmentTabActive(ctx context.Context, name string, options ...RequestOption) (*CharacterEquipmentTabActive, error) {
	return GetSingle[CharacterEquipmentTabActive](ctx, c, "/v2/characters/"+name+"/equipmenttabs/active", options...)
}

// GetCharacterHeroPoints returns character hero points.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters/(name)/heropoints
// Scopes: characters, progression
func (c *Client) GetCharacterHeroPoints(ctx context.Context, name string, options ...RequestOption) ([]CharacterHeroPoint, error) {
	return GetAll[CharacterHeroPoint](ctx, c, "/v2/characters/"+name+"/heropoints", options...)
}

// GetCharacterInventory returns character inventory.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters/(name)/inventory
// Scopes: characters, inventories
func (c *Client) GetCharacterInventory(ctx context.Context, name string, options ...RequestOption) (*CharacterInventory, error) {
	return GetSingle[CharacterInventory](ctx, c, "/v2/characters/"+name+"/inventory", options...)
}

// GetCharacterQuests returns character quests.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters/(name)/quests
// Scopes: characters, progression
func (c *Client) GetCharacterQuests(ctx context.Context, name string, options ...RequestOption) ([]CharacterQuest, error) {
	return GetAll[CharacterQuest](ctx, c, "/v2/characters/"+name+"/quests", options...)
}

// GetCharacterRecipes returns character recipes.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters/(name)/recipes
// Scopes: characters, unlocks
func (c *Client) GetCharacterRecipes(ctx context.Context, name string, options ...RequestOption) (*CharacterRecipe, error) {
	return GetSingle[CharacterRecipe](ctx, c, "/v2/characters/"+name+"/recipes", options...)
}

// GetCharacterSAB returns character SAB progress.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters/(name)/sab
// Scopes: characters, progression
func (c *Client) GetCharacterSAB(ctx context.Context, name string, options ...RequestOption) (*CharacterSAB, error) {
	return GetSingle[CharacterSAB](ctx, c, "/v2/characters/"+name+"/sab", options...)
}

// GetCharacterSkills returns character skills.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters/(name)/skills
// Scopes: characters, builds
func (c *Client) GetCharacterSkills(ctx context.Context, name string, options ...RequestOption) (*CharacterSkills, error) {
	return GetSingle[CharacterSkills](ctx, c, "/v2/characters/"+name+"/skills", options...)
}

// GetCharacterSpecializations returns character specializations.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters/(name)/specializations
// Scopes: characters, builds
func (c *Client) GetCharacterSpecializations(ctx context.Context, name string, options ...RequestOption) ([]CharacterSpecialization, error) {
	return GetAll[CharacterSpecialization](ctx, c, "/v2/characters/"+name+"/specializations", options...)
}

// GetCharacterTraining returns character training.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/characters/(name)/training
// Scopes: characters, builds
func (c *Client) GetCharacterTraining(ctx context.Context, name string, options ...RequestOption) ([]CharacterTraining, error) {
	return GetAll[CharacterTraining](ctx, c, "/v2/characters/"+name+"/training", options...)
}

// ========================
// PvP Endpoints
// ========================

// GetPvP returns PvP information.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp
// Scopes: pvp
func (c *Client) GetPvP(ctx context.Context, options ...RequestOption) (*PvPStats, error) {
	return GetSingle[PvPStats](ctx, c, "/v2/pvp", options...)
}

// GetPvPAmuletIDs returns all PvP amulet IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/amulets
// Scopes: None (public endpoint)
func (c *Client) GetPvPAmuletIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/pvp/amulets", options...)
}

// GetPvPAmulet returns a specific PvP amulet by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/amulets
// Scopes: None (public endpoint)
func (c *Client) GetPvPAmulet(ctx context.Context, id int, options ...RequestOption) (*PvPAmulet, error) {
	return GetByID[PvPAmulet](ctx, c, "/v2/pvp/amulets", id, options...)
}

// GetPvPGames returns PvP games.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/games
// Scopes: pvp
func (c *Client) GetPvPGames(ctx context.Context, options ...RequestOption) ([]PvPGame, error) {
	return GetAll[PvPGame](ctx, c, "/v2/pvp/games", options...)
}

// GetPvPHeroIDs returns all PvP hero IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/heroes
// Scopes: None (public endpoint)
func (c *Client) GetPvPHeroIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/pvp/heroes", options...)
}

// GetPvPHero returns a specific PvP hero by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/heroes
// Scopes: None (public endpoint)
func (c *Client) GetPvPHero(ctx context.Context, id string, options ...RequestOption) (*PvPHero, error) {
	return GetSingle[PvPHero](ctx, c, "/v2/pvp/heroes/"+id, options...)
}

// GetPvPRankIDs returns all PvP rank IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/ranks
// Scopes: None (public endpoint)
func (c *Client) GetPvPRankIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/pvp/ranks", options...)
}

// GetPvPRank returns a specific PvP rank by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/ranks
// Scopes: None (public endpoint)
func (c *Client) GetPvPRank(ctx context.Context, id int, options ...RequestOption) (*PvPRank, error) {
	return GetByID[PvPRank](ctx, c, "/v2/pvp/ranks", id, options...)
}

// GetPvPRewardTrackIDs returns all PvP reward track IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/rewardtracks
// Scopes: None (public endpoint)
func (c *Client) GetPvPRewardTrackIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/pvp/rewardtracks", options...)
}

// GetPvPRewardTrack returns a specific PvP reward track by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/rewardtracks
// Scopes: None (public endpoint)
func (c *Client) GetPvPRewardTrack(ctx context.Context, id int, options ...RequestOption) (*PvPRewardTrack, error) {
	return GetByID[PvPRewardTrack](ctx, c, "/v2/pvp/rewardtracks", id, options...)
}

// GetPvPRuneIDs returns all PvP rune IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/runes
// Scopes: None (public endpoint)
func (c *Client) GetPvPRuneIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/pvp/runes", options...)
}

// GetPvPRune returns a specific PvP rune by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/runes
// Scopes: None (public endpoint)
func (c *Client) GetPvPRune(ctx context.Context, id int, options ...RequestOption) (*PvPRune, error) {
	return GetByID[PvPRune](ctx, c, "/v2/pvp/runes", id, options...)
}

// GetPvPSeasonIDs returns all PvP season IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/seasons
// Scopes: None (public endpoint)
func (c *Client) GetPvPSeasonIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/pvp/seasons", options...)
}

// GetPvPSeason returns a specific PvP season by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/seasons
// Scopes: None (public endpoint)
func (c *Client) GetPvPSeason(ctx context.Context, id string, options ...RequestOption) (*PvPSeason, error) {
	return GetSingle[PvPSeason](ctx, c, "/v2/pvp/seasons/"+id, options...)
}

// GetPvPSeasonLeaderboards returns PvP season leaderboard data.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/seasons/leaderboards
// Scopes: None (public endpoint)
func (c *Client) GetPvPSeasonLeaderboards(ctx context.Context, seasonID string, options ...RequestOption) (*PvPSeasonLeaderboardEntries, error) {
	return GetSingle[PvPSeasonLeaderboardEntries](ctx, c, "/v2/pvp/seasons/"+seasonID+"/leaderboards", options...)
}

// GetPvPSigilIDs returns all PvP sigil IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/sigils
// Scopes: None (public endpoint)
func (c *Client) GetPvPSigilIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/pvp/sigils", options...)
}

// GetPvPSigil returns a specific PvP sigil by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/sigils
// Scopes: None (public endpoint)
func (c *Client) GetPvPSigil(ctx context.Context, id int, options ...RequestOption) (*PvPSigil, error) {
	return GetByID[PvPSigil](ctx, c, "/v2/pvp/sigils", id, options...)
}

// GetPvPStandings returns PvP standings.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/standings
// Scopes: pvp
func (c *Client) GetPvPStandings(ctx context.Context, options ...RequestOption) (*PvPStandings, error) {
	return GetSingle[PvPStandings](ctx, c, "/v2/pvp/standings", options...)
}

// GetPvPStats returns PvP statistics.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/pvp/stats
// Scopes: pvp
func (c *Client) GetPvPStats(ctx context.Context, options ...RequestOption) (*PvPStats, error) {
	return GetSingle[PvPStats](ctx, c, "/v2/pvp/stats", options...)
}

// ========================
// WvW Endpoints
// ========================

// GetWvWAbilityIDs returns all WvW ability IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wvw/abilities
// Scopes: None (public endpoint)
func (c *Client) GetWvWAbilityIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/wvw/abilities", options...)
}

// GetWvWAbility returns a specific WvW ability by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wvw/abilities
// Scopes: None (public endpoint)
func (c *Client) GetWvWAbility(ctx context.Context, id int, options ...RequestOption) (*WvWAbility, error) {
	return GetByID[WvWAbility](ctx, c, "/v2/wvw/abilities", id, options...)
}

// GetWvWGuilds returns WvW guilds.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wvw/guilds
// Scopes: None (public endpoint)
func (c *Client) GetWvWGuilds(ctx context.Context, options ...RequestOption) ([]WvWGuild, error) {
	return GetAll[WvWGuild](ctx, c, "/v2/wvw/guilds", options...)
}

// GetWvWMatches returns WvW matches.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wvw/matches
// Scopes: None (public endpoint)
func (c *Client) GetWvWMatches(ctx context.Context, options ...RequestOption) ([]WvWMatch, error) {
	return GetAll[WvWMatch](ctx, c, "/v2/wvw/matches", options...)
}

// GetWvWMatchOverview returns WvW match overview.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wvw/matches/overview
// Scopes: None (public endpoint)
func (c *Client) GetWvWMatchOverview(ctx context.Context, options ...RequestOption) (*WvWMatchOverview, error) {
	return GetSingle[WvWMatchOverview](ctx, c, "/v2/wvw/matches/overview", options...)
}

// GetWvWMatchScores returns WvW match scores.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wvw/matches/scores
// Scopes: None (public endpoint)
func (c *Client) GetWvWMatchScores(ctx context.Context, options ...RequestOption) (*WvWMatchScores, error) {
	return GetSingle[WvWMatchScores](ctx, c, "/v2/wvw/matches/scores", options...)
}

// GetWvWMatchStats returns WvW match statistics.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wvw/matches/stats
// Scopes: None (public endpoint)
func (c *Client) GetWvWMatchStats(ctx context.Context, options ...RequestOption) (*WvWMatchStats, error) {
	return GetSingle[WvWMatchStats](ctx, c, "/v2/wvw/matches/stats", options...)
}

// GetWvWMatchStatsTeams returns detailed WvW match statistics by team.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wvw/matches/stats/teams
// Scopes: None (public endpoint)
func (c *Client) GetWvWMatchStatsTeams(ctx context.Context, matchID string, options ...RequestOption) (*WvWMatchStatsTeams, error) {
	return GetSingle[WvWMatchStatsTeams](ctx, c, "/v2/wvw/matches/"+matchID+"/stats/teams", options...)
}

// GetWvWObjectiveIDs returns all WvW objective IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wvw/objectives
// Scopes: None (public endpoint)
func (c *Client) GetWvWObjectiveIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/wvw/objectives", options...)
}

// GetWvWObjective returns a specific WvW objective by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wvw/objectives
// Scopes: None (public endpoint)
func (c *Client) GetWvWObjective(ctx context.Context, id string, options ...RequestOption) (*WvWObjective, error) {
	return GetSingle[WvWObjective](ctx, c, "/v2/wvw/objectives/"+id, options...)
}

// GetWvWRankIDs returns all WvW rank IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wvw/ranks
// Scopes: None (public endpoint)
func (c *Client) GetWvWRankIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/wvw/ranks", options...)
}

// GetWvWRank returns a specific WvW rank by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wvw/ranks
// Scopes: None (public endpoint)
func (c *Client) GetWvWRank(ctx context.Context, id int, options ...RequestOption) (*WvWRank, error) {
	return GetByID[WvWRank](ctx, c, "/v2/wvw/ranks", id, options...)
}

// GetWvWRewardTrackIDs returns all WvW reward track IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wvw/rewardtracks
// Scopes: None (public endpoint)
func (c *Client) GetWvWRewardTrackIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/wvw/rewardtracks", options...)
}

// GetWvWRewardTrack returns a specific WvW reward track by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wvw/rewardtracks
// Scopes: None (public endpoint)
func (c *Client) GetWvWRewardTrack(ctx context.Context, id int, options ...RequestOption) (*WvWRewardTrack, error) {
	return GetByID[WvWRewardTrack](ctx, c, "/v2/wvw/rewardtracks", id, options...)
}

// GetWvWTimers returns WvW timers.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wvw/timers
// Scopes: None (public endpoint)
func (c *Client) GetWvWTimers(ctx context.Context, options ...RequestOption) (*WvWTimer, error) {
	return GetSingle[WvWTimer](ctx, c, "/v2/wvw/timers", options...)
}

// GetWvWUpgradeIDs returns all WvW upgrade IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wvw/upgrades
// Scopes: None (public endpoint)
func (c *Client) GetWvWUpgradeIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/wvw/upgrades", options...)
}

// GetWvWUpgrade returns a specific WvW upgrade by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/wvw/upgrades
// Scopes: None (public endpoint)
func (c *Client) GetWvWUpgrade(ctx context.Context, id int, options ...RequestOption) (*WvWUpgrade, error) {
	return GetByID[WvWUpgrade](ctx, c, "/v2/wvw/upgrades", id, options...)
}

// ========================
// Guild Endpoints
// ========================

// GetGuild returns guild information by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/guild/(id)
// Scopes: guilds
func (c *Client) GetGuild(ctx context.Context, id string, options ...RequestOption) (*Guild, error) {
	return GetSingle[Guild](ctx, c, "/v2/guild/"+id, options...)
}

// GetGuildLog returns guild log.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/guild/(id)/log
// Scopes: guilds
func (c *Client) GetGuildLog(ctx context.Context, id string, options ...RequestOption) ([]GuildLog, error) {
	return GetAll[GuildLog](ctx, c, "/v2/guild/"+id+"/log", options...)
}

// GetGuildMembers returns guild members.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/guild/(id)/members
// Scopes: guilds
func (c *Client) GetGuildMembers(ctx context.Context, id string, options ...RequestOption) ([]GuildMember, error) {
	return GetAll[GuildMember](ctx, c, "/v2/guild/"+id+"/members", options...)
}

// GetGuildRanks returns guild ranks.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/guild/(id)/ranks
// Scopes: guilds
func (c *Client) GetGuildRanks(ctx context.Context, id string, options ...RequestOption) ([]GuildRank, error) {
	return GetAll[GuildRank](ctx, c, "/v2/guild/"+id+"/ranks", options...)
}

// GetGuildStash returns guild stash.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/guild/(id)/stash
// Scopes: guilds
func (c *Client) GetGuildStash(ctx context.Context, id string, options ...RequestOption) ([]GuildStash, error) {
	return GetAll[GuildStash](ctx, c, "/v2/guild/"+id+"/stash", options...)
}

// GetGuildStorage returns guild storage.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/guild/(id)/storage
// Scopes: guilds
func (c *Client) GetGuildStorage(ctx context.Context, id string, options ...RequestOption) ([]GuildStorage, error) {
	return GetAll[GuildStorage](ctx, c, "/v2/guild/"+id+"/storage", options...)
}

// GetGuildTeams returns guild teams.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/guild/(id)/teams
// Scopes: guilds
func (c *Client) GetGuildTeams(ctx context.Context, id string, options ...RequestOption) ([]GuildTeam, error) {
	return GetAll[GuildTeam](ctx, c, "/v2/guild/"+id+"/teams", options...)
}

// GetGuildTreasury returns guild treasury.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/guild/(id)/treasury
// Scopes: guilds
func (c *Client) GetGuildTreasury(ctx context.Context, id string, options ...RequestOption) ([]GuildTreasury, error) {
	return GetAll[GuildTreasury](ctx, c, "/v2/guild/"+id+"/treasury", options...)
}

// GetGuildUpgrades returns guild upgrades.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/guild/(id)/upgrades
// Scopes: guilds
func (c *Client) GetGuildUpgrades(ctx context.Context, id string, options ...RequestOption) ([]GuildUpgrade, error) {
	return GetAll[GuildUpgrade](ctx, c, "/v2/guild/"+id+"/upgrades", options...)
}

// GetGuildPermissionIDs returns all guild permission IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/guild/permissions
// Scopes: None (public endpoint)
func (c *Client) GetGuildPermissionIDs(ctx context.Context, options ...RequestOption) ([]string, error) {
	return GetAll[string](ctx, c, "/v2/guild/permissions", options...)
}

// GetGuildPermission returns a specific guild permission by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/guild/permissions
// Scopes: None (public endpoint)
func (c *Client) GetGuildPermission(ctx context.Context, id string, options ...RequestOption) (*GuildPermission, error) {
	return GetSingle[GuildPermission](ctx, c, "/v2/guild/permissions/"+id, options...)
}

// GetGuildSearch returns guild search functionality.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/guild/search
// Scopes: None (public endpoint)
func (c *Client) GetGuildSearch(ctx context.Context, options ...RequestOption) (*GuildSearch, error) {
	return GetSingle[GuildSearch](ctx, c, "/v2/guild/search", options...)
}

// GetGuildUpgradeDetailIDs returns all guild upgrade detail IDs.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/guild/upgrades
// Scopes: None (public endpoint)
func (c *Client) GetGuildUpgradeDetailIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/guild/upgrades", options...)
}

// GetGuildUpgradeDetail returns a specific guild upgrade detail by ID.
// Wiki: https://wiki.guildwars2.com/wiki/API:2/guild/upgrades
// Scopes: None (public endpoint)
func (c *Client) GetGuildUpgradeDetail(ctx context.Context, id int, options ...RequestOption) (*GuildUpgradeDetail, error) {
	return GetByID[GuildUpgradeDetail](ctx, c, "/v2/guild/upgrades", id, options...)
}
