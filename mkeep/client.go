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

	"cloud.google.com/go/datastore"
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
	once sync.Once

	bucket *storage.BucketHandle
	dstore *datastore.Client

	akey *datastore.Key
	// usercol  *firestore.CollectionRef
	// mediacol *firestore.CollectionRef
	topic *pubsub.Topic
	ig    *goinsta.Instagram

	// downlist  Downlist
	// blacklist Blacklist
}

// UserDoc stores blacklist info
type UserDoc struct {
	Feed  bool
	Story bool
	Tag   bool
}

// MediaDoc stores media info
type MediaDoc struct {
	User int64
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

	// firestore
	// fire, err := firestore.NewClient(ctx, os.Getenv(envProject))
	// if err != nil {
	// 	panic(fmt.Errorf("Login Error: firestore failed: %v", err))
	// }
	// c.usercol = fire.Collection(os.Getenv(envFireUser))
	// c.mediacol = fire.Collection(os.Getenv(envFireMedia))

	// datastore
	dstore, err := datastore.NewClient(ctx, os.Getenv(envProject))
	if err != nil {
		panic(fmt.Errorf("Login Error: datastore failed: %v", err))
	}
	c.akey = datastore.NameKey("igtools", "mkeep", nil)

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
	fmt.Println("successfully completed setup")
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
	// w = c.bucket.Object(objGoinsta).NewWriter(ctx)
	// defer w.Close()
	// if err := json.NewEncoder(w).Encode(c.downlist); err != nil {
	// 	return fmt.Errorf("downlist export failed: %v", err)
	// }
	return nil
}

func (c *Client) getMediaExist(id string) bool {
	// _, err := c.mediacol.Doc(id).Get(context.Background())
	// if err != nil {
	// 	if grpc.Code(err) == codes.NotFound {
	// 		return false
	// 	}
	// 	log.Println("Unknown error getting media doc status for ", id, ": ", err)
	// 	return false
	// }
	key := datastore.NameKey("media", id, c.akey)
	var empty MediaDoc
	if err := c.dstore.Get(context.Background(), key, &empty); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return false
		}
		log.Println("Unknown error getting media doc status for ", id, ": ", err)
		return false
	}
	return true
}

func (c *Client) getUserDoc(id int64) (UserDoc, error) {
	uid := strconv.FormatInt(id, 10)
	udoc := UserDoc{true, true, true}
	key := datastore.NameKey("user", uid, c.akey)
	// dss, err := c.usercol.Doc(uid).Get(context.Background())
	// if err != nil {
	// 	if grpc.Code(err) == codes.NotFound {
	// 		_, err = c.usercol.Doc(uid).Create(context.Background(), udoc)
	// 		if err != nil {
	// 			return udoc, fmt.Errorf("Error creating userDoc for %v", id)
	// 		}
	// 		return udoc, nil
	// 	}
	// 	return udoc, fmt.Errorf("Unkown error getUserDoc: %v", err)
	// }
	// if err := dss.DataTo(&udoc); err != nil {
	// 	return udoc, fmt.Errorf("Unkown error getUserDoc: %v", err)
	// }
	if err := c.dstore.Get(context.Background(), key, &udoc); err != nil {
		if err == datastore.ErrNoSuchEntity {
			_, err = c.dstore.Put(context.Background(), key, udoc)
			if err != nil {
				return udoc, fmt.Errorf("Error creating userDoc for %v", id)
			}
			return udoc, nil
		}
		return udoc, fmt.Errorf("Unkown error getUserDoc: %v", err)
	}
	return udoc, nil
}

func (c *Client) getUsers() {
	following := c.ig.Account.Following()
	counter := 0
	for following.Next() {
		for _, user := range following.Users {
			udoc, err := c.getUserDoc(user.ID)
			if err != nil {
				log.Println("Error getting user doc for ", user.ID, " ", user.Username)
				continue
			}
			if udoc.Feed || udoc.Story || udoc.Tag {
				counter++
				c.queueUser(user)
			}
		}
	}
	fmt.Println("getUsers queued ", counter, " users")
}

func (c *Client) getFeeds(msg Message) {
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

	udoc, err := c.getUserDoc(user.ID)
	if err != nil {
		log.Println("Error getting user doc for ", user.ID, " ", user.Username)
		return
	}

	if udoc.Story {
		feed := user.Stories()
		for feed.Next() {
			c.getItems(feed.Items)
		}
	}
	if udoc.Feed {
		feed := user.Feed()
		for feed.Next() {
			c.getItems(feed.Items)
		}

	}
	if udoc.Tag {
		feed, err := user.Tags([]byte{})
		if err != nil {
			log.Printf("get tagged for %v, %v error: %v", user.ID, user.Username, err)
			return
		}
		for feed.Next() {
			c.getItems(feed.Items)
		}
	}
	return
}

func (c *Client) getItems(items []goinsta.Item) {
	for _, item := range items {
		if len(item.CarouselMedia) != 0 {
			breakout := false
			for _, it := range item.CarouselMedia {
				if c.getMediaExist(it.ID) {
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

		if c.getMediaExist(item.ID) {
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

func (c *Client) download(msg Message) {
	resp, err := c.ig.Client().Get(msg.URL)
	if err != nil {
		log.Printf("Download item %v failed: %vn", msg.ItemID, err)
		return
	}
	defer resp.Body.Close()

	obj := path.Join(objBase, "media", strconv.FormatInt(msg.UserID, 10), msg.ItemID+msg.Ext)
	w := c.bucket.Object(obj).NewWriter(context.Background())
	defer w.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Upload item %v failed: %v\n", msg.ItemID, err)
		return
	}

	// update downlist
	mdoc := MediaDoc{msg.UserID}
	key := datastore.NameKey("media", msg.ItemID, c.akey)
	_, err = c.dstore.Put(context.Background(), key, mdoc)
	// _, err = c.mediacol.Doc(msg.ItemID).Create(context.Background(), mdoc)
	if err != nil {
		log.Println("Error saving mediadoc: ", err)
	}

	err = c.save()
	if err != nil {
		log.Println("Error saving: ", err)
	}
}
