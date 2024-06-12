package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/drewbailey/nomad-deploy-notifier/internal/bot"
	"github.com/drewbailey/nomad-deploy-notifier/internal/stream"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func main() {
	os.Exit(realMain(os.Args))
}

func realMain(args []string) int {
	ctx, closer := CtxWithInterrupt(context.Background())
	defer closer()

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

	stream := stream.NewStream()

	influxClient := influxdb2.NewClient(influxURL, influxToken)
	influxWriter, err := bot.NewInfluxWriter(influxCfg, influxClient)
	if err != nil {
		panic(err)
	}

	stream.Subscribe(ctx, influxWriter)

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
