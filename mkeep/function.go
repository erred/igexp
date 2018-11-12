package mkeep

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"path"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/seankhliao/igtools/goinsta"
)

var (
	client = Client{}

	// names of environment variables
	envBucket = "BUCKET"
	// envFireUser  = "FIRE_USER"
	// envFireMedia = "FIRE_MEDIA"
	envTopic   = "TOPIC"
	envProject = "GCP_PROJECT"

	// names of storage objects
	objBase      = "mkeep"
	objBlacklist = "mkeep/blacklist.json"
	objDownlist  = "mkeep/downlist.json"
	objGoinsta   = "mkeep/goinsta.json"
)

// Mkeep function entry point, decides where to pass the message
func Mkeep(ctx context.Context, psmsg pubsub.Message) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("panic: ", r)
		}
	}()

	// ensure login
	client.once.Do(client.setup)

	msg, err := parseMessage(psmsg)
	if err != nil {
		return fmt.Errorf("Parse pubsub message: %v", err)
	}

	switch msg.Mode {
	case ModeAll:
		fmt.Println("Getting all users")
		return client.getUsers()
	case ModeUser:
		fmt.Println("Getting all feeds for ", msg.Username)
		return client.getFeeds(msg)
	case ModeItem:
		fmt.Println("Getting item ", msg.ItemID, msg.Ext, " for ", msg.Username)
		return client.download(msg)
	default:
		return fmt.Errorf("Unknown msg mode: %v", msg.Mode)
	}
}

var (
	// ModeAll triggers everything
	ModeAll = 0
	// ModeUser triggers for a user
	ModeUser = 1
	// ModeItem triggers for an item
	ModeItem = 2
)

// Message comm through pubsub
type Message struct {
	Mode     int
	UserID   int64
	Username string
	ItemID   string
	Ext      string
	URL      string
	Time     time.Time
}

func parseMessage(psmsg pubsub.Message) (Message, error) {
	var msg Message
	if err := json.Unmarshal(psmsg.Data, &msg); err != nil {
		return msg, fmt.Errorf("json unmarshal: %v", err)
	}
	return msg, nil
}

func newUserMessage(user goinsta.User) Message {
	return Message{
		Mode:     ModeUser,
		UserID:   user.ID,
		Username: user.Username,
	}

}

func newItemMessage(item goinsta.Item, uid int64, uname string) Message {
	var link string
	if len(item.Images.Versions) > 0 {
		link = item.Images.GetBest()
		link = goinsta.GetBest(item.Images.Versions)
	} else {
		link = goinsta.GetBest(item.Videos)
	}
	if link == "" {
		log.Println("empty link: ", item.Videos, item.Images)
	}
	u, err := url.Parse(link)
	if err != nil {
		log.Println("failed to parse url: ", err)
		return Message{}
	}
	return Message{
		Mode:     ModeItem,
		UserID:   uid,
		Username: uname,
		ItemID:   item.ID,
		Ext:      path.Ext(u.Path),
		URL:      link,
		Time:     time.Unix(int64(item.TakenAt), 0),
	}

}
