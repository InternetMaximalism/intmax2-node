package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"example.com/m/context"
	"example.com/m/handlers"
)

func main() {
	circuitName := flag.String("circuit", "", "circuit name")
	flag.Parse()

	if *circuitName == "" {
		fmt.Println("Please provide circuit name")
		os.Exit(1)
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx := handlers.CircuitData(context.InitCircuitData(*circuitName))
	http.HandleFunc("/health", handlers.HealthHandler)
	http.HandleFunc("/start-proof", ctx.StartProof)
	http.HandleFunc("/get-proof", ctx.GetProof)
	fmt.Printf("Server is running on port %s...", port)
	addr := fmt.Sprintf("%s:%s", host, port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
