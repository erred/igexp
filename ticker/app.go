package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/pubsub"
)

var (
	client *pubsub.Client

	envPort    = os.Getenv("PORT")
	envProject = os.Getenv("GCP_PROJECT")
)

func init() {
	var err error

	if client, err = pubsub.NewClient(context.Background(), envProject); err != nil {
		log.Fatal("Init: Create PubSub Client failed: ", err)
	}
}

func main() {
	http.HandleFunc("/fwatch", fwatchHandler)
	http.HandleFunc("/", http.NotFoundHandler())

	port := envPort
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func fwatchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, err := client.Topic("igtools-tick").Publish(ctx, &pubsub.Message{Data: []byte("app engine trigger fwatch")}).Get(ctx)
	if err != nil {
		log.Println("Fwatch trigger error: ", err)
	}
}
