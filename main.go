package main

import (
	"fmt"
	"strings"
	"test/natsacl"
)

func main() {
	builder := natsacl.NewBuilder()
	// builder.Stream("mystream").All()
	// builder.KV("mykv").List().Read().Write("me")
	builder.Stream("stream").Create()
	fmt.Printf("nsc edit signing-key --sk reader --allow-sub '%s'", strings.Join(builder.Sub(), ","))
	fmt.Printf(" --allow-pub '%s'\n", strings.Join(builder.Pub(), ","))
}
