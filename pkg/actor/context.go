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

type Context struct {
	// Message returns the current message to be processed
	Message

	// Actor returns the actor associated with this context
	*Actor
}

// SetBehaivor is used to alter the actor's behavior, i.e., how user messages are processed
func (a *Context) SetBehavior(behavior func(ctx *Context)) {
	a.behavior = behavior
}