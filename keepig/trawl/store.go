package trawl

import (
	"github.com/seankhliao/keepig/insta"
)

// Item Media types
const (
	IMAGE    = 1
	VIDEO    = 2
	CAROUSEL = 8
)

// Item States
const (
	WAITING = "waiting"
	DONE    = "done"
)

// Item represents a unit of work
type Item struct {
	insta.Media
	State  string
	UserID int64
}

// Store represents a store of data
type Store struct {
	Users []insta.User
	feeds map[int64]map[string]Item
}

// NewStore creates an empty store
func NewStore() *Store {
	return &Store{}
}

// AddItem stores an Item
func (s *Store) AddItem(item Item) {
	s.feeds[item.UserID][item.Media.ID] = item
}
