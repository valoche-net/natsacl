package main

import (
	"fmt"
	"os"

	"github.com/valoche-net/natsacl/v2/config"
)

func main() {
	f, err := os.OpenInRoot(".", "permissions.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	perms, err := config.Parse(f)
	if err != nil {
		panic(err)
	}
	for k, v := range perms {
		fmt.Println("\n", k)
		fmt.Println("-- pub -- >", v.Publish)
		fmt.Println("-- sub -- <", v.Subscribe)
	}
}
