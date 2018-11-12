package f

import (
	"context"
	"fmt"
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

	// for k := range bl {
	// 	key := datastore.NameKey("igtools-user", k, nil)
	// 	_, err := dstore.Put(context.Background(), key, &UserDoc{false, false, false})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }
	// akey := datastore.NameKey("igtools", "mkeep", nil)
	total, err := dstore.Count(ctx, datastore.NewQuery("media").KeysOnly())
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("error: " + err.Error()))
		return
	}
	for i := 0; i < total/400+1; i++ {
		q := datastore.NewQuery("media").KeysOnly().Limit(400)
		keys, err := dstore.GetAll(context.Background(), q, nil)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("error: " + err.Error()))
			return
		}
		err = dstore.DeleteMulti(ctx, keys)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("error: " + err.Error()))
			return
		}
		w.Write([]byte("deleted 400\n"))

	}
	w.WriteHeader(200)
}
