package main

import (
	"context"
	"log"
	"time"

	"github.com/pqr-info/sovereign-mesh/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:1111", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewAgentToolUseClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Println("Testing ExecuteBrowserAuth...")
	resp, err := client.ExecuteBrowserAuth(ctx, &proto.BrowserAuthRequest{TargetUrl: "https://gemini.google.com/"})
	if err != nil {
		log.Fatalf("ExecuteBrowserAuth failed: %v", err)
	}

	log.Printf("ExecuteBrowserAuth Success: %v, Token: %s", resp.Success, resp.SessionToken)
}
