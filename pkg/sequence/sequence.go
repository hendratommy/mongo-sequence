package sequence

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	DefaultCollectionName = "sequences"
	DefaultSequenceName   = "defaultSeq"
	DefaultTimeout        = 3 * time.Second
)

var ErrNotIntValueType = errors.New("value type is not int")

type Sequence struct {
	db             *mongo.Database
	collectionName string
	timeout        time.Duration
}

// New create new Sequence with given database and collection name. An collection might contains multiple sequence,
// each sequence have unique name. Each sequence operation will use specified timeout, if timeout value is zero then
// default timeout will be used (default to 3 seconds).
func New(db *mongo.Database, collectionName string, timeout time.Duration) *Sequence {
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	return &Sequence{
		db:             db,
		collectionName: collectionName,
		timeout:        timeout,
	}
}

var defaultSeq *Sequence

// SetupDefaultSequence setup database and timeout to be used by default sequence
func SetupDefaultSequence(db *mongo.Database, timeout time.Duration) {
	defaultSeq = New(db, DefaultCollectionName, timeout)
}

func (s *Sequence) collection() *mongo.Collection {
	return s.db.Collection(s.collectionName)
}

func (s *Sequence) ctx() context.Context {
	c, _ := context.WithTimeout(context.Background(), s.timeout)
	return c
}

// NextVal return the value of the sequence with given name then increment it's value,
// name must not be empty use DefaultSequenceName to use default sequence.
func (s *Sequence) NextVal(name string) (int, error) {
	opts := options.FindOneAndUpdate().SetUpsert(true)
	filter := bson.D{{"name", name}}
	inc := bson.M{"$inc": bson.M{"value": 1}}
	res := s.collection().FindOneAndUpdate(s.ctx(), filter, inc, opts)
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			// retry in case the document just been created by last operation
			res = s.collection().FindOneAndUpdate(s.ctx(), filter, inc, opts)
			err = res.Err()
		}
		if err != nil {
			return 0, res.Err()
		}
	}
	var doc bson.M
	if err := res.Decode(&doc); err != nil {
		return 0, err
	}

	switch v := doc["value"].(type) {
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	default:
		return 0, ErrNotIntValueType
	}
}

// NextVal will use default sequence to return the value of the sequence with given name then increment it's value,
// name must not be empty use DefaultSequenceName to use default sequence.
func NextVal(name string) (int, error) {
	return defaultSeq.NextVal(name)
}
