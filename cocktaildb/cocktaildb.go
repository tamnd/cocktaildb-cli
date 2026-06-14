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
const Host = "thecocktaildb.com"

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
		Rate:      200 * time.Millisecond,
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

// Search searches cocktails by name. It returns at most limit results (pass 0
// for all). If the API returns no results it returns an empty slice and nil error.
func (c *Client) Search(ctx context.Context, name string, limit int) ([]Cocktail, error) {
	u := fmt.Sprintf("%s/search.php?s=%s", c.cfg.BaseURL, neturl.QueryEscape(name))
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp drinksResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode search: %w", err)
	}
	items := make([]Cocktail, 0, len(resp.Drinks))
	for i, d := range resp.Drinks {
		items = append(items, toCocktail(d, i+1))
	}
	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}
	return items, nil
}

// Random returns one random cocktail from the API.
func (c *Client) Random(ctx context.Context) (Cocktail, error) {
	u := fmt.Sprintf("%s/random.php", c.cfg.BaseURL)
	body, err := c.get(ctx, u)
	if err != nil {
		return Cocktail{}, err
	}
	var resp drinksResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return Cocktail{}, fmt.Errorf("decode random: %w", err)
	}
	if len(resp.Drinks) == 0 {
		return Cocktail{}, fmt.Errorf("random: no drink returned")
	}
	return toCocktail(resp.Drinks[0], 1), nil
}

// Categories returns all cocktail categories.
func (c *Client) Categories(ctx context.Context) ([]Category, error) {
	u := fmt.Sprintf("%s/list.php?c=list", c.cfg.BaseURL)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp categoriesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode categories: %w", err)
	}
	cats := make([]Category, 0, len(resp.Drinks))
	for i, d := range resp.Drinks {
		cats = append(cats, Category{Rank: i + 1, Name: d.StrCategory})
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
