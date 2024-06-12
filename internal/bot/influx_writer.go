package bot

import (
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type InfluxConfig struct {
	Token  string
	URL    string
	Org    string
	Bucket string
}

type InfluxWriter struct {
	mu     sync.Mutex
	client influxdb2.Client
	write  api.WriteAPI
	L      hclog.Logger
}

func NewInfluxWriter(cfg InfluxConfig, client influxdb2.Client) (*InfluxWriter, error) {
	writeAPI := client.WriteAPI(cfg.Org, cfg.Bucket)

	writer := &InfluxWriter{
		client: client,
		write:  writeAPI,
	}

	return writer, nil
}

func (w *InfluxWriter) UpsertDeployMsg(deploy api.Deployment) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	p := influxdb2.NewPointWithMeasurement("nomad_deployment").
		AddTag("deployment_id", deploy.ID).
		AddField("status", deploy.Status).
		AddField("status_description", deploy.StatusDescription).
		SetTime(time.Now())

	for tgn, tg := range deploy.TaskGroups {
		p.AddField(fmt.Sprintf("task_group_%s_healthy", tgn), tg.HealthyAllocs)
		p.AddField(fmt.Sprintf("task_group_%s_placed", tgn), tg.PlacedAllocs)
		p.AddField(fmt.Sprintf("task_group_%s_desired_canaries", tgn), tg.DesiredCanaries)
	}

	w.write.WritePoint(p)
	w.write.Flush()
	return nil
}
