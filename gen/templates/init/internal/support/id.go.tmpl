package support

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

func MustParseID(s string) ID {
	return ID{uuid.MustParse(s)}
}

func ParseID(s string) (ID, error) {
	id, err := uuid.Parse(s)
	return ID{id}, err
}

func NewID() ID {
	return ID{uuid.New()}
}

type ID struct {
	uuid.UUID `cqrs:",squash"`
}

func (u *ID) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case string:
		id, err := ParseID(input)
		*u = id
		return err
	default:
		return fmt.Errorf("wtf")
	}
}

func (u ID) ImplementsGraphQLType(name string) bool {
	return name == "ID"
}

func (i ID) Nil() bool {
	return i.UUID == uuid.Nil
}

func (i ID) UID() uuid.UUID {
	return i.UUID
}

func (i ID) Value() (driver.Value, error) {
	return i.UUID.Value()
}

func (i *ID) UnmarshalJSON(data []byte) error {
	id, err := ParseID(string(data))
	if err != nil {
		return err
	}
	*i = id
	return nil
}
