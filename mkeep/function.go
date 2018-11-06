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

	fmt.Println("Logged in with restore")
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

	fmt.Println("saved!")
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
		fmt.Println("External trigger")
		a, err := newArchive()
		if err != nil {
			log.Println("newArchive failed:", err)
			return err
		}

		if err := a.getNewMedia(); err != nil {
			log.Println("getNewMedia failed: ", err)
			return err
		}
	} else {
		fmt.Println("Internal trigger")

		itname := path.Join(objBase, "media", di.UserID, di.ItemID+di.Ext)
		w := c.bucket.Object(itname).NewWriter(context.Background())
		defer w.Close()

		res, err := c.ig.Client().Get(di.Url)
		if err != nil {
			log.Println("failed to download: ", err)
		}
		defer res.Body.Close()

		io.Copy(w, res.Body)

		uid, err := strconv.ParseInt(di.UserID, 10, 64)
		if err != nil {
			log.Println("failed to parse int: ", err)
		}
		c.downlist[uid][di.ItemID] = struct{}{}
		fmt.Println("finished download: ", di.ItemID)
	}

	c.save()

	return nil
}

// Blacklist of media feeds to keep
// {
//	userID : { "tagged": {} }
// }
// story, feed, tag
var (
	blacklistStory = "story"
	blacklistFeed  = "feed"
	blacklistTag   = "tag"
)

type Blacklist map[int64]map[string]struct{}

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
	case storage.ErrObjectNotExist:
	default:
		return &a, fmt.Errorf("%v reader error: %v", objBlacklist, err)
	}

	r, err = c.bucket.Object(objDownlist).NewReader(ctx)
	switch err {
	case nil:
		defer r.Close()
		if err := json.NewDecoder(r).Decode(&a.blacklist); err != nil {
			return &a, fmt.Errorf("%v decode error: %v", objDownlist, err)
		}
	case storage.ErrObjectNotExist:
	default:
		return &a, fmt.Errorf("%v reader error: %v", objDownlist, err)
	}

	return &a, nil
}

func (a archive) blacklisted(userID int64, feed string) bool {
	fs, ok := a.blacklist[userID]
	if !ok {
		return false
	}
	_, ok = fs[feed]
	if !ok {
		return false
	}
	return true
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

func (a *archive) getNewMedia() error {
	following := c.ig.Account.Following()
	counter := 0
	for following.Next() {
		// if following.Error() != nil {
		// 	return fmt.Errorf("getNewMedia get following: %v", following.Error())
		// }
		for _, u := range following.Users {
			counter += 1
			fmt.Println("Processing user: #", counter, " ", u.Username)
			if !a.blacklisted(u.ID, blacklistStory) {
				stories := u.Stories()
				for stories.Next() {
					// if stories.Error() != nil {
					// 	log.Printf("getNewMedia get stories for %v, err: %v", u.Username, stories.Error())
					// }

					for _, it := range stories.Items {
						if a.downloaded(u.ID, it.ID) {
							break
						}

						if err := queue(it); err != nil {
							log.Println("queue failed: ", err)
						}
					}
				}
			}

			if !a.blacklisted(u.ID, blacklistFeed) {
				feed := u.Feed()
				for feed.Next() {
					// if feed.Error() != nil {
					// 	log.Printf("getNewMedia get feed for %v, err: %v", u.Username, feed.Error())
					// }

					for _, it := range feed.Items {
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
			}

			if !a.blacklisted(u.ID, blacklistTag) {
				feed, err := u.Tags([]byte{})
				if err != nil {
					log.Printf("getNewMedia pre get tags for %v, err: %v", u.Username, err)
				}
				for feed.Next() {
					// if feed.Error() != nil {
					// 	log.Printf("getNewMedia get tags for %v, err: %v", u.Username, feed.Error())
					// }

					for _, it := range feed.Items {
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

			}
		}
	}
	return nil
}

func queue(item goinsta.Item) error {
	buf := bytes.Buffer{}
	ctx := context.Background()
	if err := json.NewEncoder(&buf).Encode(NewDownloadItem(item)); err != nil {
		return fmt.Errorf("encode failed: %v", err)
	}
	if _, err := c.topic.Publish(ctx, &pubsub.Message{Data: buf.Bytes()}).Get(ctx); err != nil {
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
	var url string
	if len(it.Images.Versions) > 0 {
		url = it.Images.GetBest()
	} else {
		url = goinsta.GetBest(it.Videos)
	}
	return DownloadItem{
		UserID: strconv.FormatInt(it.User.ID, 10),
		ItemID: it.ID,
		Ext:    path.Ext(url),
		Url:    url,
	}
}
