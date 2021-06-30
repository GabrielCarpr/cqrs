package context_test

import (
	"context"
	"encoding/json"
	cs "github.com/GabrielCarpr/cqrs/bus/context"
	"testing"

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

type ContextSerializerSuite struct {
	suite.Suite

	serializer *cs.ContextSerializer
}

func (s *ContextSerializerSuite) SetupTest() {
	s.serializer = cs.NewContextSerializer()
	s.serializer.Register(testKey, func(j []byte) interface{} {
		var v testVal
		json.Unmarshal(j, &v)
		return v
	})
	s.serializer.Register(stringKey, func(j []byte) interface{} {
		return string(j)
	})
	s.serializer.Register(mapKey, func(j []byte) interface{} {
		var m map[string]string
		json.Unmarshal(j, &m)
		return m
	})
}

func (s *ContextSerializerSuite) TestSerializesString() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, testKey, "hi")

	serial := s.serializer.Serialize(ctx)

	s.Len(serial, 1)
	s.Equal(serial["testKey"], "hi")
}

func (s *ContextSerializerSuite) TestSerializesMap() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, testKey, map[string]string{
		"hello": "lol",
	})

	serial := s.serializer.Serialize(ctx)

	s.Len(serial, 1)
	s.Equal(serial["testKey"], "{\"hello\":\"lol\"}")
}

type testStruct struct {
	First  string            `json:"first"`
	Second map[string]string `json:"second"`
}

func (s *ContextSerializerSuite) TestSerializesStruct() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, testKey, testStruct{
		First: "hi",
		Second: map[string]string{
			"bye": "cya",
		},
	})

	serial := s.serializer.Serialize(ctx)

	s.Len(serial, 1)
	s.Equal(serial["testKey"], "{\"first\":\"hi\",\"second\":{\"bye\":\"cya\"}}")
}

func (s *ContextSerializerSuite) TestDeserializeString() {
	serial := map[string]string{
		"stringKey": "hi",
	}

	ctx := s.serializer.Deserialize(context.Background(), serial)

	val := ctx.Value(stringKey)
	s.Equal("hi", val.(string))
}

func (s *ContextSerializerSuite) TestDeserializeMap() {
	serial := map[string]string{
		"mapKey": "{\"hello\":\"hi\"}",
	}

	ctx := s.serializer.Deserialize(context.Background(), serial)

	val := ctx.Value(mapKey)
	m := val.(map[string]string)
	s.Equal("hi", m["hello"])
}

func (s *ContextSerializerSuite) TestDeserializeStruct() {
	serial := map[string]string{
		"testKey": "{\"hello\":\"struct test\"}",
	}

	ctx := s.serializer.Deserialize(context.Background(), serial)

	val := ctx.Value(testKey)
	str := val.(testVal)
	s.Equal(str.Hello, "struct test")
}

func TestContextSerializer(t *testing.T) {
	suite.Run(t, new(ContextSerializerSuite))
}
