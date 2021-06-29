package context

import (
	"context"
	"encoding/json"
	"log"
)

func NewContextSerializer() *ContextSerializer {
	return &ContextSerializer{make(map[string]keyRegistration)}
}

type keyRegistration struct {
	Key     interface{ String() string }
	Convert func(json []byte) interface{}
}

// ContextSerializer serializes contexts based on a register of keys
type ContextSerializer struct {
	register map[string]keyRegistration
}

// Register assigns a new key to a name
func (cs *ContextSerializer) Register(key interface{ String() string }, fn func(json []byte) interface{}) {
	cs.register[key.String()] = keyRegistration{key, fn}
}

// Serialize contexts a Ctx to a string -> string map
func (cs ContextSerializer) Serialize(ctx context.Context) map[string]string {
	result := make(map[string]string)

	for name, reg := range cs.register {
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

func (cs ContextSerializer) Deserialize(ctx context.Context, data map[string]string) context.Context {
	for name, reg := range cs.register {
		val, ok := data[name]
		if !ok {
			continue
		}

		ctx = context.WithValue(ctx, reg.Key, reg.Convert([]byte(val)))
	}

	return ctx
}
