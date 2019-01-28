package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/seankhliao/igtools/goinsta"
)

type Blacklist map[int64][]string

type Downlist map[int64]map[string]struct{}

func main() {
	var users map[int64]goinsta.User

	f, err := os.Open("./following.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&users)
	if err != nil {
		log.Fatal(err)
	}

	d := Downlist{}

	for id, user := range users {
		fmt.Println(id, user.Username, user.FullName)
		d[id] = map[string]struct{}{}
	}
	f2, err := os.Create("./downlist.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f2.Close()
	err = json.NewEncoder(f2).Encode(d)
	if err != nil {
		log.Fatal(err)
	}
}
