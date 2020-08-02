# mongo-sequence

Sequence implementation for mongodb. Unlike `sql` databases, `mongodb` doesn't have any built-in functionality to create
sequence.

`mongo-sequence` provide `sequence` document on `mongodb`.

## Installation

```bash
go get github.com/hendratommy/mongo-sequence
```

## Examples
```go
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
```

See [examples](https://github.com/hendratommy/mongo-sequence/tree/master/pkg/sequence/sequence_example_test.go) for more details.
