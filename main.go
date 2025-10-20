package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/miltonparedes/lazycommit/pkg/config"
	"github.com/miltonparedes/lazycommit/pkg/provider"
	"github.com/miltonparedes/lazycommit/pkg/types"
)

func main() {
	fmt.Println("lazycommit - AI-powered commit message generator")
	fmt.Println("Initializing...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	providerName := os.Getenv("LAZYCOMMIT_PROVIDER")
	if providerName == "" {
		providerName = cfg.DefaultProvider
	}

	aiProvider, err := provider.NewFromConfig(cfg, providerName)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	fmt.Printf("Using provider: %s\n", aiProvider.Name())
	fmt.Printf("Available models: %v\n", aiProvider.Models())

	ctx := context.Background()
	req := types.CompletionRequest{
		Messages: []types.Message{
			{
				Role:    "user",
				Content: "Say hello in a friendly way!",
			},
		},
		Temperature: 0.7,
		MaxTokens:   100,
		Model:       aiProvider.Models()[0],
	}

	fmt.Println("\nTesting completion...")
	resp, err := aiProvider.Complete(ctx, req)
	if err != nil {
		log.Fatalf("Completion failed: %v", err)
	}

	fmt.Printf("Response: %s\n", resp.Content)
	fmt.Printf("Usage - Prompt: %d, Completion: %d, Total: %d tokens\n",
		resp.Usage.PromptTokens,
		resp.Usage.CompletionTokens,
		resp.Usage.TotalTokens,
	)

	fmt.Println("\nTesting streaming...")
	stream, err := aiProvider.Stream(ctx, req)
	if err != nil {
		log.Fatalf("Streaming failed: %v", err)
	}

	fmt.Print("Stream: ")
	for chunk := range stream {
		if chunk.Error != nil {
			log.Fatalf("Stream error: %v", chunk.Error)
		}
		if chunk.Done {
			fmt.Println("\n\nStream completed!")
			break
		}
		fmt.Print(chunk.Content)
	}
}
