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

var contextMap map[fmt.Stringer]reflect.Type
var messageMap map[string]reflect.Type

func init() {
	contextMap = make(map[fmt.Stringer]reflect.Type)
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

func RegisterContextKey(key fmt.Stringer, val interface{}) {
	contextMap[key] = reflect.TypeOf(val)
}

func SerializeContext(ctx context.Context) SerializedContext {
	result := make(SerializedContext)

	for name := range contextMap {
		val := ctx.Value(name)
		if val == nil {
			continue
		}
		str, ok := val.(string)
		if ok {
			result[name.String()] = str
			continue
		}
		bytes, err := json.Marshal(val)
		if err != nil {
			log.Printf("Error serialising context value [%s]: %s", name, err)
			continue
		}
		result[name.String()] = string(bytes)
	}

	return result
}

func DeserializeContext(ctx context.Context, data SerializedContext) context.Context {
	for name, t := range contextMap {
		val, ok := data[name.String()]
		if !ok {
			continue
		}

		if !json.Valid([]byte(val)) {
			ctx = context.WithValue(ctx, name, val)
			continue
		}

		target := reflect.New(t).Interface()
		err := json.Unmarshal([]byte(val), target)
		if err != nil {
			log.Printf("bus.DeserializeContext: %v", err)
		}

		ctx = context.WithValue(ctx, name, reflect.ValueOf(target).Elem().Interface())
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
