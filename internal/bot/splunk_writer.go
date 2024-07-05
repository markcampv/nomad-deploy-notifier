package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/nomad/api"
	"net/http"
	"time"
)

type SplunkConfig struct {
	Token    string
	Endpoint string
}

type SplunkClient struct {
	config SplunkConfig
	client *http.Client
}

func NewSplunkClient(config SplunkConfig) *SplunkClient {
	return &SplunkClient{
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type SplunkEvent struct {
	Time       string      `json:"time"`
	Host       string      `json:"host"`
	Source     string      `json:"source"`
	Sourcetype string      `json:"sourcetype"`
	Event      interface{} `json:"event"`
}

func (sc *SplunkClient) SendEvent(event api.Event) error {
	splunkEvent := SplunkEvent{
		Time:       fmt.Sprintf("%d", time.Now().Unix()),
		Host:       "nomad-client",
		Source:     "nomad-deploy-notifier",
		Sourcetype: "_json",
		Event:      event,
	}

	data, err := json.Marshal(splunkEvent)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", sc.config.Endpoint, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Splunk "+sc.config.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := sc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send event to Splunk: %s", resp.Status)
	}

	return nil
}
