package main

import (
	"fmt"
	"log"
	"time"

	"github.com/seankhliao/igtools/goinsta"
)

func main() {
	ig := goinsta.New("", "")
	if err := ig.Login(); err != nil {
		log.Fatal("failed to login: ", err)
	}

	var u goinsta.User
	following := ig.Account.Following()
	for following.Next() {
		for _, f := range following.Users {
			u = f
			break
			// fmt.Println(f.Username)
		}
	}

	f := u.Feed()
	f.Next()

	fmt.Print(time.Unix(int64(f.Items[0].TakenAt), 0))

	// if err := ig.Export("./new_goinsta.json"); err != nil {
	// 	log.Fatal(err)
	// }
}
