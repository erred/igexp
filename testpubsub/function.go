package fun

import (
	"context"
	"fmt"
)

type Msg struct {
	Data       []byte
	TestString string
	TestBool   bool
	TestInt    int
}

func P(c context.Context, msg Msg) error {
	fmt.Println("Received: ", msg)
	return nil
}
