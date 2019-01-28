package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/ahmdrz/goinsta"
)

var (
	IG_USER   = "INSTA_USER"
	IG_PASSWD = "INSTA_PASS"

	FILENAME           = "follows.json"
	FILENAME_EVENTS    = "events.json"
	UPDATE_INTERVAL, _ = time.ParseDuration("6h")
)

func main() {
	ig := goinsta.New(os.Getenv(IG_USER), os.Getenv(IG_PASSWD))
	if err := ig.Login(); err != nil {
		log.Fatal("Failed to login: ", err)
	}

	f := newFollow(ig)
	ig.Save()

	for range time.Tick(UPDATE_INTERVAL) {
		f.update()
		ig.Save()
	}
}

type follow struct {
	ig *goinsta.Instagram

	Followers map[int64]goinsta.User
	Following map[int64]goinsta.User

	Events []event
}

type event struct {
	Timestamp time.Time
	Event     string
	ID        int64
	Username  string
	Name      string
}

func newFollow(ig *goinsta.Instagram) *follow {
	f := &follow{
		ig,
		getAllUsers(ig.Account.Followers()),
		getAllUsers(ig.Account.Following()),
		[]event{},
	}

	f.save()
	return f
}

func (f *follow) update() {
	events := []event{}
	followers := getAllUsers(f.ig.Account.Followers())
	events = append(events, diff(f.Followers, followers, "Lost Follower")...)
	events = append(events, diff(followers, f.Followers, "Gained Follower")...)
	f.Followers = followers

	following := getAllUsers(f.ig.Account.Following())
	events = append(events, diff(f.Following, following, "Stopped Following")...)
	events = append(events, diff(following, f.Following, "Started Following")...)
	f.Following = following

	for _, e := range events {
		log.Println(e)
	}
	f.Events = append(f.Events, events...)
	f.save()
}

func (f *follow) save() {
	log.Println("Followers: ", len(f.Followers), " Following: ", len(f.Following))

	b, err := json.MarshalIndent(f, "", "\t")
	if err != nil {
		log.Println("Failed to marshal data: ", err)
		return
	}
	if err := ioutil.WriteFile(FILENAME, b, 0644); err != nil {
		log.Println("Failed to save file: ", err)
	}

	b, err = json.MarshalIndent(f.Events, "", "\t")
	if err != nil {
		log.Println("Failed to marshal data: ", err)
		return
	}
	if err := ioutil.WriteFile(FILENAME_EVENTS, b, 0644); err != nil {
		log.Println("Failed to save file: ", err)
	}
}

func diff(old, cur map[int64]goinsta.User, ev string) []event {
	events := []event{}
	for k, v := range old {
		if _, ok := cur[k]; !ok {
			events = append(events, event{
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

func getAllUsers(users *goinsta.Users) map[int64]goinsta.User {
	if err := users.Error(); err != nil {
		log.Println("Unknown error users initial page: ", err)
	}
	m := map[int64]goinsta.User{}
	for users.Next() {
		if err := users.Error(); err != nil {
			if err == goinsta.ErrNoMore {
				return m
			}
			log.Println("Unknown error users next page: ", err)
		}
		for _, u := range users.Users {
			m[u.ID] = u
		}
	}
	return m
}
