package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/drewbailey/nomad-deploy-notifier/internal/bot"
	"github.com/drewbailey/nomad-deploy-notifier/internal/stream"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var (
	useInfluxDB bool
	useSplunk   bool
	topics      string
	jobName     string
)

func init() {
	flag.BoolVar(&useInfluxDB, "influxdb", false, "Send data to InfluxDB")
	flag.BoolVar(&useSplunk, "splunk", false, "Send data to Splunk HEC")
	flag.StringVar(&topics, "topics", "Deployment,Node,Job", "Comma-separated list of topics to send to Splunk")
	flag.StringVar(&jobName, "job_name", "", "Name of the job to filter events")
}

func main() {
	flag.Parse()

	if !useInfluxDB && !useSplunk {
		fmt.Println("Please specify at least one output using the -influxdb or -splunk flags")
		os.Exit(1)
	}

	os.Exit(realMain(os.Args))
}

func realMain(args []string) int {
	ctx, closer := CtxWithInterrupt(context.Background())
	defer closer()

	var influxWriter *bot.InfluxWriter
	var splunkClient *bot.SplunkClient

	if useInfluxDB {
		influxToken := os.Getenv("INFLUXDB_TOKEN")
		influxURL := os.Getenv("INFLUXDB_URL")
		influxOrg := os.Getenv("INFLUXDB_ORG")
		influxBucket := os.Getenv("INFLUXDB_BUCKET")

		influxCfg := bot.InfluxConfig{
			Token:  influxToken,
			URL:    influxURL,
			Org:    influxOrg,
			Bucket: influxBucket,
		}

		influxClient := influxdb2.NewClient(influxURL, influxToken)
		var err error
		influxWriter, err = bot.NewInfluxWriter(influxCfg, influxClient)
		if err != nil {
			panic(err)
		}
	}

	if useSplunk {
		splunkToken := os.Getenv("SPLUNK_HEC_TOKEN")
		splunkEndpoint := os.Getenv("SPLUNK_HEC_ENDPOINT")

		splunkCfg := bot.SplunkConfig{
			Token:    splunkToken,
			Endpoint: splunkEndpoint,
		}

		splunkClient = bot.NewSplunkClient(splunkCfg)
	}

	topicsList := strings.Split(strings.ToLower(topics), ",")
	for i := range topicsList {
		topicsList[i] = strings.Title(strings.TrimSpace(topicsList[i]))
	}

	stream := stream.NewStream()

	stream.Subscribe(ctx, influxWriter, splunkClient, topicsList, jobName)

	return 0
}

func CtxWithInterrupt(ctx context.Context) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	return ctx, func() {
		signal.Stop(ch)
		cancel()
	}
}
