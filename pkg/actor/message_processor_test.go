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

package actor_test

import (
	"testing"

	"errors"

	"github.com/oysterpack/oysterpack.go/pkg/actor"
	"github.com/rs/zerolog/log"
)

func TestStartMessageProcessorEngine(t *testing.T) {
	foo := actor.MessageHandlers{
		actor.MessageChannelKey{actor.CHANNEL_SYSTEM, actor.MESSAGE_TYPE_DEFAULT}: actor.MessageHandler{
			Receive: func(ctx *actor.MessageContext) error {
				t.Logf("Received message: %v", ctx.Envelope)
				return nil
			},
			Unmarshal: func(msg []byte) (*actor.Envelope, error) { return nil, errors.New("NOT SUPPORTED") },
		},
		actor.MessageChannelKey{actor.CHANNEL_LIFECYCLE, actor.MESSAGE_TYPE_DEFAULT}: actor.MessageHandler{
			Receive: func(ctx *actor.MessageContext) error {
				t.Logf("Received message: %v", ctx.Envelope)
				return nil
			},
			Unmarshal: func(msg []byte) (*actor.Envelope, error) { return nil, errors.New("NOT SUPPORTED") },
		},
	}

	processor, err := actor.StartMessageProcessorEngine(foo, log.Logger)
	if err != nil {
		t.Fatal(err)
	}

	if len(processor.ChannelNames()) != 2 {
		t.Errorf("Channel count is wrong : %v", processor.ChannelNames())
	}

	if !processor.Alive() {
		t.Error("processor should be alive")
	}

	msg := actor.PING_REQ
	processor.Channel() <- &actor.MessageContext{
		Actor:    nil,
		Envelope: actor.NewEnvelope(uid, actor.CHANNEL_SYSTEM, actor.SYS_MSG_PING_REQ, msg, nil),
	}

	processor.Kill(nil)
	deathReason := processor.Wait()
	t.Logf("death reason : %v", deathReason)

}