package stream

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

type Delta struct {
	Content string
	Done    bool
	Error   string
}

func ParseChatCompletions(r io.Reader, onDelta func(string) error) (string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var full strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			break
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
				Text string `json:"text"`
			} `json:"choices"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
			Detail any `json:"detail"`
		}

		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}

		if chunk.Error != nil && chunk.Error.Message != "" {
			return full.String(), errFromDetail(chunk.Error.Message)
		}
		if chunk.Detail != nil {
			return full.String(), errFromDetail(chunk.Detail)
		}

		for _, choice := range chunk.Choices {
			text := choice.Delta.Content
			if text == "" {
				text = choice.Message.Content
			}
			if text == "" {
				text = choice.Text
			}
			if text == "" {
				continue
			}
			full.WriteString(text)
			if onDelta != nil {
				if err := onDelta(text); err != nil {
					return full.String(), err
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return full.String(), err
	}
	return full.String(), nil
}

func errFromDetail(d any) error {
	switch v := d.(type) {
	case string:
		return &StreamError{Msg: v}
	default:
		b, _ := json.Marshal(d)
		return &StreamError{Msg: string(b)}
	}
}

type StreamError struct {
	Msg string
}

func (e *StreamError) Error() string {
	return e.Msg
}