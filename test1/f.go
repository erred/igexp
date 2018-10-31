package test1

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"cloud.google.com/go/logging"
)

var o = sync.Once
var logger *logging.Logger
var slog *log.Logger
var slog2 *log.Logger

func F(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Output to STDOUT")
	log.Println("Output to STDERR")

	defer func() {
		r := recover()
		fmt.Println("recovered")
		fmt.Println(r)
	}()

	o.Do(func() {
		client, err := logging.NewClient(context.Background(), os.Getenv("GCP_PROJECT"))
		if err != nil {
			log.Println(err)
		}

		logger = client.Logger("functions-test1-id")
		slog = logger.StandardLogger(logging.Info)
		slog2 = logger.StandardLogger(logging.Error)
	})

	slog.Println("Stackdriver logging print")
	slog2.Println("Stackdriver logging error?")
	slog.Panicln("Stackdriver logging panic")

}
