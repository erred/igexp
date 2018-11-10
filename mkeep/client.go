package mkeep

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/seankhliao/igtools/goinsta"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
)

// Blacklist of media feeds to keep
// {
//	userID : ["tagged", "feed"]
// }
// story, feed, tag
var (
	blacklistStory = "story"
	blacklistFeed  = "feed"
	blacklistTag   = "tag"
)

// Blacklist of media sources
type Blacklist map[int64][]string

// Downlist is a list of media we already have
type Downlist map[int64]map[string]struct{}

// Client is a singleton of clients
type Client struct {
	bucket    *storage.BucketHandle
	ig        *goinsta.Instagram
	once      sync.Once
	topic     *pubsub.Topic
	downlist  Downlist
	blacklist Blacklist
}

func (c *Client) setup() {
	var err error
	ctx := context.Background()

	// storage
	store, err := storage.NewClient(ctx)
	if err != nil {
		panic(fmt.Errorf("Login Error: cloud storage failed: %v", err))
	}
	c.bucket = store.Bucket(os.Getenv(envBucket))

	// pubsub
	psc, err := pubsub.NewClient(ctx, os.Getenv(envProject))
	if err != nil {
		panic(fmt.Errorf("Login Error: pubsub client failed:%v", err))
	}
	c.topic = psc.Topic(os.Getenv(envTopic))

	// goinsta
	r, err := c.bucket.Object(objGoinsta).NewReader(ctx)
	if err != nil {
		panic(fmt.Errorf("Login Error: import %v failed: %v", objGoinsta, err))
	}
	defer r.Close()

	c.ig, err = goinsta.ImportReader(r)
	if err != nil {
		panic(fmt.Errorf("Import Error: %v", err))
	}

	// downlist
	r, err = c.bucket.Object(objDownlist).NewReader(ctx)
	switch err {
	case nil:
		defer r.Close()
		if err := json.NewDecoder(r).Decode(c.downlist); err != nil {
			panic(fmt.Errorf("%v decode error: %v", objDownlist, err))
		}
	case storage.ErrObjectNotExist:
		c.downlist = map[int64]map[string]struct{}{}
	default:
		panic(fmt.Errorf("%v reader error: %v", objDownlist, err))
	}

	// blacklist
	r, err = c.bucket.Object(objBlacklist).NewReader(ctx)
	switch err {
	case nil:
		defer r.Close()
		if err := json.NewDecoder(r).Decode(&c.blacklist); err != nil {
			panic(fmt.Errorf("%v decode error: %v", objBlacklist, err))
		}
		fmt.Println("Successfully imported blacklist.json")
	case storage.ErrObjectNotExist:
	default:
		panic(fmt.Errorf("%v reader error: %v", objBlacklist, err))
	}

}

// save saves the current follows back to storage
func (c *Client) save() error {
	ctx := context.Background()

	// Goinsta state
	w := c.bucket.Object(objGoinsta).NewWriter(ctx)
	defer w.Close()
	if err := goinsta.Export(c.ig, w); err != nil {
		return fmt.Errorf("goinsta export failed: %v", err)
	}

	// downlist
	w = c.bucket.Object(objGoinsta).NewWriter(ctx)
	defer w.Close()
	if err := json.NewEncoder(w).Encode(c.downlist); err != nil {
		return fmt.Errorf("downlist export failed: %v", err)
	}
	return nil
}

func (c *Client) isBlacklisted(userID int64, feed string) bool {
	list, ok := c.blacklist[userID]
	if !ok {
		return false
	}
	for _, it := range list {
		if it == feed {
			return true
		}
	}
	return false
}

func (c *Client) isDownloaded(userID int64, mediaID string) bool {
	u, ok := c.downlist[userID]
	if !ok {
		return false
	}
	_, ok = u[mediaID]
	if !ok {
		return false
	}
	return true
}

func (c *Client) getUsers() error {
	following := c.ig.Account.Following()
	for following.Next() {
		for _, user := range following.Users {
			if !c.isBlacklisted(user.ID, "") {
				c.queueUser(user)
			}
		}
	}
	return nil
}

func (c *Client) getFeeds(msg Message) error {
	breakout := false
	user := goinsta.User{}
	following := c.ig.Account.Following()
	for following.Next() {
		for _, us := range following.Users {
			if us.ID == msg.UserID {
				user = us
				breakout = true
				break
			}
		}
		if breakout {
			break
		}
	}

	if !c.isBlacklisted(user.ID, blacklistStory) {
		feed := user.Stories()
		for feed.Next() {
			c.getItems(feed.Items)
		}
	}
	if !c.isBlacklisted(user.ID, blacklistFeed) {
		feed := user.Feed()
		for feed.Next() {
			c.getItems(feed.Items)
		}

	}
	if !c.isBlacklisted(user.ID, blacklistTag) {
		feed, err := user.Tags([]byte{})
		if err != nil {
			log.Printf("get tagged for %v, %v error: %v", user.ID, user.Username, err)
			return nil
		}
		for feed.Next() {
			c.getItems(feed.Items)
		}
	}
	return nil
}

func (c *Client) getItems(items []goinsta.Item) {
	for _, item := range items {
		if len(item.CarouselMedia) != 0 {
			breakout := false
			for _, it := range item.CarouselMedia {
				if c.isDownloaded(item.User.ID, it.ID) {
					breakout = true
					break
				}
				c.queueItem(it)
			}
			if breakout {
				break
			}
			continue
		}

		if c.isDownloaded(item.User.ID, item.ID) {
			break
		}
		c.queueItem(item)
	}
}

func (c *Client) queueUser(user goinsta.User) {
	buf := bytes.Buffer{}
	if err := json.NewEncoder(&buf).Encode(newUserMessage(user)); err != nil {
		log.Println("queueUser encode: ", err)
	}

	ctx := context.Background()
	if _, err := c.topic.Publish(ctx, &pubsub.Message{Data: buf.Bytes()}).Get(ctx); err != nil {
		log.Println("queueUser publish: ", err)
	}

}

func (c *Client) queueItem(item goinsta.Item) {
	buf := bytes.Buffer{}
	if err := json.NewEncoder(&buf).Encode(newItemMessage(item)); err != nil {
		log.Println("queueItem encode: ", err)
	}

	ctx := context.Background()
	if _, err := c.topic.Publish(ctx, &pubsub.Message{Data: buf.Bytes()}).Get(ctx); err != nil {
		log.Println("queueItem publish: ", err)
	}

}

func (c *Client) download(msg Message) error {
	resp, err := c.ig.Client().Get(msg.Url)
	if err != nil {
		return fmt.Errorf("Download item %v failed: %v", msg.ItemID, err)
	}
	defer resp.Body.Close()

	obj := path.Join(objBase, "media", strconv.FormatInt(msg.UserID, 10), msg.ItemID+msg.Ext)
	w := c.bucket.Object(obj).NewWriter(context.Background())
	defer w.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return fmt.Errorf("Upload item %v failed: %v", msg.ItemID, err)
	}

	// update downlist
	if _, ok := c.downlist[msg.UserID]; !ok {
		c.downlist[msg.UserID] = map[string]struct{}{}
	}
	c.downlist[msg.UserID][msg.ItemID] = struct{}{}

	return c.save()
}
