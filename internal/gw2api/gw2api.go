// Package gw2api provides a fully typed client for the Guild Wars 2 API v2
package gw2api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	BaseURL     = "https://api.guildwars2.com"
	UserAgent   = "gw2api-go/1.0"
	DefaultLang = LanguageEnglish
)

// Client provides access to the Guild Wars 2 API
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	language   Language
	userAgent  string
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

// NewClient creates a new GW2 API client
func NewClient(options ...ClientOption) *Client {
	c := &Client{
		baseURL:    BaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		language:   DefaultLang,
		userAgent:  UserAgent,
	}

	for _, opt := range options {
		opt(c)
	}

	return c
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Text != "" {
			return nil, nil, apiErr
		}
		return nil, nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
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

// GetBuild returns the current game build
func (c *Client) GetBuild(ctx context.Context, options ...RequestOption) (*Build, error) {
	return GetSingle[Build](ctx, c, "/v2/build", options...)
}

// GetAchievementIDs returns all available achievement IDs
func (c *Client) GetAchievementIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/achievements", options...)
}

// GetAchievement returns a specific achievement by ID
func (c *Client) GetAchievement(ctx context.Context, id int, options ...RequestOption) (*Achievement, error) {
	return GetByID[Achievement](ctx, c, "/v2/achievements", id, options...)
}

// GetAchievements returns multiple achievements by IDs
func (c *Client) GetAchievements(ctx context.Context, ids []int, options ...RequestOption) ([]*Achievement, error) {
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

// GetCurrencyIDs returns all available currency IDs
func (c *Client) GetCurrencyIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/currencies", options...)
}

// GetCurrency returns a specific currency by ID
func (c *Client) GetCurrency(ctx context.Context, id int, options ...RequestOption) (*Currency, error) {
	return GetByID[Currency](ctx, c, "/v2/currencies", id, options...)
}

// GetCurrencies returns multiple currencies by IDs
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

// GetItemIDs returns all available item IDs
func (c *Client) GetItemIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/items", options...)
}

// GetItem returns a specific item by ID
func (c *Client) GetItem(ctx context.Context, id int, options ...RequestOption) (*Item, error) {
	return GetByID[Item](ctx, c, "/v2/items", id, options...)
}

// GetItems returns multiple items by IDs
func (c *Client) GetItems(ctx context.Context, ids []int, options ...RequestOption) ([]*Item, error) {
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

// GetWorldIDs returns all available world IDs
func (c *Client) GetWorldIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/worlds", options...)
}

// GetWorld returns a specific world by ID
func (c *Client) GetWorld(ctx context.Context, id int, options ...RequestOption) (*World, error) {
	return GetByID[World](ctx, c, "/v2/worlds", id, options...)
}

// GetWorlds returns multiple worlds by IDs
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

// GetWorldsPage returns a page of worlds
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

// GetCommercePriceIDs returns all available item IDs with trading post prices
func (c *Client) GetCommercePriceIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/commerce/prices", options...)
}

// GetCommercePrice returns trading post price information for a specific item
func (c *Client) GetCommercePrice(ctx context.Context, itemID int, options ...RequestOption) (*Price, error) {
	return GetByID[Price](ctx, c, "/v2/commerce/prices", itemID, options...)
}

// GetCommercePrices returns trading post price information for multiple items
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

// GetSkillIDs returns all available skill IDs
func (c *Client) GetSkillIDs(ctx context.Context, options ...RequestOption) ([]int, error) {
	return GetIDs[int](ctx, c, "/v2/skills", options...)
}

// GetSkill returns a specific skill by ID
func (c *Client) GetSkill(ctx context.Context, id int, options ...RequestOption) (*Skill, error) {
	return GetByID[Skill](ctx, c, "/v2/skills", id, options...)
}

// GetSkills returns multiple skills by IDs
func (c *Client) GetSkills(ctx context.Context, ids []int, options ...RequestOption) ([]*Skill, error) {
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

// GetAllCurrencies returns all currencies
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

// GetAllWorlds returns all worlds
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
