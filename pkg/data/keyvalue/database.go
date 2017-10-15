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
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/coreos/bbolt"
)

const (
	DATABASE_BUCKET = "db"
	DATABASE_NAME   = "name"

	READ_MODE       os.FileMode = 0400
	READ_WRITE_MODE os.FileMode = 0600
)

var (
	database_bucket = []byte(DATABASE_BUCKET)
	database_name   = []byte(DATABASE_NAME)

	database_path = []string{DATABASE_BUCKET}
)

// Database provides a read-write view of the database
type Database interface {
	Name() string

	// Buckets return the bucket names on the returned channel. Only Top-level buckets are returned.
	// The cancel channel is used to terminate the iteration early by the client.
	Buckets(cancel <-chan struct{}) <-chan Bucket

	// Bucket returns the bucket for the specified name. If the bucket does not exist, then nil is returned.
	// If path is specified, then the database will traverse the path to locate the Bucket within its hierarchy.
	Bucket(path ...string) Bucket

	// Close releases all database resources. All transactions must be closed before closing the database.
	Close() error
}

// OpenDatabase opens the database in read-write mode
// The path must point to a bbolt file. The file must have the following structure :
// - a root bucket must exist named 'db'
// - the root 'db' bucket must contain a key named 'name', which is the database name
func OpenDatabase(path string) (Database, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, ErrPathIsRequired
	}

	options := &bolt.Options{
		Timeout: time.Second * 30,
	}

	db, err := bolt.Open(path, READ_WRITE_MODE, options)
	if err != nil {
		return nil, err
	}

	var dbName []byte
	db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket(database_bucket)
		if root != nil {
			dbName = root.Get(database_name)
		} else {
			err = fmt.Errorf("Root database bucket does not exist : %s", DATABASE_BUCKET)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(dbName) == 0 {
		return nil, ErrDatabaseNameIsRequired
	}
	dbBucket := &bucket{&bucketView{name: string(dbName), path: database_path, db: db}}
	return &database{dbBucket}, nil
}

func CreateDatabaseIfNotExists(path string, name string) (Database, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, ErrPathIsRequired
	}

	options := &bolt.Options{
		Timeout: time.Second * 30,
	}

	db, err := bolt.Open(path, READ_WRITE_MODE, options)
	if err != nil {
		return nil, err
	}

	var dbName []byte
	err = db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists(database_bucket)
		if err != nil {
			return err
		}
		dbName = root.Get(database_name)
		if dbName == nil {
			if tx.Bucket(database_name) != nil {
				return ErrBucketWasFoundForDatabaseNameValue
			}
			dbName = []byte(name)
			root.Put([]byte(database_name), dbName)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if string(dbName) != name {
		return nil, databaseNameDoesNotMatch(name, string(dbName))
	}
	dbBucket := &bucket{&bucketView{name: string(dbName), path: database_path, db: db}}
	return &database{dbBucket}, nil
}

type database struct {
	*bucket
}

func (a *database) Name() string {
	return string(a.Get(DATABASE_NAME))
}

func (a *database) Close() error {
	return a.db.Close()
}
