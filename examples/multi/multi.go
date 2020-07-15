package main

import (
	"context"
	"fmt"
	"github.com/hendratommy/mongo-sequence/pkg/sequence"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func main() {
	// create mongodb client
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://root:12345@localhost:27017"))
	if err != nil {
		panic(err)
	}
	if err = client.Connect(context.TODO()); err != nil {
		panic(err)
	}

	// instead using default sequence, create new sequence
	seq := sequence.New(client.Database("myDB"), "mySequenceCollection", 30*time.Second)
	val, err := seq.NextVal("mySeq")
	if err != nil {
		panic(err)
	}
	fmt.Printf("value is: %d\n", val) // value is: 1
	val, err = seq.NextVal("orderSeq")
	if err != nil {
		panic(err)
	}
	fmt.Printf("value is: %d\n", val) // value is: 1
	val, err = seq.NextVal("mySeq")
	if err != nil {
		panic(err)
	}
	fmt.Printf("value is: %d\n", val) // value is: 2
}
