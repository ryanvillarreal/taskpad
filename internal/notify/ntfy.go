package notify

import (
	"fmt"
	"net/http"
	"strings"
)

type NtfyNotifier struct {
	URL   string
	Topic string
}

func NewNtfy(url, topic string) *NtfyNotifier {
	return &NtfyNotifier{URL: url, Topic: topic}
}

func (n *NtfyNotifier) Send(title, body string) error {
	endpoint := fmt.Sprintf("%s/%s", strings.TrimRight(n.URL, "/"), n.Topic)
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Title", title)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy: status %d", resp.StatusCode)
	}
	return nil
}
