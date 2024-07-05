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

func (s *Stream) Subscribe(ctx context.Context, influxWriter *bot.InfluxWriter, splunkClient *bot.SplunkClient, topicsList []string, jobName string) {
	events := s.nomad.EventStream()

	topicsMap := make(map[api.Topic][]string)
	for _, topic := range topicsList {
		topicsMap[api.Topic(topic)] = []string{"*"}
	}

	eventCh, err := events.Stream(ctx, topicsMap, 0, &api.QueryOptions{})
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
				// Process events and log only if detailed information is available
				if deployment, err := e.Deployment(); err == nil && deployment != nil {
					if jobName == "" || deployment.JobID == jobName {
						if influxWriter != nil {
							if err := influxWriter.UpsertDeployMsg(*deployment); err != nil {
								s.L.Warn("error decoding payload for InfluxDB", "error", err)
							}
						}

						if splunkClient != nil {
							if err := splunkClient.SendEvent(e); err != nil {
								s.L.Warn("error sending event to Splunk", "error", err)
							}
						}
					}
				} else if node, err := e.Node(); err == nil && node != nil {
					if influxWriter != nil {
						// Handle node-related InfluxDB logic if needed
					}

					if splunkClient != nil {
						if err := splunkClient.SendEvent(e); err != nil {
							s.L.Warn("error sending event to Splunk", "error", err)
						}
					}
				} else if job, err := e.Job(); err == nil && job != nil {
					if jobName == "" || *job.ID == jobName {
						if influxWriter != nil {
							// Handle job-related InfluxDB logic if needed
						}

						if splunkClient != nil {
							if err := splunkClient.SendEvent(e); err != nil {
								s.L.Warn("error sending event to Splunk", "error", err)
							}
						}
					}
				}
			}
		}
	}
}
