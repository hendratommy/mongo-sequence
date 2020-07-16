package sequence

import (
	"context"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"sync"
	"testing"
	"time"
)

func createClient() *mongo.Client {
	var err error
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		panic(err)
	}
	if err = client.Connect(ctx()); err != nil {
		panic(err)
	}
	return client
}

func timeout() time.Duration {
	return 3 * time.Second
}

func ctx() context.Context {
	c, _ := context.WithTimeout(context.Background(), timeout())
	return c
}

func dropDB(client *mongo.Client, db string) {
	client.Database(db).Drop(ctx())
}

func TestDefaultSequence(t *testing.T) {
	dbName := "sequence_test"
	client := createClient()
	// clean test db after test run
	defer dropDB(client, dbName)
	// this to ensure collection exists, because create collections cannot be done in multi document transactions
	client.Database(dbName).CreateCollection(ctx(), DefaultCollectionName)
	SetupDefaultSequence(client.Database(dbName), timeout())

	Convey("Test default sequence functionalities", t, func() {

		Convey("NextVal should be equals 1", func() {
			val, err := NextVal(DefaultSequenceName)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 1)
		})
		Convey("NextVal should be equals 2", func() {
			val, err := NextVal(DefaultSequenceName)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 2)
		})
		Convey("NextVal should be equals 3", func() {
			val, err := NextVal(DefaultSequenceName)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 3)
		})

		Convey("Should be able to query using manual find", func() {
			coll := client.Database(dbName).Collection(DefaultCollectionName)
			cursor, err := coll.Find(ctx(), bson.D{{"name", DefaultSequenceName}})
			So(err, ShouldBeNil)
			var results []bson.M
			err = cursor.All(ctx(), &results)
			So(err, ShouldBeNil)

			Convey("Length should be 1 and value should be 4", func() {
				So(len(results), ShouldEqual, 1)
				So(results[0]["value"], ShouldEqual, 4)
			})
		})
	})
}

func TestNewSequence(t *testing.T) {
	client := createClient()
	dbName := "mySequenceDB_Test"
	collName := "my_sequences"
	seqName := "mySeq"
	// clean test db after test run
	defer dropDB(client, dbName)
	// this to ensure collection exists, because create collections cannot be done in multi document transactions
	client.Database(dbName).CreateCollection(ctx(), collName)

	Convey("Test new sequence functionalities", t, func() {
		seq := New(client.Database(dbName), collName, timeout())
		Convey("NextVal should be equals 1", func() {
			val, err := seq.NextVal(seqName)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 1)
		})
		Convey("NextVal should be equals 2", func() {
			val, err := seq.NextVal(seqName)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 2)
		})
		Convey("NextVal should be equals 3", func() {
			val, err := seq.NextVal(seqName)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 3)
		})

		Convey("Should be able to queried by manual find", func() {
			coll := client.Database(dbName).Collection(collName)
			cursor, err := coll.Find(ctx(), bson.D{{"name", seqName}})
			So(err, ShouldBeNil)
			var results []bson.M
			err = cursor.All(ctx(), &results)
			So(err, ShouldBeNil)

			Convey("Length should be 1 and value should be 4", func() {
				So(len(results), ShouldEqual, 1)
				So(results[0]["value"], ShouldEqual, 4)
			})
		})
	})
}

func TestConcurrentDefaultSequence(t *testing.T) {
	n := 1024
	var wg sync.WaitGroup
	var (
		mux    sync.Mutex
		valMap = make(map[int]int)
	)
	dbName := "con_sequence_test"
	client := createClient()
	// clean test db after test run
	defer dropDB(client, dbName)
	// this to ensure collection exists, because create collections cannot be done in multi document transactions
	client.Database(dbName).CreateCollection(ctx(), DefaultCollectionName)
	SetupDefaultSequence(client.Database(dbName), timeout())

	Convey("Test concurrent sequence using defaultSequence", t, func() {
		Convey(fmt.Sprintf("Run n=%d concurrent sequence request", n), func() {
			for i := 0; i < n; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					val, err := NextVal(DefaultSequenceName)
					if err != nil {
						t.Errorf("NextVal returned an error: %v", err)
					}
					mux.Lock()
					valMap[val] = valMap[val]+1
					mux.Unlock()
				}()
			}
			wg.Wait()
		})

		Convey(fmt.Sprintf("Value map's length should be = %d and should not have duplicate value", n), func() {
			So(len(valMap), ShouldEqual, n)
			for _, v := range valMap {
				So(v, ShouldEqual, 1)
			}
		})

		Convey(fmt.Sprintf("The end value should be %d", n+1), func() {
			coll := client.Database(dbName).Collection(DefaultCollectionName)
			cursor, err := coll.Find(ctx(), bson.D{{"name", DefaultSequenceName}})
			So(err, ShouldBeNil)
			var results []bson.M
			err = cursor.All(ctx(), &results)
			So(err, ShouldBeNil)
			So(len(results), ShouldEqual, 1)
			So(results[0]["value"], ShouldEqual, n+1)
		})
	})
}

func TestNewSequenceWithExistingColl(t *testing.T) {
	client := createClient()
	dbName := "mySequenceDB_Test"
	collName := "my_ex_sequences"
	seqName := "mySeq"
	wrongSeq := "wrongSeq"
	// clean test db after test run
	defer dropDB(client, dbName)
	// this to ensure collection exists, because create collections cannot be done in multi document transactions
	client.Database(dbName).CreateCollection(ctx(), collName)

	if _, err := client.Database(dbName).Collection(collName).InsertOne(ctx(), bson.M{ "name": seqName, "value": 100 }); err != nil {
		panic(err)
	}
	if _, err := client.Database(dbName).Collection(collName).InsertOne(ctx(), bson.M{ "name": wrongSeq, "value": "100" }); err != nil {
		panic(err)
	}

	Convey("Test sequence with existing sequence", t, func() {
		Convey("Test existing sequence with correct value", func() {
			seq := New(client.Database(dbName), collName, timeout())
			Convey("NextVal should be equals 100", func() {
				val, err := seq.NextVal(seqName)
				So(err, ShouldBeNil)
				So(val, ShouldEqual, 100)
			})
			Convey("NextVal should be equals 101", func() {
				val, err := seq.NextVal(seqName)
				So(err, ShouldBeNil)
				So(val, ShouldEqual, 101)
			})
			Convey("NextVal should be equals 102", func() {
				val, err := seq.NextVal(seqName)
				So(err, ShouldBeNil)
				So(val, ShouldEqual, 102)
			})
		})

		Convey("Test existing sequence with wrong value", func() {
			seq := New(client.Database(dbName), collName, timeout())
			Convey("String value should return mongo.CommandError", func() {
				val, err := seq.NextVal(wrongSeq)
				So(val, ShouldEqual, 0)
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveSameTypeAs, mongo.CommandError{})
			})
		})
	})
}
