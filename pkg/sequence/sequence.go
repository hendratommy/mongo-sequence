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
	DefaultSequenceName = "defaultSeq"
	DefaultTimeout      = 1000 * time.Millisecond
)

var ErrNotIntValueType = errors.New("value type is not int")

type Sequence struct {
	coll    *mongo.Collection
	timeout time.Duration
}

// New create new Sequence with given collection. An collection might contains multiple sequence,
// each sequence have their name as their id (_id). Each sequence operation will use specified timeout, if timeout value
// is zero then default timeout will be used (default to 1 seconds).
func New(coll *mongo.Collection, timeout time.Duration) *Sequence {
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	return &Sequence{
		coll:    coll,
		timeout: timeout,
	}
}

var defaultSeq *Sequence

// SetupDefaultSequence setup database and timeout to be used by default sequence
func SetupDefaultSequence(coll *mongo.Collection, timeout time.Duration) {
	defaultSeq = New(coll, timeout)
}

func (s *Sequence) ctx() context.Context {
	//c, _ := context.WithTimeout(context.Background(), s.timeout)
	//return c

	return context.Background()
}

// NextVal return the value of the sequence with given name then increment it's value, meaning value that are saved
// in mongodb is the "next" value.
// Name must not be empty use DefaultSequenceName to use default sequence.
func (s *Sequence) NextVal(name string) (int, error) {
	opts := options.FindOneAndUpdate().SetUpsert(true)
	filter := bson.D{{Key: "_id", Value: name}}
	inc := bson.M{"$inc": bson.M{"value": 1}}
	res := s.coll.FindOneAndUpdate(s.ctx(), filter, inc, opts)
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			// retry in case the document just been created by last operation
			res = s.coll.FindOneAndUpdate(s.ctx(), filter, inc, opts)
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
