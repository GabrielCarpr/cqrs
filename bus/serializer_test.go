package bus_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type Key string

var testKey Key = "testKey"

var stringKey Key = "stringKey"

var mapKey Key = "mapKey"

func (k Key) String() string {
	return string(k)
}

type testVal struct {
	Hello string `json:"hello"`
}

type testCmd struct {
	bus.CommandType

	Name string
}

func (testCmd) Command() string {
	return "test.cmd"
}

type testEventSerial struct {
	bus.EventType

	Name string
}

func (testEventSerial) Event() string {
	return "test.event.serial"
}

type SerializerSuite struct {
	suite.Suite
}

func (s *SerializerSuite) SetupTest() {
	bus.RegisterContextKey(testKey, func(j []byte) interface{} {
		var v testVal
		json.Unmarshal(j, &v)
		return v
	})
	bus.RegisterContextKey(stringKey, func(j []byte) interface{} {
		return string(j)
	})
	bus.RegisterContextKey(mapKey, func(j []byte) interface{} {
		var m map[string]string
		json.Unmarshal(j, &m)
		return m
	})

	bus.RegisterMessage(&testEventSerial{})
	bus.RegisterMessage(testCmd{})
}

func (s *SerializerSuite) TestSerializesString() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, testKey, "hi")

	serial := bus.SerializeContext(ctx)

	s.Len(serial, 1)
	s.Equal(serial["testKey"], "hi")
}

func (s *SerializerSuite) TestSerializesMap() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, testKey, map[string]string{
		"hello": "lol",
	})

	serial := bus.SerializeContext(ctx)

	s.Len(serial, 1)
	s.Equal(serial["testKey"], "{\"hello\":\"lol\"}")
}

type testStruct struct {
	First  string            `json:"first"`
	Second map[string]string `json:"second"`
}

func (s *SerializerSuite) TestSerializesStruct() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, testKey, testStruct{
		First: "hi",
		Second: map[string]string{
			"bye": "cya",
		},
	})

	serial := bus.SerializeContext(ctx)

	s.Len(serial, 1)
	s.Equal(serial["testKey"], "{\"first\":\"hi\",\"second\":{\"bye\":\"cya\"}}")
}

func (s *SerializerSuite) TestDeserializeString() {
	serial := map[string]string{
		"stringKey": "hi",
	}

	ctx := bus.DeserializeContext(context.Background(), serial)

	val := ctx.Value(stringKey)
	s.Equal("hi", val.(string))
}

func (s *SerializerSuite) TestDeserializeMap() {
	serial := map[string]string{
		"mapKey": "{\"hello\":\"hi\"}",
	}

	ctx := bus.DeserializeContext(context.Background(), serial)

	val := ctx.Value(mapKey)
	m := val.(map[string]string)
	s.Equal("hi", m["hello"])
}

func (s *SerializerSuite) TestDeserializeStruct() {
	serial := map[string]string{
		"testKey": "{\"hello\":\"struct test\"}",
	}

	ctx := bus.DeserializeContext(context.Background(), serial)

	val := ctx.Value(testKey)
	str := val.(testVal)
	s.Equal(str.Hello, "struct test")
}

func (s *SerializerSuite) TestSerializeCommandAsGob() {
	cmd := testCmd{Name: "Hello"}

	data, err := bus.SerializeMessage(cmd, bus.Gob)
	s.Require().NoError(err)
	s.Require().Greater(len(data), 0)

	msg, err := bus.DeserializeMessage(data)
	s.Require().NoError(err)

	cmd = msg.(testCmd)
	s.Equal("Hello", cmd.Name)
}

func (s *SerializerSuite) TestSerializeEventAsGob() {
	event := &testEventSerial{Name: "Hi"}

	data, err := bus.SerializeMessage(event, bus.Gob)
	s.Require().NoError(err)
	s.Require().Greater(len(data), 0)

	msg, err := bus.DeserializeMessage(data)
	s.Require().NoError(err)

	event = msg.(*testEventSerial)
	s.Equal("Hi", event.Name)
}

func (s *SerializerSuite) TestSerializeCommandAsJson() {
	cmd := testCmd{Name: "Hello"}

	data, err := bus.SerializeMessage(cmd, bus.Json)
	s.Require().NoError(err)
	s.Require().True(json.Valid(data))
	s.T().Log(string(data))
	var j map[string]interface{}
	err = json.Unmarshal(data, &j)
	s.Require().NoError(err)
	_, ok := j["__type"]
	s.True(ok)
	s.Equal("Hello", j["Name"])

	msg, err := bus.DeserializeMessage(data)
	s.Require().NoError(err)
	cmd = msg.(testCmd)
	s.Equal("Hello", cmd.Name)
}

func (s *SerializerSuite) TestSerializeEventAsJson() {
	id := uuid.New()
	buffer := bus.NewEventBuffer(id, "testEvent")
	buffer.Buffer(true, &testEventSerial{Name: "Lol"})
	event := buffer.Events()[0]

	data, err := bus.SerializeMessage(event, bus.Json)
	s.Require().NoError(err)
	s.Require().True(json.Valid(data))
	s.T().Log(string(data))
	var j map[string]interface{}
	err = json.Unmarshal(data, &j)
	s.Require().NoError(err)
	_, ok := j["__type"]
	s.True(ok)
	s.Equal("Lol", j["Name"])

	msg, err := bus.DeserializeMessage(data)
	s.Require().NoError(err)
	e := msg.(*testEventSerial)
	s.Equal("Lol", e.Name)
	s.Equal(int64(1), e.Versioned())
	s.Equal(id.String(), e.Owner)
	s.Equal("testEvent", e.Aggregate)
}

func TestSerializer(t *testing.T) {
	suite.Run(t, new(SerializerSuite))
}
