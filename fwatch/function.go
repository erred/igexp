package fwatch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/seankhliao/igtools/goinsta"

	"cloud.google.com/go/storage"
)

var (
	c = client{}

	// names of environment variables
	envBucket  = "BUCKET"
	envGoinsta = "GOINSTA"
	envPass    = "PASS"
	envProject = "GCP_PROJECT"
	envUser    = "USER"

	// names of storage objects
	objEvents    = "fwatch/events.json"
	objFollowers = "fwatch/followers.json"
	objFollowing = "fwatch/following.json"
	objGoinsta   = "fwatch/goinsta.json"
)

type client struct {
	bucket *storage.BucketHandle
	ig     *goinsta.Instagram
	once   sync.Once
}

func (cl *client) Login() {
	var err error
	ctx := context.Background()

	// storage
	store, err := storage.NewClient(ctx)
	if err != nil {
		panic(fmt.Errorf("Login Error: cloud storage failed: %v", err))
	}
	cl.bucket = store.Bucket(envBucket)

	// goinsta
	if os.Getenv(envGoinsta) != "" {
		f, err := ioutil.TempFile("", "goinsta")
		if err != nil {
			panic(fmt.Errorf("Login Error: create temp file failed: %v", err))
		}
		cl.ig, err = goinsta.Import(f.Name())
		if err != nil {
			panic(fmt.Errorf("Login Error: import goinsta state failed: %v", err))
		}
		if err = os.Remove(f.Name()); err != nil {
			log.Println("failed to clean up goinsta restore file: ", err)
		}
		fmt.Println("Logged in with restore")
	} else {
		cl.ig = goinsta.New(os.Getenv(envUser), os.Getenv(envPass))
		if err := cl.ig.Login(); err != nil {
			panic(fmt.Errorf("Login Error: goinsta failed: %v", err))
		}
	}
}

// Fwatch is the entrypoint and pubsub handler
func Fwatch(ctx context.Context, msg struct{}) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("panic: ", r)
		}
	}()

	// ensure login
	c.once.Do(c.Login)

	f := follow{}

	if err := f.restore(); err != nil {
		log.Printf("Restore failed: %v", err)
	}

	if err := f.update(); err != nil {
		log.Printf("Update failed: %v", err)
	} else {
		if err := f.save(); err != nil {
			log.Printf("Save failed: %v", err)
		}
	}

	return nil
}

type follow struct {
	Followers map[int64]goinsta.User
	Following map[int64]goinsta.User

	Events []Event
}

// Event logs a follow/unfollow event
type Event struct {
	Timestamp time.Time
	Event     string
	ID        int64
	Username  string
	Name      string
}

// restore restores previous list of follows from storage
func (f *follow) restore() error {
	ctx := context.Background()

	r, err := c.bucket.Object(objEvents).NewReader(ctx)
	switch err {
	case nil:
		defer r.Close()
		if err := json.NewDecoder(r).Decode(&f.Events); err != nil {
			return fmt.Errorf("%v decode error: %v", objEvents, err)
		}
	case storage.ErrObjectNotExist:
	default:
		return fmt.Errorf("%v reader error: %v", objEvents, err)
	}

	r, err = c.bucket.Object(objFollowers).NewReader(ctx)
	switch err {
	case nil:
		defer r.Close()
		if err := json.NewDecoder(r).Decode(&f.Followers); err != nil {
			return fmt.Errorf("%v decode error: %v", objFollowers, err)
		}
	case storage.ErrObjectNotExist:
	default:
		return fmt.Errorf("%v reader error: %v", objFollowers, err)
	}

	r, err = c.bucket.Object(objFollowing).NewReader(ctx)
	switch err {
	case nil:
		defer r.Close()
		if err := json.NewDecoder(r).Decode(&f.Following); err != nil {
			return fmt.Errorf("%v decode error: %v", objFollowing, err)
		}
	case storage.ErrObjectNotExist:
	default:
		return fmt.Errorf("%v reader error: %v", objFollowing, err)
	}

	return nil
}

// update gets the current follows and calculates the changes
func (f *follow) update() error {
	cfwer, err := getAllUsers(c.ig.Account.Followers())
	if err != nil {
		return fmt.Errorf("getAllUsers followers error: %v", err)
	}
	f.Events = append(f.Events, diff(f.Followers, cfwer, "Lost Follower")...)
	f.Events = append(f.Events, diff(cfwer, f.Followers, "Gained Follower")...)
	f.Followers = cfwer

	cfwing, err := getAllUsers(c.ig.Account.Following())
	if err != nil {
		return fmt.Errorf("getAllUsers following error: %v", err)
	}
	f.Events = append(f.Events, diff(f.Following, cfwing, "Stopped Following")...)
	f.Events = append(f.Events, diff(cfwing, f.Following, "Started Following")...)
	f.Following = cfwing

	return nil
}

// save saves the current follows back to storage
func (f *follow) save() {
	ctx := context.Background()

	w := c.bucket.Object(objEvents).NewWriter(ctx)
	defer w.Close()
	if err := json.NewEncoder(w).Encode(f.Events); err != nil {
		log.Printf("Saving %v failed: %v", objEvents, err)
	}

	w = c.bucket.Object(objFollowers).NewWriter(ctx)
	defer w.Close()
	if err := json.NewEncoder(w).Encode(f.Followers); err != nil {
		log.Printf("Saving %v failed: %v", objFollowers, err)
	}

	w = c.bucket.Object(objFollowing).NewWriter(ctx)
	defer w.Close()
	if err := json.NewEncoder(w).Encode(f.Following); err != nil {
		log.Printf("Saving %v failed: %v", objFollowing, err)
	}

	// Goinsta state
	f, err := ioutil.TempFile("", "goinsta")
	if err != nil {
		log.Printf("Saving goinsta state failed: %v", err)
		return
	}
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			log.Println("Cleanup goinsta state file from save failed: ", err)
		}
	}()
	if err := c.ig.Export(f.Name()); err != nil {
		log.Println("Goinsta export failed: ", err)
		return
	}
	w = c.bucket.Object(objGoinsta).NewWriter(ctx)
	defer w.Close()
	_, err = io.Copy(w, f)
	if err != nil {
		log.Println("Uploading goinsta state failed: ", err)
		return
	}

}

func diff(old, cur map[int64]goinsta.User, ev string) []Event {
	events := []Event{}
	for k, v := range old {
		if _, ok := cur[k]; !ok {
			events = append(events, Event{
				time.Now(),
				ev,
				v.ID,
				v.Username,
				v.FullName,
			})
		}
	}
	return events
}

func getAllUsers(users *goinsta.Users) (map[int64]goinsta.User, error) {
	if err := users.Error(); err != nil {
		return nil, fmt.Errorf("Unknown error users initial page: %v", err)
	}
	m := map[int64]goinsta.User{}
	for users.Next() {
		for _, u := range users.Users {
			m[u.ID] = u
		}
		if err := users.Error(); err != nil {
			if err == goinsta.ErrNoMore {
				return m, nil
			}
			return m, fmt.Errorf("Unknown error users next page: %v", err)
		}
	}
	return m, nil
}
