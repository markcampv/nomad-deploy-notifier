package stream

import (
	"context"
	"os"

	"github.com/drewbailey/nomad-deploy-notifier/internal/bot"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
)

type Stream struct {
	nomad *api.Client
	L     hclog.Logger
}

func NewStream() *Stream {
	client, _ := api.NewClient(&api.Config{})
	return &Stream{
		nomad: client,
		L:     hclog.Default(),
	}
}

func (s *Stream) Subscribe(ctx context.Context, influxWriter *bot.InfluxWriter, splunkClient *bot.SplunkClient) {
	events := s.nomad.EventStream()

	topics := map[api.Topic][]string{
		api.Topic("Deployment"): {"*"},
		api.Topic("Node"):       {"*"},
		// Add more topics as needed
	}

	eventCh, err := events.Stream(ctx, topics, 0, &api.QueryOptions{})
	if err != nil {
		s.L.Error("error creating event stream client", "error", err)
		os.Exit(1)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-eventCh:
			if event.Err != nil {
				s.L.Warn("error from event stream", "error", err)
				break
			}
			if event.IsHeartbeat() {
				continue
			}

			for _, e := range event.Events {
				// Handle InfluxDB writing if needed
				if err = splunkClient.SendEvent(e); err != nil {
					s.L.Warn("error sending event to Splunk", "error", err)
				}
			}
		}
	}
}
