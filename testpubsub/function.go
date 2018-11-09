package fun

import (
	"bytes"
	"context"
	"encoding/base64"
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
		_, err = cl.Topic("igtools-testpubsub").Publish(context.Background(), &pubsub.Message{Data: buf.Bytes()}).Get(context.Background())
		if err != nil {
			log.Println(err)
		}

	}

	dst := make([]byte, 100)
	_, err = base64.StdEncoding.Decode(dst, msg.Data)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(dst))
	return nil
}
