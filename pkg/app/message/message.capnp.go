// Code generated by capnpc-go. DO NOT EDIT.

package message

import (
	strconv "strconv"

	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
)

type Message struct{ capnp.Struct }
type Message_deadline Message
type Message_deadline_Which uint16

const (
	Message_deadline_Which_timeoutMSec Message_deadline_Which = 0
	Message_deadline_Which_expiresOn   Message_deadline_Which = 1
)

func (w Message_deadline_Which) String() string {
	const s = "timeoutMSecexpiresOn"
	switch w {
	case Message_deadline_Which_timeoutMSec:
		return s[0:11]
	case Message_deadline_Which_expiresOn:
		return s[11:20]

	}
	return "Message_deadline_Which(" + strconv.FormatUint(uint64(w), 10) + ")"
}

// Message_TypeID is the unique identifier for the type Message.
const Message_TypeID = 0xc768aaf640842a35

func NewMessage(s *capnp.Segment) (Message, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 48, PointerCount: 1})
	return Message{st}, err
}

func NewRootMessage(s *capnp.Segment) (Message, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 48, PointerCount: 1})
	return Message{st}, err
}

func ReadRootMessage(msg *capnp.Message) (Message, error) {
	root, err := msg.RootPtr()
	return Message{root.Struct()}, err
}

func (s Message) String() string {
	str, _ := text.Marshal(0xc768aaf640842a35, s.Struct)
	return str
}

func (s Message) Id() uint64 {
	return s.Struct.Uint64(0)
}

func (s Message) SetId(v uint64) {
	s.Struct.SetUint64(0, v)
}

func (s Message) Type() uint64 {
	return s.Struct.Uint64(8)
}

func (s Message) SetType(v uint64) {
	s.Struct.SetUint64(8, v)
}

func (s Message) CorrelationID() uint64 {
	return s.Struct.Uint64(16)
}

func (s Message) SetCorrelationID(v uint64) {
	s.Struct.SetUint64(16, v)
}

func (s Message) Timestamp() int64 {
	return int64(s.Struct.Uint64(24))
}

func (s Message) SetTimestamp(v int64) {
	s.Struct.SetUint64(24, uint64(v))
}

func (s Message) Compression() Message_Compression {
	return Message_Compression(s.Struct.Uint16(32) ^ 1)
}

func (s Message) SetCompression(v Message_Compression) {
	s.Struct.SetUint16(32, uint16(v)^1)
}

func (s Message) Data() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return []byte(p.Data()), err
}

func (s Message) HasData() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetData(v []byte) error {
	return s.Struct.SetData(0, v)
}

func (s Message) Deadline() Message_deadline { return Message_deadline(s) }

func (s Message_deadline) Which() Message_deadline_Which {
	return Message_deadline_Which(s.Struct.Uint16(36))
}
func (s Message_deadline) TimeoutMSec() uint16 {
	if s.Struct.Uint16(36) != 0 {
		panic("Which() != timeoutMSec")
	}
	return s.Struct.Uint16(34)
}

func (s Message_deadline) SetTimeoutMSec(v uint16) {
	s.Struct.SetUint16(36, 0)
	s.Struct.SetUint16(34, v)
}

func (s Message_deadline) ExpiresOn() int64 {
	if s.Struct.Uint16(36) != 1 {
		panic("Which() != expiresOn")
	}
	return int64(s.Struct.Uint64(40))
}

func (s Message_deadline) SetExpiresOn(v int64) {
	s.Struct.SetUint16(36, 1)
	s.Struct.SetUint64(40, uint64(v))
}

// Message_List is a list of Message.
type Message_List struct{ capnp.List }

// NewMessage creates a new list of Message.
func NewMessage_List(s *capnp.Segment, sz int32) (Message_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 48, PointerCount: 1}, sz)
	return Message_List{l}, err
}

func (s Message_List) At(i int) Message { return Message{s.List.Struct(i)} }

func (s Message_List) Set(i int, v Message) error { return s.List.SetStruct(i, v.Struct) }

func (s Message_List) String() string {
	str, _ := text.MarshalList(0xc768aaf640842a35, s.List)
	return str
}

// Message_Promise is a wrapper for a Message promised by a client call.
type Message_Promise struct{ *capnp.Pipeline }

func (p Message_Promise) Struct() (Message, error) {
	s, err := p.Pipeline.Struct()
	return Message{s}, err
}

func (p Message_Promise) Deadline() Message_deadline_Promise {
	return Message_deadline_Promise{p.Pipeline}
}

// Message_deadline_Promise is a wrapper for a Message_deadline promised by a client call.
type Message_deadline_Promise struct{ *capnp.Pipeline }

func (p Message_deadline_Promise) Struct() (Message_deadline, error) {
	s, err := p.Pipeline.Struct()
	return Message_deadline{s}, err
}

type Message_Compression uint16

// Message_Compression_TypeID is the unique identifier for the type Message_Compression.
const Message_Compression_TypeID = 0xf8f433c185247295

// Values of Message_Compression.
const (
	Message_Compression_none Message_Compression = 0
	Message_Compression_zlib Message_Compression = 1
)

// String returns the enum's constant name.
func (c Message_Compression) String() string {
	switch c {
	case Message_Compression_none:
		return "none"
	case Message_Compression_zlib:
		return "zlib"

	default:
		return ""
	}
}

// Message_CompressionFromString returns the enum value with a name,
// or the zero value if there's no such value.
func Message_CompressionFromString(c string) Message_Compression {
	switch c {
	case "none":
		return Message_Compression_none
	case "zlib":
		return Message_Compression_zlib

	default:
		return 0
	}
}

type Message_Compression_List struct{ capnp.List }

func NewMessage_Compression_List(s *capnp.Segment, sz int32) (Message_Compression_List, error) {
	l, err := capnp.NewUInt16List(s, sz)
	return Message_Compression_List{l.List}, err
}

func (l Message_Compression_List) At(i int) Message_Compression {
	ul := capnp.UInt16List{List: l.List}
	return Message_Compression(ul.At(i))
}

func (l Message_Compression_List) Set(i int, v Message_Compression) {
	ul := capnp.UInt16List{List: l.List}
	ul.Set(i, uint16(v))
}

type Ping struct{ capnp.Struct }

// Ping_TypeID is the unique identifier for the type Ping.
const Ping_TypeID = 0x9bce611bc724ff89

func NewPing(s *capnp.Segment) (Ping, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Ping{st}, err
}

func NewRootPing(s *capnp.Segment) (Ping, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Ping{st}, err
}

func ReadRootPing(msg *capnp.Message) (Ping, error) {
	root, err := msg.RootPtr()
	return Ping{root.Struct()}, err
}

func (s Ping) String() string {
	str, _ := text.Marshal(0x9bce611bc724ff89, s.Struct)
	return str
}

// Ping_List is a list of Ping.
type Ping_List struct{ capnp.List }

// NewPing creates a new list of Ping.
func NewPing_List(s *capnp.Segment, sz int32) (Ping_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0}, sz)
	return Ping_List{l}, err
}

func (s Ping_List) At(i int) Ping { return Ping{s.List.Struct(i)} }

func (s Ping_List) Set(i int, v Ping) error { return s.List.SetStruct(i, v.Struct) }

func (s Ping_List) String() string {
	str, _ := text.MarshalList(0x9bce611bc724ff89, s.List)
	return str
}

// Ping_Promise is a wrapper for a Ping promised by a client call.
type Ping_Promise struct{ *capnp.Pipeline }

func (p Ping_Promise) Struct() (Ping, error) {
	s, err := p.Pipeline.Struct()
	return Ping{s}, err
}

type Pong struct{ capnp.Struct }

// Pong_TypeID is the unique identifier for the type Pong.
const Pong_TypeID = 0xf6486a286fedf2f6

func NewPong(s *capnp.Segment) (Pong, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Pong{st}, err
}

func NewRootPong(s *capnp.Segment) (Pong, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Pong{st}, err
}

func ReadRootPong(msg *capnp.Message) (Pong, error) {
	root, err := msg.RootPtr()
	return Pong{root.Struct()}, err
}

func (s Pong) String() string {
	str, _ := text.Marshal(0xf6486a286fedf2f6, s.Struct)
	return str
}

// Pong_List is a list of Pong.
type Pong_List struct{ capnp.List }

// NewPong creates a new list of Pong.
func NewPong_List(s *capnp.Segment, sz int32) (Pong_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0}, sz)
	return Pong_List{l}, err
}

func (s Pong_List) At(i int) Pong { return Pong{s.List.Struct(i)} }

func (s Pong_List) Set(i int, v Pong) error { return s.List.SetStruct(i, v.Struct) }

func (s Pong_List) String() string {
	str, _ := text.MarshalList(0xf6486a286fedf2f6, s.List)
	return str
}

// Pong_Promise is a wrapper for a Pong promised by a client call.
type Pong_Promise struct{ *capnp.Pipeline }

func (p Pong_Promise) Struct() (Pong, error) {
	s, err := p.Pipeline.Struct()
	return Pong{s}, err
}

const schema_aa44738dedfed9a1 = "x\xda\x9c\x92\xcfKTQ\x1c\xc5\xcf\xb9\xf7\xbd\x19\x85" +
	"t\xba\xbe\x97Ai\xe3\xc2\x85\xce\"\x12\x93\xa4\x8dF" +
	".*\x90\xbcJ\xd0*x\xce\\\xec\x85\xf3\xde03" +
	"\xa6\xb9\xce\x85A\xb4\x886\x16\x84\xcb\\\xf5s\x11(" +
	"\xe4B\xacE\xed\x0a\xfa\x1fr\x13*\x12\xe4\x8d\xa7\xe9" +
	"\x98\xb5j\xf9=\xf7r\xbe\xf7~\xf8\x9c\xe9b\xbf\xe8" +
	"r'%\xa0;\xdc\x94\x9d\xb5\xed\xab'\x83O\x8f\xa1" +
	"\x1ai\xe7\xbfn\xaf\xdd\xaf\x0c,\xc0I\x03\xde<\xe7" +
	"\xbcgLC\xda\x9e\xdb\xcb\x1b\xa7z_=\x81nf" +
	"\xca\xf6\xe4\xee\xf6o.\xdc\\\xc55\xa6)\xd8\xe4\xcd" +
	"\xf0\x07\xe8\xcdr\x12\xac\x1d\xeaF\xa6j\x85.\x93\xc6" +
	"c\xe2\xa5\xd7*\x8e\x03\xdd\x9d\xe2\x01A\xbb\xf9}-" +
	"\xee\xb8ui\xf3\x1f\xdb?\xca9\xef\xb3L\xb6?*" +
	"\xb7\xcf,w\xafoA5\x8bZ?\xd8\xfdB6\xd1" +
	"[\x96\xc9\xed%y\x0e\xbd\xb6h*\x95`\xcc\x9cf" +
	">(E\xa5\xf3C\xa1\x8c\xc6\x86\xc8\xfd\\\xec\xe6\x83" +
	"\xbf\xc7\x82\xe9\x0b\x0a\xe3adt\x9dtZ\xacU>" +
	"S\x80\xea\x1cM\xd8H\xea\xb3\x82\xad\xdc\xb6\xae\x9f<" +
	"^u\x0d\xab\x9e\xac\xbe.\xa9\x0b\x82\xb6\x1a\x16M<" +
	"Q\x1dDz\xc4\xe4\x99\x86`\x1a\xb4f\xaa\x14\x96M" +
	"\xe5*\x18i\x87\xc2\xdex\xf8T/}\xb9\xb7\x02\xed" +
	"\x08^\xf0\xc9#\x80\xe2\xb4\x9d\x88\xc2\xa9\xb6(\x88\xd0" +
	"\x17\xb7%M\x00]\x08\xba\xe0\xe1?\x0c\x9a\xec\xce\xac" +
	"\x1d\xf2\x00\x09\x8e\xda\x8bq\xb1T6\x95\x0a\xd2a\x1c" +
	"\xe9\x16\xe9\x00\x0e\x01\xf5\xe6\x04\xa0\x9fK\xeaEAE" +
	"\xfaL\xc2\xb79@\xbf\x96\xd4\xef\x04\x95\x10>\x05\xa0" +
	"\x96\xca\x80^\x94\xd4\xef\x05\x95\x94>%\xa0V\x86\xd5" +
	"\x87\xac\xfe&\xa9\xb7\x04\x95s\xd4\xa7C\xaa\x8d\x84\xc9" +
	"\xba\xe40\x05\xe9\xfat\x01\xf53\xe9\xdc\x92\x1cq\x92" +
	"0E\xd6<\xf1\xc8+\x102,\xb0\x1e\x82\xf5`\xa6" +
	"z\xa7d\xf6\x06\x9b\x8f\xcbe3\x1eT\x91\x0d\xe3\xe8" +
	"\xf2\xc0~\x9e\xa0\xa8T\x83\"X\xfa_|\xf9\x83T" +
	"\x98\xa9!\x03\xdd\x0c\xc1L!\xa8\x06l\x80`\x03h" +
	"\x0bfW\x00\x00\x7f\xc9\x13\x1f\x92G\xfe)\xcf\x1e\xfe" +
	"0f4D\xea\xba\x1d\xa4*\x07\x90\xaa>\x07d\xa2" +
	"82\x99\xe9\xf1p\xf4W\x00\x00\x00\xff\xff_\x95\xf6" +
	"\x86"

func init() {
	schemas.Register(schema_aa44738dedfed9a1,
		0x9bce611bc724ff89,
		0x9cb3381ef5c17635,
		0xc768aaf640842a35,
		0xf6486a286fedf2f6,
		0xf8f433c185247295)
}
