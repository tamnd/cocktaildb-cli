// Package cocktaildb is the library behind the cocktaildb command line:
// the HTTP client, request shaping, and the typed data models for TheCocktailDB.
//
// The free v1 API uses the static key "1" baked into the base URL path
// (https://www.thecocktaildb.com/api/json/v1/1). No login, no OAuth. The
// Client sets a polite User-Agent, paces requests, and retries transient
// failures (429 and 5xx) with a capped exponential backoff.
package cocktaildb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"sync"
	"time"
)

// Host is the site this client talks to.
const Host = "www.thecocktaildb.com"

// Config holds all tunable parameters for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns a Config with sensible defaults for the free v1 API.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://www.thecocktaildb.com/api/json/v1/1",
		UserAgent: "Mozilla/5.0 (compatible; cocktaildb-cli/dev; +https://github.com/tamnd/cocktaildb-cli)",
		Rate:      300 * time.Millisecond,
		Timeout:   15 * time.Second,
		Retries:   3,
	}
}

// Client talks to TheCocktailDB over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// Search searches drinks by name. It returns at most limit results (pass 0
// for all). If the API returns no results it returns an empty slice and nil error.
func (c *Client) Search(ctx context.Context, name string, limit int) ([]Drink, error) {
	u := fmt.Sprintf("%s/search.php?s=%s", c.cfg.BaseURL, neturl.QueryEscape(name))
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp drinksResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode search: %w", err)
	}
	items := make([]Drink, 0, len(resp.Drinks))
	for _, d := range resp.Drinks {
		items = append(items, toDrink(d))
	}
	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}
	return items, nil
}

// Lookup returns a drink by ID. Returns an error if the ID is not found.
func (c *Client) Lookup(ctx context.Context, id string) (Drink, error) {
	u := fmt.Sprintf("%s/lookup.php?i=%s", c.cfg.BaseURL, neturl.QueryEscape(id))
	body, err := c.get(ctx, u)
	if err != nil {
		return Drink{}, err
	}
	var resp drinksResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return Drink{}, fmt.Errorf("decode lookup: %w", err)
	}
	if len(resp.Drinks) == 0 {
		return Drink{}, fmt.Errorf("drink %s: not found", id)
	}
	return toDrink(resp.Drinks[0]), nil
}

// Random returns one random drink from the API.
func (c *Client) Random(ctx context.Context) (Drink, error) {
	u := fmt.Sprintf("%s/random.php", c.cfg.BaseURL)
	body, err := c.get(ctx, u)
	if err != nil {
		return Drink{}, err
	}
	var resp drinksResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return Drink{}, fmt.Errorf("decode random: %w", err)
	}
	if len(resp.Drinks) == 0 {
		return Drink{}, fmt.Errorf("random: no drink returned")
	}
	return toDrink(resp.Drinks[0]), nil
}

// ListCategories returns all entries for a given list type.
// listType must be one of: "categories", "alcoholic", "glass", "ingredients".
func (c *Client) ListCategories(ctx context.Context, listType string) ([]Category, error) {
	var param, field string
	switch listType {
	case "alcoholic":
		param = "a=list"
		field = "strAlcoholic"
	case "glass":
		param = "g=list"
		field = "strGlass"
	case "ingredients":
		param = "i=list"
		field = "strIngredient1"
	default: // "categories" or empty
		param = "c=list"
		field = "strCategory"
	}
	u := fmt.Sprintf("%s/list.php?%s", c.cfg.BaseURL, param)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp listResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode list (%s): %w", listType, err)
	}
	cats := make([]Category, 0, len(resp.Drinks))
	for _, d := range resp.Drinks {
		name := d[field]
		if name == "" {
			// fallback: try any value
			for _, v := range d {
				if v != "" {
					name = v
					break
				}
			}
		}
		if name != "" {
			cats = append(cats, Category{Name: name})
		}
	}
	return cats, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	return b, err != nil, err
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	return min(time.Duration(attempt)*500*time.Millisecond, 5*time.Second)
}
