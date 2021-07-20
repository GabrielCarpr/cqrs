package entities

import (
	"database/sql/driver"
	"fmt"
	"encoding/json"
	"errors"
)

func NewRoleID(name string) RoleID {
	return RoleID{name}
}

type RoleID struct {
	Name string
}

func (r RoleID) Equals(id RoleID) bool {
	return r.Name == id.Name
}

func (r RoleID) String() string {
	return r.Name
}

func (r RoleID) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Name)
}

func (r RoleID) Value() (driver.Value, error) {
	return r.Name, nil
}

func (r RoleID) ImplementsGraphQLType(name string) bool {
	return name == "ID"
}

func (r *RoleID) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case string:
		id := NewRoleID(input)
		*r = id
		return nil
	default:
		return fmt.Errorf("wtf")
	}
}

func (r *RoleID) Bind(data interface{}) error {
	str, ok := data.(string)
	if !ok {
		return errors.New("RoleID is not string")
	}
	(*r).Name = str
	return nil
}

// Scope is an access that is provided to a user by some mean
type Scope struct {
	Name string
}

func (s Scope) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Name)
}
