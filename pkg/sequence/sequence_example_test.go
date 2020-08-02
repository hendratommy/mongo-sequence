package sequence_test

import (
	"context"
	"fmt"
	"github.com/hendratommy/mongo-sequence/pkg/sequence"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"time"
)

func ExampleNextVal() {
	// create mongodb client
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		panic(err)
	}
	if err = client.Connect(context.TODO()); err != nil {
		panic(err)
	}
	db := client.Database("myDB")

	// clean up test
	defer func() {
		_ = db.Drop(context.TODO())
		_ = client.Disconnect(context.TODO())
	}()

	// set `sequences` collection to use to store sequence records
	coll := db.Collection("sequences")
	// coll := db.Collection("mySequences")

	// set default sequence to use `sequences` collection
	sequence.SetupDefaultSequence(coll, 1*time.Second)

	// get next value from `DefaultSequenceName` sequence
	val, err := sequence.NextVal(sequence.DefaultSequenceName)
	if err != nil {
		panic(err)
	}
	fmt.Printf("value is: %d\n", val)
	val, err = sequence.NextVal(sequence.DefaultSequenceName)
	if err != nil {
		panic(err)
	}
	fmt.Printf("value is: %d\n", val)
	val, err = sequence.NextVal(sequence.DefaultSequenceName)
	if err != nil {
		panic(err)
	}
	fmt.Printf("value is: %d\n", val)

	// get next value from `orderSeq` sequence
	val, err = sequence.NextVal("orderSeq")
	if err != nil {
		panic(err)
	}
	fmt.Printf("value is: %d\n", val)

	// Output:
	// value is: 1
	// value is: 2
	// value is: 3
	// value is: 1
}

func ExampleNew() {
	// create mongodb client
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		panic(err)
	}
	if err = client.Connect(context.TODO()); err != nil {
		panic(err)
	}
	db := client.Database("myDB")

	// cleanup test
	defer func() {
		_ = db.Drop(context.TODO())
		_ = client.Disconnect(context.TODO())
	}()

	// set `myApp` collection to use to store sequence records
	myAppColl := db.Collection("myApp")
	// use `myApp` collection to store sequence
	myAppSeq := sequence.New(myAppColl, 1*time.Second)

	// set `myOtherApp` collection to use to store sequence records
	myOtherAppColl := db.Collection("myOtherApp")
	// use `myOtherApp` collection to store sequence
	myOtherAppSeq := sequence.New(myOtherAppColl, 1*time.Second)

	// get next value from `DefaultSequenceName` sequence
	val, err := myAppSeq.NextVal(sequence.DefaultSequenceName)
	if err != nil {
		panic(err)
	}
	fmt.Printf("value is: %d\n", val)
	val, err = myAppSeq.NextVal(sequence.DefaultSequenceName)
	if err != nil {
		panic(err)
	}
	fmt.Printf("value is: %d\n", val)
	// get next value from `orderSeq` sequence
	val, err = myAppSeq.NextVal("orderSeq")
	if err != nil {
		panic(err)
	}
	fmt.Printf("value is: %d\n", val)

	// get next value from `DefaultSequenceName` sequence
	val, err = myOtherAppSeq.NextVal(sequence.DefaultSequenceName)
	if err != nil {
		panic(err)
	}
	fmt.Printf("value is: %d\n", val)
	val, err = myOtherAppSeq.NextVal(sequence.DefaultSequenceName)
	if err != nil {
		panic(err)
	}
	fmt.Printf("value is: %d\n", val)
	// get next value from `orderSeq` sequence
	val, err = myOtherAppSeq.NextVal("orderSeq")
	if err != nil {
		panic(err)
	}
	fmt.Printf("value is: %d\n", val)

	// Output:
	// value is: 1
	// value is: 2
	// value is: 1
	// value is: 1
	// value is: 2
	// value is: 1
}
