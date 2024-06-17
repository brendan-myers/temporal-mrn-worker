package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"crypto/tls"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func RegionWorkflow(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 2 * time.Second,
	})

	var region string

	for i := 0; i < 36; i++ { // 36 * 5s sleep = 3 minutes

		err := workflow.ExecuteActivity(ctx, GetRegion).Get(ctx, &region)
		if err != nil {
			slog.Error(err.Error())
			return err
		}

		workflow.Sleep(ctx, time.Second*5)
	}

	return nil
}

func GetRegion(ctx context.Context) (string, error) {
	return os.Getenv("TEMPORAL_REGION"), nil
}

func main() {
	slog.Info("Starting")

	var host, hostport string
	namespace, namespaceSet := os.LookupEnv("TEMPORAL_NAMESPACE")
	if namespaceSet {
		host = namespace + ".tmprl.cloud"
		hostport = namespace + ".tmprl.cloud:7233"
	} else {
		slog.Error("TEMPORAL_NAMESPACE not set")
		return
	}

	slog.Info(host)

	clientKeyPath := os.Getenv("TEMPORAL_TLS_KEY")
	clientCertPath := os.Getenv("TEMPORAL_TLS_CERT")

	cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		slog.Error("Unable to load cert and key pair.", err)
		return
	}

	temporalClient, err := client.Dial(client.Options{
		HostPort:  hostport,
		Namespace: namespace,
		ConnectionOptions: client.ConnectionOptions{
			TLS: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				ServerName:         host,
				InsecureSkipVerify: true,
			},
		},
	})
	if err != nil {
		slog.Error("Unable to create client", err)
		return
	}
	defer temporalClient.Close()

	regionWOrker := worker.New(temporalClient, "mrn-test", worker.Options{})
	regionWOrker.RegisterWorkflow(RegionWorkflow)
	regionWOrker.RegisterActivity(GetRegion)

	err = regionWOrker.Run(worker.InterruptCh())
	if err != nil {
		slog.Error("Unable to start Worker", err)
	}

	slog.Info("Stopping")
}
