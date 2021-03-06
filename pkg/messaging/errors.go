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

package messaging

import "errors"

var (
	// ErrTopicMustNotBeBlank - Topic must not be blank
	ErrTopicMustNotBeBlank = errors.New("Topic must not be blank")
	// ErrReplyToMustNotBeBlank - ReplyTo must not be blank
	ErrReplyToMustNotBeBlank = errors.New("ReplyTo must not be blank")
	// ErrConnectionIsClosed - Conn is closed
	ErrConnectionIsClosed = errors.New("Conn is closed")
	// ErrClientAlreadyRegsiteredForSameCluster - A Client is already registered for the same cluster
	ErrClientAlreadyRegsiteredForSameCluster = errors.New("A Client is already registered for the same cluster")
)
