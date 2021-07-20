package support

import (
	"example/internal/errs"
	"example/pkg/util"
	"log"
	"regexp"

	"github.com/mitchellh/mapstructure"
)

const (
	ASC  Order = "ASC"
	DESC Order = "DESC"
)

func NewPaginatedQuery(d interface{}, c int) PaginatedQuery {
	return PaginatedQuery{Data: d, Metadata: struct{ Count int32 `json:"count"` }{int32(c)}}
}

type PaginatedQuery struct {
	Data     interface{} `json:"data"`
	Metadata struct {
		Count int32 `json:"count"`
	} `json:"metadata"`
}

type Order string

type Paging struct {
	Page    *int
	PerPage *int
	SortBy  *string
	Order   *Order
}

func (Paging) ImplementsGraphQLType(name string) bool {
	return name == "Pagination"
}

func (Paging) Nullable() {}

func (p *Paging) UnmarshalGraphQL(input interface{}) error {
	return mapstructure.Decode(input, p)
}

func (q Paging) Valid() error {
	switch true {
	case !util.Empty(q.Page) && *q.Page < 1:
		return errs.ValidationError("Cannot request page less than 1")
	case !util.Empty(q.PerPage) && *q.PerPage > 250:
		return errs.ValidationError("Cannot request more than 250 results")
	}
	return nil
}

func (q Paging) OrderBy() string {
	if util.Empty(q.Order) {
		return string(ASC)
	}
	return string(*q.Order)
}

func (q Paging) OrderField() string {
	if util.Empty(q.SortBy) {
		return "id"
	}

	re := regexp.MustCompile(`^[A-Za-z0-9_]+$`)
	if re.MatchString(*q.SortBy) {
		return *q.SortBy
	}
	log.Panicf("Stopped SQL injection: %s", *q.SortBy)
	return ""
}

func (q Paging) Limit() int {
	if util.Empty(q.PerPage) {
		return 100
	}

	return *q.PerPage
}

func (q Paging) Offset() int {
	if util.Empty(q.Page) {
		return 0
	}

	return (*q.Page - 1) * q.Limit()
}
