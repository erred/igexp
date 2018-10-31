package function

// package test1
//
// import (
// 	"context"
// 	"log"
// 	"net/http"
// 	"os"
// 	"sync"
//
// 	"cloud.google.com/go/logging"
// )
//
// var o sync.Once
// var logger *logging.Logger
// var slog *log.Logger
// var slog2 *log.Logger
//
// func F(w http.ResponseWriter, r *http.Request) {
// 	o.Do(func() {
// 		client, err := logging.NewClient(context.Background(), os.Getenv("GCP_PROJECT"))
// 		if err != nil {
// 			log.Println(err)
// 		}
//
// 		logger = client.Logger(os.Getenv("FUNCTION_NAME"))
// 		slog = logger.StandardLogger(logging.Info)
// 		slog2 = logger.StandardLogger(logging.Error)
// 	})
// 	slog.Println("Stackdriver logging print")
// 	slog2.Println("Stackdriver logging error?")
// 	err := logger.LogSync(context.Background(), logging.Entry{Payload: "something happened!"})
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	logger.Flush()
// }
