package main

import (
	"fmt"
	"math/rand"

	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/encoding/protobuf"
	"github.com/philippgille/gokv/examples/protobuf_encoding/tutorialpb"
	"github.com/philippgille/gokv/gomap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	// Create store with customized options (using the protobuf codec)
	options := gomap.Options{
		Codec: protobuf.Codec,
	}
	store := gomap.NewStore(options)
	defer func() { _ = store.Close() }()

	// Store, retrieve and print a value
	interactWithStore(store)
}

// interactWithStore stores, retrieves and prints a value.
// It's completely independent of the store implementation.
func interactWithStore(store gokv.Store) {
	// Store value
	val := &tutorialpb.Person{
		Name:  "John Doe",
		Id:    rand.Int31(),
		Email: "johndoe@example.com",
		Phones: []*tutorialpb.Person_PhoneNumber{
			{Number: "0123-456789", Type: tutorialpb.Person_HOME},
			{Number: "0987-654321", Type: tutorialpb.Person_WORK},
		},
		LastUpdated: timestamppb.Now(),
	}
	err := store.Set("foo123", val)
	if err != nil {
		panic(err)
	}

	// Retrieve value
	retrievedVal := new(tutorialpb.Person)
	found, err := store.Get("foo123", retrievedVal)
	if err != nil {
		panic(err)
	}
	if !found {
		panic("Value not found")
	}

	fmt.Printf("Person: %+v\n", retrievedVal) // Prints `Person: name:"John Doe"  id:1987919731  email:"johndoe@example.com"  phones:{number:"0123-456789"  type:HOME}  phones:{number:"0987-654321"  type:WORK}  last_updated:{seconds:1699139227  nanos:396703774}`
}
