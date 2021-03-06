// Copyright (c) 2017 OysterPack, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package keyvalue

import (
	"strings"

	"fmt"

	"github.com/coreos/bbolt"
)

// BucketView is a read-only view of the Bucket
type BucketView interface {
	// Name returns the Bucket name
	Name() string

	// Path returns the bucket's path location
	Path() []string

	// Exists returns true if the bucket exists. It might have been deleted.
	Exists() bool

	// Get returns the value for the specified key
	// If the key does not exist, or if the key actually refers to a child Bucket, then nil is returned
	// If bucket does not exist, then by definition, the bucket has no data and nil is returned.
	Get(key string) []byte

	// Keys returns the keys stored in this bucket. Keys are sorted, thus seek may be used to seek a position to start iterating.
	// seek is optional - if specified, then seek moves the cursor to a given key and returns it. If the key does not exist then the next key is used.
	Keys(seek string, cancel <-chan struct{}) <-chan string

	// KeyValues iterates through all key-value pairs and returns them via the channel.
	// The cancel channel is used to terminate the iteration early by the client.
	KeyValues(seek string, cancel <-chan struct{}) <-chan *KeyValue

	// BucketViews iterate through the top-level children buckets and returns via the returned channel.
	// The cancel channel is used to terminate the iteration early by the client.
	BucketViews(cancel <-chan struct{}) <-chan BucketView

	// BucketView returns the bucket for the specified path. If the bucket does not exist, then nil is returned.
	// The path is traversed to locate the Bucket within its hierarchy.
	BucketView(path ...string) BucketView
}

type bucketView struct {
	path []string

	db *bolt.DB
}

func (a *bucketView) String() string {
	return fmt.Sprintf("[%s]", strings.Join(a.path, " -> "))
}

func (a *bucketView) Name() string {
	return a.path[len(a.path)-1]
}

func (a *bucketView) Path() []string {
	return a.path
}

func (a *bucketView) Exists() bool {
	data := make(chan bool, 1)

	a.db.View(func(tx *bolt.Tx) error {
		data <- lookupBucket(tx, a.path) != nil
		return nil
	})

	return <-data
}

// Get returns the value for the specified key
// If the key does not exist, or if the key actually refers to a child Bucket, then nil is returned
func (a *bucketView) Get(key string) []byte {
	data := make(chan []byte, 1)

	a.db.View(func(tx *bolt.Tx) error {
		b := lookupBucket(tx, a.path)
		if b == nil {
			data <- nil
			return nil
		}

		data <- b.Get([]byte(key))
		return nil
	})

	return <-data
}

// Keys returns the keys stored in this bucket. Keys are sorted, thus seek may be used to seek a position to start iterating.
// seek is optional - if specified, i.e., not "", then seek moves the cursor to a given key and returns it. If the key does not exist then the next key is used.
// cancel is option, i.e., it can be nil
func (a *bucketView) Keys(seek string, cancel <-chan struct{}) <-chan string {
	c := make(chan chan string)

	go a.db.View(func(tx *bolt.Tx) error {
		data := make(chan string)
		c <- data
		b := lookupBucket(tx, a.path)
		if b == nil {
			close(data)
			return nil
		}

		cursor := b.Cursor()
		if cancel != nil {
			for k, _ := cursor.Seek([]byte(seek)); k != nil; k, _ = cursor.Next() {
				select {
				case <-cancel:
					break
				default:
					data <- string(k)
				}
			}
		} else {
			for k, _ := cursor.Seek([]byte(seek)); k != nil; k, _ = cursor.Next() {
				data <- string(k)
			}
		}

		close(data)

		return nil
	})

	return <-c
}

// KeyValues iterates through all key-value pairs and returns them via the channel.
// seek is optional - if specified, i.e., not "", then seek moves the cursor to a given key and returns it. If the key does not exist then the next key is used.
// The cancel channel is optional, i.e., it can be nil. The cancel channel is used to terminate the iteration early by the client.
func (a *bucketView) KeyValues(seek string, cancel <-chan struct{}) <-chan *KeyValue {
	c := make(chan chan *KeyValue)

	go a.db.View(func(tx *bolt.Tx) error {
		data := make(chan *KeyValue)
		c <- data
		b := lookupBucket(tx, a.path)
		if b == nil {
			close(data)
			return nil
		}

		cursor := b.Cursor()
		if cancel != nil {
			for k, v := cursor.Seek([]byte(seek)); k != nil; k, v = cursor.Next() {
				select {
				case <-cancel:
					break
				default:
					data <- &KeyValue{string(k), v}
				}

			}
		} else {
			for k, v := cursor.Seek([]byte(seek)); k != nil; k, v = cursor.Next() {
				data <- &KeyValue{string(k), v}
			}
		}

		close(data)

		return nil
	})

	return <-c
}

func (a *bucketView) BucketViews(cancel <-chan struct{}) <-chan BucketView {
	c := make(chan chan BucketView)

	go a.db.View(func(tx *bolt.Tx) error {
		buckets := make(chan BucketView)
		c <- buckets
		b := lookupBucket(tx, a.path)
		if b == nil {
			close(buckets)
			return nil
		}

		cursor := b.Cursor()
		if cancel != nil {
			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
				select {
				case <-cancel:
					break
				default:
					if v == nil {
						buckets <- &bucketView{append(a.path, string(k)), a.db}
					}
				}
			}
		} else {
			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
				if v == nil {
					buckets <- &bucketView{append(a.path, string(k)), a.db}
				}
			}
		}

		close(buckets)

		return nil
	})

	return <-c
}

func (a *bucketView) BucketView(path ...string) BucketView {
	view := a.bucketView(path...)
	if view == nil {
		return nil
	}
	return view
}

func (a *bucketView) bucketView(path ...string) *bucketView {
	if len(path) == 0 {
		return a
	}

	b := make(chan *bucketView, 1)
	a.db.View(func(tx *bolt.Tx) error {
		parent := lookupBucket(tx, a.path)
		if parent == nil {
			b <- nil
			return nil
		}
		target := lookupChildBucket(tx, parent, path)
		if target == nil {
			b <- nil
			return nil
		}

		b <- &bucketView{append(a.path, path...), a.db}

		return nil
	})

	return <-b
}
