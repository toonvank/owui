package api

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/toonvank/owui/internal/config"
	"github.com/toonvank/owui/internal/stream"
)

type Client struct {
	cfg    config.Config
	http   *http.Client
	base   string
	apiKey string
}

func New(cfg config.Config) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.InsecureTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &Client{
		cfg: cfg,
		http: &http.Client{
			Timeout:   time.Duration(cfg.TimeoutSec) * time.Second,
			Transport: transport,
		},
		base:   strings.TrimRight(cfg.BaseURL, "/"),
		apiKey: cfg.APIKey,
	}
}

func (c *Client) request(method, path string, body any, out any) error {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		data, status, err := c.doRequest(method, path, body)
		if err != nil {
			lastErr = err
			continue
		}
		if status == 502 || status == 503 || status == 504 {
			lastErr = parseAPIError(status, data)
			continue
		}
		if status >= 400 {
			return parseAPIError(status, data)
		}
		if out == nil || len(data) == 0 {
			return nil
		}
		return json.Unmarshal(data, out)
	}
	return lastErr
}

func (c *Client) doRequest(method, path string, body any) ([]byte, int, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.base+path, reader)
	if err != nil {
		return nil, 0, err
	}

	c.setAuth(req)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return data, resp.StatusCode, nil
}

func (c *Client) setAuth(req *http.Request) {
	if c.cfg.CustomHeader != "" {
		req.Header.Set(c.cfg.CustomHeader, c.apiKey)
		return
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
}

func parseAPIError(code int, data []byte) error {
	var errBody struct {
		Detail any    `json:"detail"`
		Error  string `json:"error"`
	}
	_ = json.Unmarshal(data, &errBody)

	msg := errBody.Error
	if msg == "" && errBody.Detail != nil {
		switch d := errBody.Detail.(type) {
		case string:
			msg = d
		default:
			b, _ := json.Marshal(d)
			msg = string(b)
		}
	}
	if msg == "" {
		msg = strings.TrimSpace(string(data))
	}
	if msg == "" {
		msg = http.StatusText(code)
	}
	err := fmt.Errorf("%s (HTTP %d)", msg, code)
	return enrichAPIError(code, msg, err)
}

func enrichAPIError(code int, msg string, err error) error {
	lower := strings.ToLower(msg)
	if code == 404 || strings.Contains(lower, "unexpected mimetype") {
		return fmt.Errorf("%w — upstream model provider may be down (check ollama-hack / Open WebUI connections)", err)
	}
	if code == 502 || code == 503 || code == 504 {
		return fmt.Errorf("%w — backend overloaded or unreachable", err)
	}
	return err
}

type ServerConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Status  bool   `json:"status"`
}

func (c *Client) Health() error {
	var cfg ServerConfig
	return c.request(http.MethodGet, "/health", nil, &cfg)
}

func (c *Client) Config() (ServerConfig, error) {
	var cfg ServerConfig
	err := c.request(http.MethodGet, "/api/config", nil, &cfg)
	return cfg, err
}

type Model struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	OwnedBy    string `json:"owned_by"`
	Preset     bool   `json:"preset"`
	Connection string `json:"connection_type"`
	Info       any    `json:"info"`
	Object     string `json:"object"`
}

func (m Model) DisplayName() string {
	if m.Name != "" && m.Name != m.ID {
		return m.Name
	}
	return ""
}

func (m Model) IsCustom() bool {
	return m.Preset || m.Info != nil
}

type ModelsResponse struct {
	Data []Model `json:"data"`
}

func (c *Client) ListModels() ([]Model, error) {
	var raw json.RawMessage
	if err := c.request(http.MethodGet, "/api/models", nil, &raw); err != nil {
		return nil, err
	}

	var wrapped ModelsResponse
	if err := json.Unmarshal(raw, &wrapped); err == nil && len(wrapped.Data) > 0 {
		return normalizeModels(wrapped.Data), nil
	}

	var direct []Model
	if err := json.Unmarshal(raw, &direct); err == nil && len(direct) > 0 {
		return normalizeModels(direct), nil
	}

	return nil, fmt.Errorf("unexpected models response format")
}

func normalizeModels(models []Model) []Model {
	out := make([]Model, 0, len(models))
	for _, m := range models {
		if m.ID == "" {
			m.ID = m.Name
		}
		if m.ID != "" {
			out = append(out, m)
		}
	}
	return out
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model      string         `json:"model"`
	Messages   []Message      `json:"messages"`
	Stream     bool           `json:"stream"`
	ChatID     string         `json:"chat_id,omitempty"`
	ToolIDs    []string       `json:"tool_ids,omitempty"`
	FilterIDs  []string       `json:"filter_ids,omitempty"`
	Features   map[string]any `json:"features,omitempty"`
	Files      []any          `json:"files,omitempty"`
	Collection string         `json:"collection,omitempty"`
}

type ChatOptions struct {
	ToolIDs           []string
	FilterIDs         []string
	FileIDs           []string
	Collection        string
	ChatID            string
	Features          map[string]any
	SkipModelFeatures bool
	ExplicitFilters   bool
	ExplicitTools     bool
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func (c *Client) Chat(messages []Message, model string, streamEnabled bool, onDelta func(string) error) (string, error) {
	return c.ChatWithOptions(messages, model, streamEnabled, nil, onDelta)
}

func (c *Client) ChatWithOptions(messages []Message, model string, streamEnabled bool, opts *ChatOptions, onDelta func(string) error) (string, error) {
	if model == "" {
		model = c.cfg.DefaultModel
	}

	reqBody := ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   streamEnabled,
	}
	if opts != nil {
		reqBody.ToolIDs = opts.ToolIDs
		reqBody.FilterIDs = opts.FilterIDs
		reqBody.Features = opts.Features
		reqBody.Collection = opts.Collection
		reqBody.ChatID = opts.ChatID
		if len(opts.FileIDs) > 0 {
			files := make([]any, len(opts.FileIDs))
			for i, id := range opts.FileIDs {
				files[i] = map[string]string{"id": id}
			}
			reqBody.Files = files
		}
	}

	c.applyModelIntegrations(&reqBody, opts)

	if streamEnabled {
		reply, err := c.chatStream(reqBody, onDelta)
		if err == nil {
			return reply, nil
		}
		// Some reverse-proxy setups break SSE; fall back to non-streaming.
		if onDelta != nil {
			reqBody.Stream = false
			reply, err2 := c.chatSync(reqBody)
			if err2 == nil {
				_ = onDelta(reply)
				return reply, nil
			}
		}
		return reply, err
	}

	return c.chatSync(reqBody)
}

func (c *Client) applyModelIntegrations(req *ChatRequest, opts *ChatOptions) {
	if opts != nil && opts.SkipModelFeatures {
		return
	}
	if !c.cfg.ShouldApplyModelFeatures() {
		return
	}

	model, _ := c.ModelByID(req.Model)
	if model.ID == "" {
		return
	}

	if req.Features == nil {
		req.Features = FeaturesFromModel(model)
	}
	explicitFilters := opts != nil && opts.ExplicitFilters
	if len(req.FilterIDs) == 0 && !explicitFilters {
		extra := c.cfg.FilterIDs
		if opts != nil && len(opts.FilterIDs) > 0 {
			extra = append(extra, opts.FilterIDs...)
		}
		req.FilterIDs = FilterIDsFromModel(model, extra)
	}
	explicitTools := opts != nil && opts.ExplicitTools
	if len(req.ToolIDs) == 0 && !explicitTools {
		if opts != nil && len(opts.ToolIDs) > 0 {
			req.ToolIDs = opts.ToolIDs
		} else if len(c.cfg.ToolIDs) > 0 {
			req.ToolIDs = c.cfg.ToolIDs
		}
	}
}

func (c *Client) chatSync(reqBody ChatRequest) (string, error) {
	reqBody.Stream = false
	var resp ChatResponse
	if err := c.request(http.MethodPost, "/api/chat/completions", reqBody, &resp); err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response from model")
	}
	return resp.Choices[0].Message.Content, nil
}

func (c *Client) chatStream(reqBody ChatRequest, onDelta func(string) error) (string, error) {
	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, c.base+"/api/chat/completions", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return "", parseAPIError(resp.StatusCode, data)
	}

	return stream.ParseChatCompletions(resp.Body, onDelta)
}

type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

func (c *Client) SignIn(email, password string) (AuthResponse, error) {
	var resp AuthResponse
	err := c.request(http.MethodPost, "/api/v1/auths/signin", SignInRequest{Email: email, Password: password}, &resp)
	return resp, err
}

type ChatSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	UpdatedAt int64  `json:"updated_at"`
	CreatedAt int64  `json:"created_at"`
	Pinned    bool   `json:"pinned,omitempty"`
}

func (c *Client) ListChats(page int) ([]ChatSummary, error) {
	var chats []ChatSummary
	path := fmt.Sprintf("/api/v1/chats/?page=%d", page)
	err := c.request(http.MethodGet, path, nil, &chats)
	return chats, err
}

func (c *Client) GetChat(id string) (map[string]any, error) {
	var chat map[string]any
	err := c.request(http.MethodGet, "/api/v1/chats/"+id, nil, &chat)
	return chat, err
}

func (c *Client) DeleteChat(id string) error {
	return c.request(http.MethodDelete, "/api/v1/chats/"+id, nil, nil)
}

type KnowledgeItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type KnowledgeListResponse struct {
	Items []KnowledgeItem `json:"items"`
	Total int             `json:"total"`
}

func (c *Client) ListKnowledge() ([]KnowledgeItem, error) {
	var resp KnowledgeListResponse
	if err := c.request(http.MethodGet, "/api/v1/knowledge/", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) CreateKnowledge(name, description string) (KnowledgeItem, error) {
	var item KnowledgeItem
	err := c.request(http.MethodPost, "/api/v1/knowledge/create", map[string]string{
		"name":        name,
		"description": description,
	}, &item)
	return item, err
}

func (c *Client) DeleteKnowledge(id string) error {
	return c.request(http.MethodDelete, "/api/v1/knowledge/"+id+"/delete", nil, nil)
}

func (c *Client) AddFileToKnowledge(knowledgeID, fileID string) error {
	return c.request(http.MethodPost, "/api/v1/knowledge/"+knowledgeID+"/file/add", map[string]string{
		"file_id": fileID,
	}, nil)
}

type FileUploadResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (c *Client) UploadFile(path string) (FileUploadResponse, error) {
	f, err := os.Open(path)
	if err != nil {
		return FileUploadResponse{}, err
	}
	defer f.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return FileUploadResponse{}, err
	}
	if _, err := io.Copy(part, f); err != nil {
		return FileUploadResponse{}, err
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, c.base+"/api/v1/files/", &body)
	if err != nil {
		return FileUploadResponse{}, err
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return FileUploadResponse{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return FileUploadResponse{}, err
	}
	if resp.StatusCode >= 400 {
		return FileUploadResponse{}, parseAPIError(resp.StatusCode, data)
	}

	var out FileUploadResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return FileUploadResponse{}, err
	}
	return out, nil
}

type FileStatus struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

func (c *Client) FileStatus(id string) (FileStatus, error) {
	var status FileStatus
	err := c.request(http.MethodGet, "/api/v1/files/"+id+"/process/status", nil, &status)
	return status, err
}

func (c *Client) DeleteFile(id string) error {
	return c.request(http.MethodDelete, "/api/v1/files/"+id, nil, nil)
}

func (c *Client) PullModel(name string) error {
	return c.PullModelWithProgress(name, nil)
}

// PullModelWithProgress pulls an Ollama model, optionally reporting status lines.
func (c *Client) PullModelWithProgress(name string, onStatus func(string)) error {
	body := map[string]any{
		"name":   name,
		"stream": onStatus != nil,
	}
	if onStatus == nil {
		return c.request(http.MethodPost, "/ollama/api/pull", body, nil)
	}

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.base+"/ollama/api/pull", bytes.NewReader(b))
	if err != nil {
		return err
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return parseAPIError(resp.StatusCode, data)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var ev struct {
			Status    string `json:"status"`
			Completed int    `json:"completed"`
			Total     int    `json:"total"`
			Error     string `json:"error"`
		}
		if err := json.Unmarshal(line, &ev); err != nil {
			continue
		}
		if ev.Error != "" {
			return fmt.Errorf("%s", ev.Error)
		}
		msg := ev.Status
		if ev.Total > 0 {
			pct := int(float64(ev.Completed) / float64(ev.Total) * 100)
			msg = fmt.Sprintf("%s (%d%%)", ev.Status, pct)
		}
		if msg != "" {
			onStatus(msg)
		}
	}
	return scanner.Err()
}

func (c *Client) ListOllamaModels() ([]map[string]any, error) {
	var resp struct {
		Models []map[string]any `json:"models"`
	}
	if err := c.request(http.MethodGet, "/ollama/api/tags", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Models, nil
}