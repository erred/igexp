package test1

import (
	"context"
	"fmt"
	"log"
	"sync"

	"cloud.google.com/go/storage"
)

var o sync.Once
var bucket *storage.BucketHandle

func F(ctx context.Context, data struct{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("function panicked, recovered with: %v", r)
		}
	}()

	o.Do(func() {
		store, err := storage.NewClient(context.Background())
		if err != nil {
			log.Println(err)
		}
		bucket = store.Bucket("igtools-storage")
		fmt.Println("Connected to bucket")
	})

	obj := bucket.Object("ps-test")

	w := obj.NewWriter(ctx)
	defer w.Close()
	_, err := w.Write([]byte("pubsub test success!"))
	if err != nil {
		log.Println("write error: ", err)
	}
}
