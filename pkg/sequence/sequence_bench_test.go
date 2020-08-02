package sequence_test

import (
	"context"
	"github.com/hendratommy/mongo-sequence/pkg/sequence"
	"github.com/stretchr/testify/assert"
	"testing"
)

func BenchmarkDefaultSequence(b *testing.B) {
	client := createClient()
	db := client.Database("sequence_test")
	// clean test db after test run
	defer func() {
		dropDB(db)
		_ = client.Disconnect(context.TODO())
	}()
	coll := db.Collection("sequences")
	sequence.SetupDefaultSequence(coll, timeout())

	n := b.N
	for i := 0; i < n; i++ {
		_, err := sequence.NextVal(sequence.DefaultSequenceName)
		assert.NoError(b, err)
	}

	val, err := sequence.NextVal(sequence.DefaultSequenceName)
	if assert.NoError(b, err) {
		assert.EqualValues(b, n+1, val)
	}
}

func BenchmarkDefaultSequence_Parallel(b *testing.B) {
	client := createClient()
	db := client.Database("sequence_test")
	// clean test db after test run
	defer func() {
		dropDB(db)
		_ = client.Disconnect(context.TODO())
	}()
	coll := db.Collection("sequences")
	sequence.SetupDefaultSequence(coll, timeout())

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := sequence.NextVal(sequence.DefaultSequenceName)
			assert.NoError(b, err)
		}
	})
}
