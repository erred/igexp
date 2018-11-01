package test1

import (
	"context"
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
		r := recover()
		log.Printf("function panicked, recovered with: %v", r)
	}()
	o.Do(func() {

		store, err := storage.NewClient(context.Background())
		if err != nil {
			log.Println(err)
		}
		bucket = store.Bucket("igtools-storage")
	})

	w.Write([]byte("\nRead object:\n"))

	re, err := bucket.Object("test-object").NewReader(context.Background())
	if err != nil {
		log.Println(err)
	}
	defer re.Close()
	_, err = io.Copy(w, re)
	if err != nil {
		log.Println(err)
	}

	w.Write([]byte("\nRead nested object:\n"))

	re, err = bucket.Object("nested/test-object").NewReader(context.Background())
	if err != nil {
		log.Println(err)
	}
	defer re.Close()
	_, err = io.Copy(w, re)
	if err != nil {
		log.Println(err)
	}
	//
	// w.Write([]byte("\nRead non-existent object:\n"))
	//
	// re, err = bucket.Object("non-existent-object").NewReader(context.Background())
	// if err != nil {
	// 	log.Println(err)
	// }
	// defer re.Close()
	// _, err = io.Copy(w, re)
	// if err != nil {
	// 	log.Println(err)
	// }
}
