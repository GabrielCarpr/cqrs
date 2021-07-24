package bus

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/GabrielCarpr/cqrs/bus/message"
)

const (
	Gob SerialType = iota + 1
	Json
)

type SerialType int

type SerializedContext map[string]string

type contextRegistration struct {
	Key     fmt.Stringer
	Convert func(json []byte) interface{}
}

var contextMap map[string]contextRegistration
var messageMap map[string]reflect.Type

func init() {
	contextMap = make(map[string]contextRegistration)
	messageMap = make(map[string]reflect.Type)
}

func msgKey(msg message.Message) string {
	return reflect.TypeOf(msg).String()
}

type typeReader struct {
	Type string `json:"__type"`
}

func RegisterMessage(msg message.Message) {
	gob.Register(&msg)
	gob.Register(msg)

	messageMap[msgKey(msg)] = reflect.TypeOf(msg)
}

func RegisterContextKey(key fmt.Stringer, fn func([]byte) interface{}) {
	contextMap[key.String()] = contextRegistration{Key: key, Convert: fn}
}

func SerializeContext(ctx context.Context) SerializedContext {
	result := make(map[string]string)

	for name, reg := range contextMap {
		val := ctx.Value(reg.Key)
		if val == nil {
			continue
		}
		str, ok := val.(string)
		if ok {
			result[name] = str
			continue
		}
		bytes, err := json.Marshal(val)
		if err != nil {
			log.Printf("Error serialising context value [%s]: %s", name, err)
			continue
		}
		result[name] = string(bytes)
	}

	return result
}

func DeserializeContext(ctx context.Context, data SerializedContext) context.Context {
	for name, reg := range contextMap {
		val, ok := data[name]
		if !ok {
			continue
		}

		ctx = context.WithValue(ctx, reg.Key, reg.Convert([]byte(val)))
	}

	return ctx
}

func SerializeMessage(msg message.Message, t SerialType) ([]byte, error) {
	if t == Gob {
		b := &bytes.Buffer{}
		enc := gob.NewEncoder(b)
		err := enc.Encode(&msg)
		return b.Bytes(), err
	}

	if t == Json {
		_, ok := messageMap[msgKey(msg)]
		if !ok {
			return []byte{}, errors.New("bus.SerializeMessage: cannot serialize as JSON, not registered " + msgKey(msg))
		}

		data, err := json.Marshal(msg)
		if err != nil {
			return []byte{}, err
		}

		modified := strings.Replace(
			string(data),
			"{",
			fmt.Sprintf(`{"__type":"%s",`, msgKey(msg)),
			1,
		)
		if modified == `{,}` {
			modified = "{}"
		}
		return []byte(modified), nil
	}

	return []byte{}, errors.New("bus.Serializer: unknown encoding")
}

func DeserializeMessage(data []byte) (message.Message, error) {
	if json.Valid(data) {
		var reader typeReader
		err := json.Unmarshal(data, &reader)
		if err != nil {
			return nil, err
		}

		msgType, ok := messageMap[reader.Type]
		if !ok {
			return nil, errors.New("bus.DeserializeMessage: unknown type " + reader.Type)
		}

		msg := reflect.New(msgType).Interface()
		err = json.Unmarshal(data, msg)
		return reflect.ValueOf(msg).Elem().Interface().(message.Message), err
	}

	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)
	var result message.Message
	err := dec.Decode(&result)
	return result, err
}
