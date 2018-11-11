package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
)

type UserDoc struct {
	Feed  bool
	Story bool
	Tag   bool
}

var bl = map[string][]string{
	"1440598382": []string{"tag", "feed", "story"},
	"28527810":   []string{"tag", "feed", "story"},
	"6507691":    []string{"tag", "feed", "story"},
	"1809062259": []string{"tag", "feed", "story"},
	"1455529947": []string{"tag", "feed", "story"},
	"1352857355": []string{"tag", "feed", "story"},
}

func F(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	dstore, err := datastore.NewClient(ctx, os.Getenv("GCP_PROJECT"))
	if err != nil {
		panic(fmt.Errorf("Login Error: datastore failed: %v", err))
	}
	akey := datastore.NameKey("igtools", "mkeep", nil)

	for k := range bl {
		key := datastore.NameKey("user", k, akey)
		_, err := dstore.Put(context.Background(), key, &UserDoc{true, true, true})
		if err != nil {
			log.Fatal(err)
		}
	}
	w.WriteHeader(200)
}
