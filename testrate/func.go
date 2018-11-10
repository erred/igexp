package fun

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
)

type Data struct {
	Data     []byte
	External bool
}

var o sync.Once
var c *pubsub.Client

func F(ctx context.Context, data Data) error {
	o.Do(func() {
		var err error
		c, err = pubsub.NewClient(context.Background(), os.Getenv("GCP_PROJECT"))
		if err != nil {
			log.Println("error ps client: ", err)
		}
	})

	fmt.Println("received: ", string(data.Data))

	base := "bmsg"
	var results []*pubsub.PublishResult
	if data.External {
		for i := 0; i < 300; i++ {
			r := c.Topic("test-ratelimit").Publish(context.Background(), &pubsub.Message{Data: []byte(fmt.Sprintf(base+"-%v", i))})
			results = append(results, r)
		}
		for _, r := range results {
			r.Get(ctx)
		}

	} else {
		time.Sleep(10 * time.Second)
	}

	return nil
}
