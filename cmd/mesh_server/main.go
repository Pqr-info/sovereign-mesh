package main

import (
	"context"
	"github.com/pqr-info/sovereign-mesh"
)

func main() {
	swarm := sovereign.NewController("my-project", "us-central1")
	swarm.InitMemoryBus()
	swarm.Start(context.Background())

	// Block main thread to keep gRPC server alive
	select {}
}
