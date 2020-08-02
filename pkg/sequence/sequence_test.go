package sequence_test

import (
	"context"
	"github.com/hendratommy/mongo-sequence/pkg/sequence"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

func createClient() *mongo.Client {
	var err error
	opts := options.Client().ApplyURI(os.Getenv("MONGODB_URI"))
	client, err := mongo.NewClient(opts)
	if err != nil {
		log.Fatal(err)
	}
	if err = client.Connect(context.TODO()); err != nil {
		log.Fatal(err)
	}
	if err = client.Ping(context.TODO(), nil); err != nil {
		log.Fatal(err)
	}
	return client
}

func dropDB(db *mongo.Database) {
	_ = db.Drop(context.TODO())
}

func timeout() time.Duration {
	return 0 * time.Millisecond
}

func try(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func TestDefaultSequence(t *testing.T) {
	client := createClient()
	db := client.Database("sequence_test")
	// clean test db after test run
	defer func() {
		dropDB(db)
		_ = client.Disconnect(context.TODO())
	}()
	coll := db.Collection("sequences")
	sequence.SetupDefaultSequence(coll, timeout())

	val, err := sequence.NextVal(sequence.DefaultSequenceName)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, val)

		val, err = sequence.NextVal(sequence.DefaultSequenceName)
		if assert.NoError(t, err) {
			assert.Equal(t, 2, val)

			val, err = sequence.NextVal(sequence.DefaultSequenceName)
			if assert.NoError(t, err) {
				assert.Equal(t, 3, val)

				cursor, err := coll.Find(context.TODO(), bson.D{{Key: "_id", Value: sequence.DefaultSequenceName}})
				try(err)
				var results []bson.M
				err = cursor.All(context.TODO(), &results)
				try(err)

				assert.Len(t, results, 1)
				assert.EqualValues(t, 4, results[0]["value"])
			}
		}
	}
}

func TestNewSequence(t *testing.T) {
	client := createClient()
	db := client.Database("mySequenceDB_Test")
	// clean test db after test run
	defer func() {
		dropDB(db)
		_ = client.Disconnect(context.TODO())
	}()
	coll := db.Collection("my_sequences")

	seq := sequence.New(coll, timeout())
	seqName := "mySeq"
	val, err := seq.NextVal(seqName)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, val)

		val, err = seq.NextVal(seqName)
		if assert.NoError(t, err) {
			assert.Equal(t, 2, val)

			val, err = seq.NextVal(seqName)
			if assert.NoError(t, err) {
				assert.Equal(t, 3, val)

				cursor, err := coll.Find(context.TODO(), bson.D{{Key: "_id", Value: seqName}})
				try(err)
				var results []bson.M
				err = cursor.All(context.TODO(), &results)
				try(err)

				assert.Len(t, results, 1)
				assert.EqualValues(t, 4, results[0]["value"])
			}
		}
	}
}

func TestConcurrentDefaultSequence(t *testing.T) {
	n := 1024
	var wg sync.WaitGroup
	client := createClient()
	db := client.Database("con_sequence_test")
	// clean test db after test run
	defer func() {
		dropDB(db)
		_ = client.Disconnect(context.TODO())
	}()
	coll := db.Collection("sequences")
	sequence.SetupDefaultSequence(coll, timeout())

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := sequence.NextVal(sequence.DefaultSequenceName)
			if err != nil {
				assert.NoError(t, err)
			}
		}()
	}
	wg.Wait()

	cursor, err := coll.Find(context.TODO(), bson.D{{Key: "_id", Value: sequence.DefaultSequenceName}})
	try(err)
	var results []bson.M
	err = cursor.All(context.TODO(), &results)
	try(err)

	assert.Len(t, results, 1)
	assert.EqualValues(t, n+1, results[0]["value"])
}

func TestNewSequenceWithExistingColl(t *testing.T) {
	client := createClient()
	db := client.Database("mySequenceDB_Test")
	// clean test db after test run
	defer func() {
		dropDB(db)
		_ = client.Disconnect(context.TODO())
	}()
	coll := db.Collection("my_ex_sequences")
	seqName := "mySeq"
	wrongSeq := "wrongSeq"

	if _, err := coll.InsertOne(context.TODO(), bson.M{"_id": seqName, "value": 100}); err != nil {
		log.Fatal(err)
	}
	if _, err := coll.InsertOne(context.TODO(), bson.M{"_id": wrongSeq, "value": "100"}); err != nil {
		log.Fatal(err)
	}

	seq := sequence.New(coll, timeout())
	val, err := seq.NextVal(seqName)
	if assert.NoError(t, err) {
		assert.Equal(t, 100, val)

		val, err = seq.NextVal(seqName)
		if assert.NoError(t, err) {
			assert.Equal(t, 101, val)

			val, err = seq.NextVal(seqName)
			if assert.NoError(t, err) {
				assert.Equal(t, 102, val)
			}
		}
	}

	val, err = seq.NextVal(wrongSeq)
	if assert.Error(t, err) {
		assert.IsType(t, mongo.CommandError{}, err)
	}
	assert.Equal(t, 0, val)
}
