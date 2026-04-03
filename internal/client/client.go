package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rvillarreal/taskpad/internal/model"
)

// Client is an HTTP client for the taskpad API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New creates a new taskpad API client. apiKey may be empty to disable auth.
func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// --- Todos ---

func (c *Client) CreateTodo(req model.CreateTodoRequest) (*model.Todo, error) {
	var todo model.Todo
	if err := c.post("/api/v1/todos", req, &todo); err != nil {
		return nil, err
	}
	return &todo, nil
}

func (c *Client) GetTodo(id string) (*model.Todo, error) {
	var todo model.Todo
	if err := c.get("/api/v1/todos/"+id, nil, &todo); err != nil {
		return nil, err
	}
	return &todo, nil
}

func (c *Client) ListTodos(params map[string]string) (*model.ListResult[model.Todo], error) {
	var result model.ListResult[model.Todo]
	if err := c.get("/api/v1/todos", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateTodo(id string, req model.UpdateTodoRequest) (*model.Todo, error) {
	var todo model.Todo
	if err := c.put("/api/v1/todos/"+id, req, &todo); err != nil {
		return nil, err
	}
	return &todo, nil
}

func (c *Client) DeleteTodo(id string) error {
	return c.del("/api/v1/todos/" + id)
}

func (c *Client) CompleteTodo(id string, completed bool) (*model.Todo, error) {
	var todo model.Todo
	body := map[string]bool{"completed": completed}
	if err := c.patch("/api/v1/todos/"+id+"/complete", body, &todo); err != nil {
		return nil, err
	}
	return &todo, nil
}

// --- Notes ---

func (c *Client) CreateNote(req model.CreateNoteRequest) (*model.Note, error) {
	var note model.Note
	if err := c.post("/api/v1/notes", req, &note); err != nil {
		return nil, err
	}
	return &note, nil
}

func (c *Client) GetNote(id string) (*model.Note, error) {
	var note model.Note
	if err := c.get("/api/v1/notes/"+id, nil, &note); err != nil {
		return nil, err
	}
	return &note, nil
}

func (c *Client) ListNotes(params map[string]string) (*model.ListResult[model.Note], error) {
	var result model.ListResult[model.Note]
	if err := c.get("/api/v1/notes", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateNote(id string, req model.UpdateNoteRequest) (*model.Note, error) {
	var note model.Note
	if err := c.put("/api/v1/notes/"+id, req, &note); err != nil {
		return nil, err
	}
	return &note, nil
}

func (c *Client) DeleteNote(id string) error {
	return c.del("/api/v1/notes/" + id)
}

// --- HTTP helpers ---

func (c *Client) get(path string, params map[string]string, dest any) error {
	u := c.baseURL + path
	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		u += "?" + q.Encode()
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}
	c.setAuth(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	return c.handleResponse(resp, dest)
}

func (c *Client) post(path string, body any, dest any) error {
	return c.doJSON("POST", path, body, dest)
}

func (c *Client) put(path string, body any, dest any) error {
	return c.doJSON("PUT", path, body, dest)
}

func (c *Client) patch(path string, body any, dest any) error {
	return c.doJSON("PATCH", path, body, dest)
}

func (c *Client) del(path string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+path, nil)
	if err != nil {
		return err
	}
	c.setAuth(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return c.parseError(resp)
	}
	return nil
}

func (c *Client) doJSON(method, path string, body any, dest any) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	return c.handleResponse(resp, dest)
}

func (c *Client) handleResponse(resp *http.Response, dest any) error {
	if resp.StatusCode >= 400 {
		return c.parseError(resp)
	}
	if dest != nil {
		return json.NewDecoder(resp.Body).Decode(dest)
	}
	return nil
}

func (c *Client) setAuth(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}

func (c *Client) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var apiErr struct {
		Error string `json:"error"`
		Code  int    `json:"code"`
	}
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Error != "" {
		return fmt.Errorf("%s", apiErr.Error)
	}
	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
}
