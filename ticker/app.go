package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"cloud.google.com/go/pubsub"
)

var (
	client *pubsub.Client

	envPort    = os.Getenv("PORT")
	envProject = os.Getenv("GOOGLE_CLOUD_PROJECT")
)

func init() {
	var err error

	if client, err = pubsub.NewClient(context.Background(), envProject); err != nil {
		log.Fatal("Init: Create PubSub Client failed: ", err)
	}
}

func main() {
	http.HandleFunc("/", defaultHandler)

	port := envPort
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Started on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Appengine-Cron") != "true" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	ctx := context.Background()
	topic := path.Base(r.URL.Path)

	if _, err := client.Topic(topic).Publish(ctx, &pubsub.Message{Data: []byte("ping!")}).Get(ctx); err != nil {
		log.Println("Publish failed: ", err)
		return
	}
	fmt.Println("Publish empty message to ", topic)
}
