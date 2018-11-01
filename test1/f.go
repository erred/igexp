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
	o.Do(func() {

		store, err := storage.NewClient(context.Background())
		if err != nil {
			log.Println(err)
		}
		bucket = store.Bucket("igtools-storage")
	})

	obj := bucket.Object("test-object").NewWriter(context.Background())
	defer obj.Close()
	_, err := obj.Write([]byte("hello world\n"))
	if err != nil {
		log.Println(err)
	}

	obj = bucket.Object("nested/test-object").NewWriter(context.Background())
	defer obj.Close()
	_, err = obj.Write([]byte("hello nested world\n"))
	if err != nil {
		log.Println(err)
	}

	w.Write([]byte("Read object:\n"))

	re, err := bucket.Object("test-object").NewReader(context.Background())
	if err != nil {
		log.Println(err)
	}
	defer re.Close()
	_, err = io.Copy(w, re)
	if err != nil {
		log.Println(err)
	}

	w.Write([]byte("Read nested object:\n"))

	re, err = bucket.Object("nested/test-object").NewReader(context.Background())
	if err != nil {
		log.Println(err)
	}
	defer re.Close()
	_, err = io.Copy(w, re)
	if err != nil {
		log.Println(err)
	}

	w.Write([]byte("Read non-existent object:\n"))

	re, err = bucket.Object("non-existent-object").NewReader(context.Background())
	if err != nil {
		log.Println(err)
	}
	defer re.Close()
	_, err = io.Copy(w, re)
	if err != nil {
		log.Println(err)
	}
}
