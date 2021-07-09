package users

import (
    "github.com/GabrielCarpr/cqrs/bus"
)

type Users struct {}

func (u Users) Services() []bus.Def {
    return []bus.Def{}
}

func (u Users) Commands(b bus.CmdBuilder) {}

func (u Users) Queries(b bus.QueryBuilder) {}

func (u Users) EventRules() bus.EventRules {
    return bus.EventRules{}
}
