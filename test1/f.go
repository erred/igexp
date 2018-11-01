package test1

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"cloud.google.com/go/storage"
)

var o sync.Once
var bucket *storage.BucketHandle

func F(w http.ResponseWriter, r *http.Request) {
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

	re, err := bucket.Object("nested/test-object").NewReader(context.Background())
	if err != nil {
		log.Println("Read nested object: ", err)
	}
	defer re.Close()
	_, err = io.Copy(w, re)
	if err != nil {
		log.Println("Copy nested object: ", err)
	}
	//
	w.Write([]byte("\nRead non-existent object:\n"))

	re, err = bucket.Object("non-existent-object").NewReader(context.Background())
	if err != nil {
		log.Println("Read non existent object: ", err)
	} else {
		defer re.Close()
		_, err = io.Copy(w, re)
		if err != nil {
			log.Println("Copy non existent object: ", err)
		}

	}
}
