package trawl

import (
	"log"
	"net/http"
	"time"

	"github.com/seankhliao/keepig/insta"
)

// Trawler object makes requests to get media
type Trawler struct {
	refresh time.Duration // seconds between automatic updates
	tick    *time.Ticker

	ig       *insta.Instagram
	client   *http.Client
	throttle time.Duration // milliseconds between each request

	store *Store
}

// NewTrawler creates a zero initialized trawler
func NewTrawler(refresh int64, throttle int64) *Trawler {
	t := Trawler{
		refresh: time.Duration(refresh) * time.Second,

		ig:       insta.NewInstagram("", ""),
		client:   &http.Client{},
		throttle: time.Duration(throttle) * time.Millisecond,

		store: NewStore(),
	}
	t.client.Transport = &t.ig.Transport
	return &t
}

// RestoreState from disk
func (t *Trawler) RestoreState() {

}

// SaveState to disk
func (t *Trawler) SaveState() {

}

func (t *Trawler) getClient() *http.Client {
	t.client.Jar = t.ig.Instagram.Cookiejar
	return t.client
}

// Schedule a trawl to occur every N time
func (t *Trawler) Schedule() {
	t.tick = time.NewTicker(t.refresh)
	for range t.tick.C {
		t.Trawl()
	}
}

// Trawl a userlist for media and saves
func (t *Trawler) Trawl() {
	if err := t.getUsers(); err != nil {
		log.Println(err)
	}
	if err := t.getFeeds(); err != nil {
		log.Println(err)
	}
}

// getUsers gets list of accounts you wish to trawl
func (t *Trawler) getUsers() error {
	users, err := t.ig.GetFollowersAll()
	if err != nil {
		return err
	}
	t.store.Users = users
	return nil
}

// gets the 3 feeds: main, tagged, reel
func (t *Trawler) getFeeds() error {
	for _, user := range t.store.Users {
		feed := t.ig.CreateFeed(user.ID)

		for more := true; more; {
			medias, err := feed.MainNext()
			if err != nil {
				// TODO handle
				// might be empty
			}
			for _, media := range medias {
				if _, ok := t.store.feeds[user.ID][media.ID]; ok {
					more = false
					break
				}
				b, err := media.Get(t.getClient())
				if err != nil {
					log.Println("error getting media")
				}
				// TODO store the media b
			}
		}

		for more := true; more; {
			medias, err := feed.TaggedNext()
			if err != nil {
				// TODO handle
				// might be empty
			}
			for _, media := range medias {
				if _, ok := t.store.feeds[user.ID][media.ID]; ok {
					more = false
					break
				}
				b, err := media.Get(t.getClient())
				if err != nil {
					log.Println("error getting media")
				}
				// TODO store the media b
			}
		}

		// TODO reelfeed

	}
	return nil
}
