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

	"time"

	"github.com/oysterpack/oysterpack.go/pkg/actor"
	"github.com/rs/zerolog/log"
)

func TestNewSystem(t *testing.T) {
	system, err := actor.NewSystem("oysterpack", log.Logger)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		log.Logger.Info().Msg("Killing system ...")
		system.Kill(nil)
		log.Logger.Info().Msg("Kill signalled. Waiting for system to terminate ...")
		if err := system.Wait(); err != nil {
			t.Error(err)
		}
		log.Logger.Info().Msg("System terminated")
	}()

	log.Logger.Info().Msgf("address = %v, name = %s, path = %v, id = %s", system.Address(), system.Name(), system.Path(), system.Id())

	if !system.Alive() {
		t.Error("system is not alive")
	}

	if system.Name() != "oysterpack" {
		t.Errorf("name did not match : %v", system.Name())
	}

	bar, err := system.Spawn("bar",
		barMessageProcessor,
		actor.RESTART_ACTOR_STRATEGY,
		log.Logger)

	if err != nil {
		t.Fatalf("Failed to spawn the bar actor : %v", err)
	}

	foo, err := system.Spawn("foo",
		func() actor.MessageProcessor {
			return fooMessageProcessor(10, bar.Address())
		},
		actor.RESTART_ACTOR_STRATEGY,
		log.Logger)

	if err != nil {
		t.Fatalf("Failed to spawn the foo actor : %v", err)
	}

	if err := foo.Wait(); err != nil {
		t.Errorf("foo actor died with an error : v", err)
	}

}

func fooMessageProcessor(count int, address *actor.Address) actor.MessageProcessor {
	return actor.MessageHandlers{
		actor.MessageChannelKey{actor.CHANNEL_INBOX, actor.SYS_MSG_HEARTBEAT_RESP}: actor.MessageHandler{func(ctx *actor.MessageContext) error {
			// when a heartbeat response is received, send another heartbeat request
			heartbeatReq := ctx.RequestEnvelope(ctx.Envelope.ReplyTo().Channel, actor.SYS_MSG_HEARTBEAT_REQ, actor.PING_REQ, actor.CHANNEL_INBOX)
			if err := ctx.Send(heartbeatReq, ctx.Envelope.ReplyTo().Address); err != nil {
				return err
			}
			count--
			ctx.Logger().Debug().Int("count", count).Msg("")
			if count == 0 {
				ctx.Logger().Debug().Msg("foo is done")
				ctx.Kill(nil)
			}
			return nil
		},
			actor.UnmarshalHeartbeatResponse,
		},
		actor.MessageChannelKey{actor.CHANNEL_LIFECYCLE, actor.LIFECYCLE_MSG_STARTED}: actor.MessageHandler{
			func(ctx *actor.MessageContext) error {
				// when the actor is started, send the bar actor a heartbeat request to initiate the communication
				var ref *actor.Actor
				for ; ref == nil; ref = ctx.System().LookupActor(address) {
					time.Sleep(100 * time.Millisecond)
				}
				heartbeatReq := ctx.RequestEnvelope(actor.CHANNEL_SYSTEM, actor.SYS_MSG_HEARTBEAT_REQ, actor.PING_REQ, actor.CHANNEL_INBOX)
				ref.Tell(heartbeatReq)
				return nil
			},
			actor.UnmarshalStartedMessage,
		},
	}
}

func barMessageProcessor() actor.MessageProcessor {
	return actor.MessageHandlers{
		actor.MessageChannelKey{actor.CHANNEL_SYSTEM, actor.SYS_MSG_HEARTBEAT_REQ}: actor.MessageHandler{
			actor.HandleHearbeatRequest,
			actor.UnmarshalHeartbeatRequest,
		},
	}
}

// Benchmark results
// =================
//
//	BenchmarkActorMessaging/Heartbeat_Request/Response_-_MessageChannelHandlers_based-8               300000              4496 ns/op             352 B/op          8 allocs/op
//	BenchmarkActorMessaging/Heartbeat_Request/Response_-_switch_based-8                               300000              3999 ns/op             368 B/op          9 allocs/op
//	BenchmarkActorMessaging/Heartbeat_Messaging_-_fire_and_forget-8                                   200000              7497 ns/op             662 B/op          6 allocs/op
//
//	BenchmarkActors/Heartbeat_Messaging_-_MessageChannelHandlers_based-8              300000              4290 ns/op             352 B/op          8 allocs/op
//
//		oysterpack/foo --HeartbeatRequest--> oysterpack/bar
//		oysterpack/foo <--HeartbeatResponse-- oysterpack/bar
//
//	BenchmarkActors/Heartbeat_Messaging_-_switch_based-8                              300000              4165 ns/op             368 B/op          9 allocs/op
//
//		oysterpack/foo --HeartbeatRequest--> oysterpack
//		oysterpack/foo <--HeartbeatResponse-- oysterpack
//
// The map based MessageChannelHandlers performs just as well as using native.
func BenchmarkActorMessaging(b *testing.B) {
	system, err := actor.NewSystem("oysterpack", log.Logger)
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		log.Logger.Info().Msg("Killing system ...")
		system.Kill(nil)
		log.Logger.Info().Msg("Kill signalled. Waiting for system to terminate ...")
		if err := system.Wait(); err != nil {
			b.Error(err)
		}
		log.Logger.Info().Msg("System terminated")
	}()

	bar, err := system.Spawn("bar",
		barMessageProcessor,
		actor.RESTART_ACTOR_STRATEGY,
		log.Logger)

	if err != nil {
		b.Fatalf("Failed to spawn the bar actor : %v", err)
	}

	b.Run("Heartbeat Request/Response - MessageChannelHandlers based", func(b *testing.B) {
		foo, err := system.Spawn("foo",
			func() actor.MessageProcessor {
				return fooMessageProcessor(b.N, bar.Address())
			},
			actor.RESTART_ACTOR_STRATEGY,
			log.Logger)
		if err != nil {
			b.Fatalf("Failed to spawn the foo actor : %v", err)
		}
		b.ResetTimer()
		foo.Wait()
	})

	b.Run("Heartbeat Request/Response - switch based", func(b *testing.B) {
		foo, err := system.Spawn("foo",
			func() actor.MessageProcessor {
				return &FooMessageProcessor{b.N, system.Address()}
			},
			actor.RESTART_ACTOR_STRATEGY,
			log.Logger)
		if err != nil {
			b.Fatalf("Failed to spawn the foo actor : %v", err)
		}
		b.ResetTimer()
		foo.Wait()
	})

	b.Run("Heartbeat Messaging - fire and forget", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			system.Tell(system.MessageEnvelope(actor.CHANNEL_SYSTEM, actor.SYS_MSG_HEARTBEAT_REQ, actor.HEARTBEAT_REQ))
		}
	})

	b.Run("Heartbeat Request/Response - MessageChannelHandlers based - buffered chan", func(b *testing.B) {
		bar.Kill(nil)
		bar.Wait()

		bar, err = system.Spawn("bar",
			barMessageProcessor,
			actor.RESTART_ACTOR_STRATEGY,
			log.Logger,
			&actor.ChannelSettings{Channel: actor.CHANNEL_SYSTEM, BufSize: 1})
		if err != nil {
			b.Fatalf("Failed to spawn the foo actor : %v", err)
		}

		foo, err := system.Spawn("foo",
			func() actor.MessageProcessor {
				return fooMessageProcessor(b.N, bar.Address())
			},
			actor.RESTART_ACTOR_STRATEGY,
			log.Logger,
			&actor.ChannelSettings{Channel: actor.CHANNEL_INBOX, BufSize: 1})
		if err != nil {
			b.Fatalf("Failed to spawn the foo actor : %v", err)
		}

		b.ResetTimer()
		foo.Wait()
	})

	b.Run("Heartbeat Messaging - fire and forget - buffered chan", func(b *testing.B) {
		bar.Kill(nil)
		bar.Wait()

		bar, err = system.Spawn("bar",
			barMessageProcessor,
			actor.RESTART_ACTOR_STRATEGY,
			log.Logger)
		if err != nil {
			b.Fatalf("Failed to spawn the foo actor : %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bar.Tell(bar.MessageEnvelope(actor.CHANNEL_SYSTEM, actor.SYS_MSG_HEARTBEAT_REQ, actor.HEARTBEAT_REQ))
		}
	})
}

type FooMessageProcessor struct {
	count   int
	address *actor.Address
}

// ChannelNames returns the channels that are supported by this message processor
func (a *FooMessageProcessor) ChannelNames() []actor.Channel {
	return []actor.Channel{actor.CHANNEL_INBOX, actor.CHANNEL_LIFECYCLE}
}

// MessageTypes returns the message types that are supported for the specified channel
func (a *FooMessageProcessor) MessageTypes(channel actor.Channel) []actor.MessageType {
	switch channel {
	case actor.CHANNEL_INBOX:
		return []actor.MessageType{actor.SYS_MSG_HEARTBEAT_RESP}
	case actor.CHANNEL_LIFECYCLE:
		return []actor.MessageType{actor.LIFECYCLE_MSG_STARTED}
	default:
		return nil
	}
}

// Handler returns a Receive function that will be used to handle messages on the specified channel for the
// specified message type
func (a *FooMessageProcessor) Handler(channel actor.Channel, msgType actor.MessageType) actor.Receive {
	switch channel {
	case actor.CHANNEL_INBOX:
		switch msgType {
		case actor.SYS_MSG_HEARTBEAT_RESP:
			return a.handleHeartbeatResponse
		default:
			return nil
		}
	case actor.CHANNEL_LIFECYCLE:
		switch msgType {
		case actor.LIFECYCLE_MSG_STARTED:
			return a.started
		default:
			return nil
		}
	default:
		return nil
	}
}

func (a *FooMessageProcessor) started(ctx *actor.MessageContext) error {
	// when the actor is started, send the bar actor a heartbeat request to initiate the communication
	var ref *actor.Actor
	for ; ref == nil; ref = ctx.System().LookupActor(a.address) {
		time.Sleep(100 * time.Millisecond)
	}
	heartbeatReq := ctx.RequestEnvelope(actor.CHANNEL_SYSTEM, actor.SYS_MSG_HEARTBEAT_REQ, actor.PING_REQ, actor.CHANNEL_INBOX)
	ref.Tell(heartbeatReq)
	return nil
}

func (a *FooMessageProcessor) handleHeartbeatResponse(ctx *actor.MessageContext) error {
	heartbeatReq := ctx.RequestEnvelope(ctx.Envelope.ReplyTo().Channel, actor.SYS_MSG_HEARTBEAT_REQ, actor.PING_REQ, actor.CHANNEL_INBOX)
	if err := ctx.Send(heartbeatReq, ctx.Envelope.ReplyTo().Address); err != nil {
		return err
	}
	a.count--
	ctx.Logger().Debug().Int("count", a.count).Msg("")
	if a.count == 0 {
		ctx.Logger().Debug().Msg("foo is done")
		ctx.Kill(nil)
	}
	return nil
}

func (a *FooMessageProcessor) UnmarshalMessage(channel actor.Channel, msgType actor.MessageType, msg []byte) (*actor.Envelope, error) {
	switch channel {
	case actor.CHANNEL_INBOX:
		switch msgType {
		case actor.SYS_MSG_HEARTBEAT_RESP:
			return actor.UnmarshalHeartbeatResponse(msg)
		default:
			return nil, &actor.ChannelMessageTypeNotSupportedError{channel, msgType}
		}
	default:
		return nil, &actor.ChannelNotSupportedError{channel}
	}
}

// Stopped is invoked after the message processor is stopped. The message processor should perform any cleanup here.
func (a *FooMessageProcessor) Stopped() {}