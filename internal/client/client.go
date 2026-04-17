package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

var ErrNotFound = errors.New("note not found")
var ErrUnreachable = errors.New("server unreachable")

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) Get(id string) ([]byte, error) {
	url := c.baseURL + "/notes/" + id
	slog.Debug("client get", "url", url)

	resp, err := c.http.Get(url)
	if err != nil {
		slog.Debug("client get failed", "err", err)
		return nil, ErrUnreachable
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (c *Client) Save(id string, body []byte) error {
	url := c.baseURL + "/notes/" + id
	slog.Debug("client save", "url", url, "bytes", len(body))

	resp, err := c.http.Post(url, "text/markdown", bytes.NewReader(body))
	if err != nil {
		slog.Debug("client save failed", "err", err)
		return ErrUnreachable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) Delete(id string) error {
	url := c.baseURL + "/notes/" + id
	slog.Debug("client delete", "url", url)

	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	resp, err := c.http.Do(req)
	if err != nil {
		slog.Debug("client delete failed", "err", err)
		return ErrUnreachable
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}
