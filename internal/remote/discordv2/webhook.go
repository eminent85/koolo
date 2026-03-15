package discordv2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// webhookSender delivers messages via Discord webhook HTTP requests.
type webhookSender struct {
	url    string
	client *http.Client
}

func newWebhookSender(url string) *webhookSender {
	return &webhookSender{
		url: strings.TrimSpace(url),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (w *webhookSender) Send(ctx context.Context, content, fileName string, fileData []byte) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("content", content); err != nil {
		writer.Close()
		return fmt.Errorf("webhook: write content field: %w", err)
	}

	if len(fileData) > 0 && fileName != "" {
		part, err := writer.CreateFormFile("file", fileName)
		if err != nil {
			writer.Close()
			return fmt.Errorf("webhook: create file field: %w", err)
		}
		if _, err := part.Write(fileData); err != nil {
			writer.Close()
			return fmt.Errorf("webhook: write file data: %w", err)
		}
	}

	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		return fmt.Errorf("webhook: finalize payload: %w", err)
	}

	return w.doPost(ctx, contentType, &body)
}

func (w *webhookSender) SendEmbed(ctx context.Context, embed *discordgo.MessageEmbed) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	payload := struct {
		Embeds []*discordgo.MessageEmbed `json:"embeds"`
	}{
		Embeds: []*discordgo.MessageEmbed{embed},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		writer.Close()
		return fmt.Errorf("webhook: marshal embed: %w", err)
	}

	if err := writer.WriteField("payload_json", string(data)); err != nil {
		writer.Close()
		return fmt.Errorf("webhook: write embed field: %w", err)
	}

	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		return fmt.Errorf("webhook: finalize embed payload: %w", err)
	}

	return w.doPost(ctx, contentType, &body)
}

func (w *webhookSender) doPost(ctx context.Context, contentType string, body io.Reader) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, body)
	if err != nil {
		return fmt.Errorf("webhook: create request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return nil
}
