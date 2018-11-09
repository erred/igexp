package fun

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/pubsub"
)

type Msg struct {
	Data []byte
	C    bool
}

type MyData struct {
	Hello string
	World bool
}

func P(c context.Context, msg Msg) error {
	if msg.C {
		buf := bytes.Buffer{}

		mydata := MyData{
			Hello: "push this string",
			World: true,
		}

		if err := json.NewEncoder(&buf).Encode(mydata); err != nil {
			log.Println(err)
		}
		cl, err := pubsub.NewClient(context.Background(), os.Getenv("GCP_PROJECT"))
		if err != nil {
			log.Println(err)
		}
		_, err = cl.Topic("igtools-testpubsub").Publish(context.Background(), &pubsub.Message{Data: buf.Bytes()}).Get(context.Background())
		if err != nil {
			log.Println(err)
		}

	}

	var decoded MyData
	if err := json.Unmarshal(msg.Data, &decoded); err != nil {
		log.Println(err)
	}
	fmt.Println("json decoded: ", decoded)

	return nil
}
