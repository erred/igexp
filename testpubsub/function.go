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
	Data       []byte
	TestString string
	TestBool   bool
	TestInt    int
}

func P(c context.Context, msg Msg) error {
	fmt.Println("Received: ", msg)
	buf := bytes.Buffer{}
	m2 := Msg{
		Data:       []byte("string to data byte"),
		TestString: "string test publish",
		TestBool:   false,
		TestInt:    411,
	}
	err := json.NewEncoder(&buf).Encode(m2)
	if err != nil {
		log.Println(err)
	}

	if msg.TestBool {
		cl, err := pubsub.NewClient(context.Background(), os.Getenv("GCP_PROJECT"))
		if err != nil {
			log.Println(err)
		}
		cl.Topic("testpubsub").Publish(context.Background(), &pubsub.Message{Data: buf.Bytes()})

	}
	return nil
}
