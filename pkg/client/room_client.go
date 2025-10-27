package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/admin/wordle/pkg/api"
)

// RoomClient handles HTTP communication for multiplayer rooms
type RoomClient struct {
	serverURL string
	client    *http.Client
	roomID    string
	playerID  string
	nickname  string
}

// NewRoomClient creates a new room client
func NewRoomClient(serverURL string) *RoomClient {
	return &RoomClient{
		serverURL: serverURL,
		client:    &http.Client{},
	}
}

// CreateRoom creates a new multiplayer room
func (c *RoomClient) CreateRoom(nickname string, maxPlayers int) (*api.CreateRoomResponse, error) {
	req := api.CreateRoomRequest{
		Nickname:   nickname,
		MaxPlayers: maxPlayers,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/room/create", c.serverURL)
	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp api.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var response api.CreateRoomResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	// Extract player ID from message
	c.roomID = response.RoomID
	c.nickname = nickname
	// Message format: "Room created! You are the host. Player ID: player-xxx"
	fmt.Sscanf(response.Message, "Room created! You are the host. Player ID: %s", &c.playerID)

	return &response, nil
}

// JoinRoom joins an existing room
func (c *RoomClient) JoinRoom(roomID, nickname string) (*api.JoinRoomResponse, error) {
	req := api.JoinRoomRequest{
		Nickname: nickname,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/room/%s/join", c.serverURL, roomID)
	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var response api.JoinRoomResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	c.roomID = roomID
	c.nickname = nickname
	// Extract player ID from message
	fmt.Sscanf(response.Message, "Joined room successfully! Player ID: %s", &c.playerID)

	return &response, nil
}

// StartGame starts the game (host only)
func (c *RoomClient) StartGame() error {
	url := fmt.Sprintf("%s/room/%s/start?player_id=%s", c.serverURL, c.roomID, c.playerID)
	resp, err := c.client.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("server error: %s", errResp.Error)
	}

	return nil
}

// MakeGuess submits a guess
func (c *RoomClient) MakeGuess(guess string) (*api.GuessResponse, error) {
	req := api.RoomGuessRequest{
		PlayerID: c.playerID,
		Guess:    guess,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/room/%s/guess", c.serverURL, c.roomID)
	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var response api.GuessResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetProgress gets the current progress with long polling
func (c *RoomClient) GetProgress(version int) (*api.RoomProgressResponse, error) {
	url := fmt.Sprintf("%s/room/%s/progress?version=%d", c.serverURL, c.roomID, version)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error: %s", string(bodyBytes))
	}

	var response api.RoomProgressResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetRoomStatus gets the room status
func (c *RoomClient) GetRoomStatus() (*api.RoomStatusResponse, error) {
	url := fmt.Sprintf("%s/room/%s/status", c.serverURL, c.roomID)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var response api.RoomStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// ListRooms lists all available rooms
func (c *RoomClient) ListRooms() (*api.ListRoomsResponse, error) {
	url := fmt.Sprintf("%s/room/list", c.serverURL)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var response api.ListRoomsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetRoomID returns the current room ID
func (c *RoomClient) GetRoomID() string {
	return c.roomID
}

// GetPlayerID returns the current player ID
func (c *RoomClient) GetPlayerID() string {
	return c.playerID
}

// GetNickname returns the player's nickname
func (c *RoomClient) GetNickname() string {
	return c.nickname
}
