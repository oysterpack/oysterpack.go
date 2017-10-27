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

package actor

import (
	"errors"
	"fmt"
)

// InvalidChannelError indicates the channel is unknown
type InvalidChannelError struct {
	Channel string
}

func (a *InvalidChannelError) Error() string {
	return a.String()
}

func (a *InvalidChannelError) String() string {
	return fmt.Sprintf("Invalid channel : %s", a.Channel)
}

// InvalidChannelMessageTypeError indicates the channel does not support the specified message type
type InvalidChannelMessageTypeError struct {
	Channel string
	Message interface{}
}

func (a *InvalidChannelMessageTypeError) Error() string {
	return a.String()
}

func (a *InvalidChannelMessageTypeError) String() string {
	return fmt.Sprintf("Channel (%s) does not support message type : %T", a.Channel, a.Message)
}

// InvalidMessageError indicates the message invalid.
type InvalidMessageError struct {
	Channel string
	Message interface{}
	Err     error
}

func (a *InvalidMessageError) Error() string {
	return a.String()
}

func (a *InvalidMessageError) String() string {
	return fmt.Sprintf("Message was invalid : %q[%T] : %v", a.Channel, a.Message, a.Err.Error())
}

// ActorNotAliveError indicates an invalid operation was performed on an actor that is not alive.
type ActorNotAliveError struct {
	Address
}

func (a *ActorNotAliveError) Error() string {
	return a.String()
}

func (a *ActorNotAliveError) String() string {
	return fmt.Sprintf("Actor is not alive : %s", a.Address)
}

var (
	ErrStillAlive = errors.New("Actor is still alive")
)

type ChildAlreadyExists struct {
	Path []string
}

func (a *ChildAlreadyExists) Error() string {
	return a.String()
}

func (a *ChildAlreadyExists) String() string {
	return fmt.Sprintf("Child already exists at path : %v", a.Path)
}

type ProducerError struct {
	Err error
}

func (a *ProducerError) Error() string {
	return a.String()
}

func (a *ProducerError) String() string {
	return fmt.Sprintf("Child already exists at path : %v", a.Err)
}