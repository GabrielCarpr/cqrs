package rest

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

// MustBind calls binds and aborts the request if an error is raised
func MustBind(c *gin.Context, target interface{}) error {
	if err := Bind(c, target); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return err
	}
	return nil
}

// Bind maps a request onto a command/query/event
func Bind(c *gin.Context, target interface{}) error {
	if reflect.ValueOf(target).Kind() != reflect.Ptr {
		return errors.New("Bind target must be a pointer")
	}

	intermediary := make(map[string]interface{})
	err := bindJSON(c, &intermediary)
	if err != nil {
		return err
	}
	err = bindQuery(c, &intermediary)
	if err != nil {
		return err
	}
	err = bindForm(c, &intermediary)
	if err != nil {
		return err
	}
	err = bindURI(c, &intermediary)
	if err != nil {
		return err
	}

	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ZeroFields:       false,
		WeaklyTypedInput: true,
		Result:           target,
		TagName:          "cqrs",
		Squash:           true,
		DecodeHook:       unmarshalDecodeHook,
	})
	if err != nil {
		return err
	}
	err = d.Decode(intermediary)
	if err != nil {
		return err
	}

	return nil
}

func unmarshalDecodeHook(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
	if to.Kind() != reflect.Struct {
		return data, nil
	}
	target, ok := reflect.New(to).Interface().(Binder)
	if !ok {
		return data, nil
	}

	target.Bind(data)
	return target, nil
}

// Binder should be implemented on structs with embedded scalar
// to aid binding, such a value objects
type Binder interface {
	Bind(interface{}) error
}

func bindJSON(c *gin.Context, target *map[string]interface{}) error {
	if c.ContentType() != "application/json" {
		return nil
	}
	err := c.BindJSON(target)
	return err
}

func bindQuery(c *gin.Context, target *map[string]interface{}) error {
	query := c.Request.URL.Query()
	return bindURLValues(query, target)
}

func bindForm(c *gin.Context, target *map[string]interface{}) error {
	err := c.Request.ParseForm()
	if err != nil {
		return err
	}
	return bindURLValues(c.Request.PostForm, target)
}

func bindURLValues(vals url.Values, target *map[string]interface{}) error {
	for key, val := range vals {
		if len(val) == 0 {
			continue
		} else if len(val) > 1 {
			(*target)[key] = val
		} else {
			(*target)[key] = val[0]
		}
	}
	return nil
}

func bindURI(c *gin.Context, target *map[string]interface{}) error {
	params := c.Params
	for _, param := range params {
		(*target)[param.Key] = param.Value
	}
	return nil
}
