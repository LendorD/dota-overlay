package opendota

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://api.opendota.com/api"

type Client struct {
	http    *http.Client
	baseURL string
	apiKey  string
}

type Hero struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	LocalizedName string `json:"localized_name"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: defaultBaseURL,
		apiKey:  apiKey,
	}
}

func (c *Client) GetHeroMatchups(ctx context.Context, heroID int) ([]HeroMatchup, error) {
	endpoint := fmt.Sprintf("%s/heroes/%d/matchups", c.baseURL, heroID)
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	if c.apiKey != "" {
		q := u.Query()
		q.Set("api_key", c.apiKey)
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("opendota: unexpected status %s", resp.Status)
	}

	var matchups []HeroMatchup
	if err := json.NewDecoder(resp.Body).Decode(&matchups); err != nil {
		return nil, err
	}
	return matchups, nil
}

func (c *Client) GetHeroes(ctx context.Context) ([]Hero, error) {
	endpoint := fmt.Sprintf("%s/heroes", c.baseURL)
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	if c.apiKey != "" {
		q := u.Query()
		q.Set("api_key", c.apiKey)
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("opendota: unexpected status %s", resp.Status)
	}

	var heroes []Hero
	if err := json.NewDecoder(resp.Body).Decode(&heroes); err != nil {
		return nil, err
	}
	return heroes, nil
}
