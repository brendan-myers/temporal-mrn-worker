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
	slog.Info("Worklow started")
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: 1 * time.Minute,
	})

	var region string

	for !workflow.GetInfo(ctx).GetContinueAsNewSuggested() {
		slog.Info("Executing 'GetRegion' activity")
		err := workflow.ExecuteActivity(ctx, GetRegion).Get(ctx, &region)
		if err != nil {
			slog.Error(err.Error())
			return err
		}
		slog.Info("Execution finished: " + region)
	}

	slog.Info("Workflow finished - continue as new")
	return workflow.NewContinueAsNewError(ctx, RegionWorkflow)
}

func GetRegion(ctx context.Context) (string, error) {
	slog.Info(("Activity started"))
	time.Sleep(15 * time.Second)
	slog.Info("Activity finished")
	return os.Getenv("TEMPORAL_REGION"), nil
}

func main() {
	slog.Info("Starting worker")

	var host, hostport string
	namespace, namespaceSet := os.LookupEnv("TEMPORAL_NAMESPACE")
	if namespaceSet {
		host = namespace + ".tmprl.cloud"
		hostport = namespace + ".tmprl.cloud:7233"
	} else {
		slog.Error("TEMPORAL_NAMESPACE not set")
		return
	}

	slog.Info("Connecting to " + host)

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

	slog.Info("Stopping worker")
}
