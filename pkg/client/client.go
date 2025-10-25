package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/admin/wordle/pkg/api"
)

// Client represents a Wordle game client
type Client struct {
	serverURL string
	gameID    string
	client    *http.Client
}

// NewClient creates a new game client
func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: serverURL,
		client:    &http.Client{},
	}
}

// NewGame creates a new game on the server
func (c *Client) NewGame() (*api.NewGameResponse, error) {
	// Server uses its own configuration
	req := api.NewGameRequest{}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(
		c.serverURL+"/game/new",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var response api.NewGameResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	c.gameID = response.GameID
	return &response, nil
}

// MakeGuess submits a guess to the server
func (c *Client) MakeGuess(guess string) (*api.GuessResponse, error) {
	if c.gameID == "" {
		return nil, fmt.Errorf("no active game, call NewGame first")
	}

	req := api.GuessRequest{
		Guess: guess,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/game/%s/guess", c.serverURL, c.gameID)
	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var response api.GuessResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetStatus retrieves the current game status
func (c *Client) GetStatus() (*api.GameStatusResponse, error) {
	if c.gameID == "" {
		return nil, fmt.Errorf("no active game, call NewGame first")
	}

	url := fmt.Sprintf("%s/game/%s/status", c.serverURL, c.gameID)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var response api.GameStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// parseError parses error response from server
func (c *Client) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var errResp api.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil {
		return fmt.Errorf("server error: %s", errResp.Error)
	}
	return fmt.Errorf("server returned status %d", resp.StatusCode)
}

// GetGameID returns the current game ID
func (c *Client) GetGameID() string {
	return c.gameID
}
