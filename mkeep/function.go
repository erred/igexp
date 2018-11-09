package mkeep

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/seankhliao/igtools/goinsta"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
)

var (
	c = client{}

	// names of environment variables
	envBucket  = "BUCKET"
	envTopic   = "TOPIC"
	envProject = "GCP_PROJECT"

	// names of storage objects
	objBase      = "mkeep"
	objBlacklist = "mkeep/blacklist.json"
	objDownlist  = "mkeep/downlist.json"
	objGoinsta   = "mkeep/goinsta.json"
)

type client struct {
	bucket   *storage.BucketHandle
	ig       *goinsta.Instagram
	once     sync.Once
	topic    *pubsub.Topic
	downlist Downlist
}

func (cl *client) Login() {
	var err error
	ctx := context.Background()

	// storage
	store, err := storage.NewClient(ctx)
	if err != nil {
		panic(fmt.Errorf("Login Error: cloud storage failed: %v", err))
	}
	cl.bucket = store.Bucket(os.Getenv(envBucket))

	// pubsub
	psc, err := pubsub.NewClient(ctx, os.Getenv(envProject))
	if err != nil {
		panic(fmt.Errorf("Login Error: pubsub client failed:%v", err))
	}
	cl.topic = psc.Topic(os.Getenv(envTopic))

	// goinsta
	r, err := cl.bucket.Object(objGoinsta).NewReader(ctx)
	if err != nil {
		panic(fmt.Errorf("Login Error: import %v failed: %v", objGoinsta, err))
	}
	defer r.Close()

	cl.ig, err = goinsta.ImportReader(r)
	if err != nil {
		panic(fmt.Errorf("Import Error: %v", err))
	}

	// downlist
	r, err = cl.bucket.Object(objDownlist).NewReader(ctx)
	if err != nil {
		if err != storage.ErrObjectNotExist {
			panic(fmt.Errorf("Login Error: import %v failed: %v", objDownlist, err))
		}
	} else {
		defer r.Close()

		if err = json.NewDecoder(r).Decode(&cl.downlist); err != nil {
			panic(fmt.Errorf("Decode downlist: %v", err))
		}

	}
}

// save saves the current follows back to storage
func (cl *client) save() {
	ctx := context.Background()

	// Goinsta state
	w := cl.bucket.Object(objGoinsta).NewWriter(ctx)
	defer w.Close()
	if err := goinsta.Export(c.ig, w); err != nil {
		log.Println("Goinsta export failed: ", err)
	}

	// downlist
	w = cl.bucket.Object(objGoinsta).NewWriter(ctx)
	defer w.Close()
	if err := json.NewEncoder(w).Encode(cl.downlist); err != nil {
		log.Println("Downlist export failed: ", err)
	}
}

func Mkeep(ctx context.Context, di DownloadItem) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("panic: ", r)
		}
	}()

	// ensure login
	c.once.Do(c.Login)

	// trigger

	if di.ExternalTrigger {
		fmt.Println("External trigger", di)
		a, err := newArchive()
		if err != nil {
			log.Println("newArchive failed:", err)
			return err
		}

		a.getNewMedia()
	} else {
		fmt.Println("Internal trigger", di)

		// itname := path.Join(objBase, "media", di.UserID, di.ItemID+di.Ext)
		// w := c.bucket.Object(itname).NewWriter(context.Background())
		// defer w.Close()
		//
		// res, err := c.ig.Client().Get(di.Url)
		// if err != nil {
		// 	log.Println("failed to download: ", err)
		// }
		// defer res.Body.Close()
		//
		// io.Copy(w, res.Body)

		// uid, err := strconv.ParseInt(di.UserID, 10, 64)
		// if err != nil {
		// 	log.Println("failed to parse int: ", err)
		// }
		// c.downlist[uid][di.ItemID] = struct{}{}
		fmt.Println("finished download: ", di.ItemID)
	}

	// c.save()

	return nil
}

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

type Blacklist map[int64][]string

// Downlist is a list of media we already have
type Downlist map[int64]map[string]struct{}

type archive struct {
	blacklist Blacklist
	downlist  Downlist
}

func newArchive() (*archive, error) {
	ctx := context.Background()
	a := archive{}

	r, err := c.bucket.Object(objBlacklist).NewReader(ctx)
	switch err {
	case nil:
		defer r.Close()
		if err := json.NewDecoder(r).Decode(&a.blacklist); err != nil {
			return &a, fmt.Errorf("%v decode error: %v", objBlacklist, err)
		}
		fmt.Println("Successfully imported blacklist.json")
	case storage.ErrObjectNotExist:
	default:
		return &a, fmt.Errorf("%v reader error: %v", objBlacklist, err)
	}

	r, err = c.bucket.Object(objDownlist).NewReader(ctx)
	switch err {
	case nil:
		defer r.Close()
		if err := json.NewDecoder(r).Decode(&a.downlist); err != nil {
			return &a, fmt.Errorf("%v decode error: %v", objDownlist, err)
		}
		fmt.Println("Successfully imported downlist.json")
	case storage.ErrObjectNotExist:
	default:
		return &a, fmt.Errorf("%v reader error: %v", objDownlist, err)
	}

	return &a, nil
}

func (a archive) blacklisted(userID int64, feed string) bool {
	list, ok := a.blacklist[userID]
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

func (a archive) downloaded(userID int64, mediaID string) bool {
	u, ok := a.downlist[userID]
	if !ok {
		return false
	}
	_, ok = u[mediaID]
	if !ok {
		return false
	}
	return true
}

func (a *archive) getItems(its []goinsta.Item, u goinsta.User) {
	for _, it := range its {
		breakout := false
		if len(it.CarouselMedia) != 0 {
			for _, i := range it.CarouselMedia {
				if a.downloaded(u.ID, i.ID) {
					breakout = true
					break
				}
				if err := queue(it); err != nil {
					log.Println("queue failed: ", err)
				}
			}
			if breakout {
				break
			}
		}
		if a.downloaded(u.ID, it.ID) {
			break
		}
		if err := queue(it); err != nil {
			log.Println("queue failed: ", err)
		}
	}

}

func (a *archive) getUserMedia(u goinsta.User) {
	if !a.blacklisted(u.ID, blacklistStory) {
		stories := u.Stories()
		for stories.Next() {
			a.getItems(stories.Items, u)
		}
	}

	if !a.blacklisted(u.ID, blacklistFeed) {
		feed := u.Feed()
		for feed.Next() {
			a.getItems(feed.Items, u)
		}
	}

	if !a.blacklisted(u.ID, blacklistTag) {
		feed, err := u.Tags([]byte{})
		if err != nil {
			log.Printf("getNewMedia pre get tags for %v, err: %v", u.Username, err)
		}
		for feed.Next() {
			a.getItems(feed.Items, u)
		}

	}

}

func (a *archive) getNewMedia() {
	following := c.ig.Account.Following()
	counter := 0
	for following.Next() {
		for _, u := range following.Users {
			counter++
			fmt.Println("Processing user: #", counter, " ", u.Username)
			a.getUserMedia(u)

			if counter > 1 {
				return
			}
		}
	}
}

func queue(item goinsta.Item) error {
	buf := bytes.Buffer{}
	ctx := context.Background()
	if err := json.NewEncoder(&buf).Encode(NewDownloadItem(item)); err != nil {
		return fmt.Errorf("encode failed: %v", err)
	}
	msg := pubsub.Message{Data: buf.Bytes()}
	fmt.Println("q encoded: ", buf.String(), "pubsub msg: ", msg)
	if _, err := c.topic.Publish(ctx, &msg).Get(ctx); err != nil {
		return fmt.Errorf("queue failed: %v", err)
	}
	return nil
}

type DownloadItem struct {
	ExternalTrigger bool
	UserID          string
	ItemID          string
	Ext             string
	Url             string
}

func NewDownloadItem(it goinsta.Item) DownloadItem {
	var link string
	if len(it.Images.Versions) > 0 {
		link = it.Images.GetBest()
	} else {
		link = goinsta.GetBest(it.Videos)
	}
	u, err := url.Parse(link)
	if err != nil {
		log.Println("failed to parse url: ", err)
		return DownloadItem{}
	}
	return DownloadItem{
		UserID: strconv.FormatInt(it.User.ID, 10),
		ItemID: it.ID,
		Ext:    path.Ext(u.Path),
		Url:    link,
	}
}
