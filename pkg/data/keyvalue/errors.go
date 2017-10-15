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
	"errors"
	"fmt"
)

var (
	ErrPathIsRequired                     = errors.New("Path is required")
	ErrDatabaseNameIsRequired             = errors.New("Database name is required")
	ErrBucketWasFoundForDatabaseNameValue = errors.New("A bucket was stored using the database 'name' key in the root 'db' bucket")
)

func bucketDoesNotExist(path []string) error {
	return fmt.Errorf("Bucket does not exist at path : %v", path)
}

func databaseNameDoesNotMatch(expected, actual string) error {
	return fmt.Errorf("Database name does not match : expected = %q, actual = %q", expected, actual)
}
