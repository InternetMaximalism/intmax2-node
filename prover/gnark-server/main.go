package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"gnark-server/context"
	"gnark-server/handlers"
)

func main() {
	circuitName := flag.String("circuit", "", "circuit name")
	flag.Parse()

	if *circuitName == "" {
		fmt.Println("Please provide circuit name")
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Preparing server with circuit: %s\n", *circuitName)
	ctx := handlers.CircuitData(context.InitCircuitData(*circuitName))
	http.HandleFunc("/health", handlers.HealthHandler)
	http.HandleFunc("/start-proof", ctx.StartProof)
	http.HandleFunc("/get-proof", ctx.GetProof)
	fmt.Printf("Server is running on port %s...", port)
	addr := fmt.Sprintf(":%s", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
