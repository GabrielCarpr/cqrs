package bus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type tc struct {
	CommandType
}

func (tc) Valid() error {
	return nil
}

func (tc) Command() string {
	return "tc"
}

type te struct {
	EventType
}

func (te) Event() string {
	return "te"
}

type te2 struct {
	EventType
}

func (te2) Event() string {
	return "te2"
}

func TestEventRulesMerges(t *testing.T) {
	rules := eventRules{}
	rules = rules.Merge(EventRules{
		&te{}:  []string{"test1", "test2"},
		&te2{}: []string{"hi there"},
	})

	assert.Len(t, rules, 2)

	rules = rules.Merge(EventRules{
		&te{}:  []string{"nah bye"},
		&te2{}: []string{"ahh u scary"},
	})

	assert.Len(t, rules, 2)
	assert.Len(t, rules[te{}.Event()], 3)
	assert.Contains(t, rules[te{}.Event()], "nah bye")
	assert.Contains(t, rules[te2{}.Event()], "hi there")
}

func TestDedupesEvents(t *testing.T) {
	rules := eventRules{}
	rules = rules.Merge(EventRules{
		&te{}: []string{"test", "hi"},
	})

	rules = rules.Merge(EventRules{
		&te{}: []string{"bye", "test"},
	})

	assert.Len(t, rules, 1)
	assert.Len(t, rules[te{}.Event()], 3)
}

type tq struct {
	QueryType
}

func (tq) Query() string {
	return "tq"
}

func (tq) Valid() error {
	return nil
}

type tq2 struct {
	QueryType
}

func (tq2) Query() string {
	return "tq2"
}

func (tq2) Valid() error {
	return nil
}

func TestQueryRules(t *testing.T) {
	rules := queryRules{}
	rules = rules.Merge(QueryRules{
		tq{}: "handler",
	})

	rules = rules.Merge(QueryRules{
		tq2{}: "anotherhandler",
	})

	assert.Len(t, rules, 2)
}

type testCmd struct {
	CommandType

	Return string
}

func (testCmd) Valid() error {
	return nil
}

func (testCmd) Command() string {
	return "testcmd"
}

func TestRoutesCommands(t *testing.T) {
	router := NewMessageRouter()
	router.Extend(CommandRules{
		testCmd{}: "hello",
	})

	handlers := router.Route(testCmd{})
	assert.Len(t, handlers, 1)
	assert.Equal(t, "hello", handlers[0])
}

type TestEvent struct {
	EventType
}

func (TestEvent) Event() string {
	return "test-event"
}

func TestRoutesEvents(t *testing.T) {
	router := NewMessageRouter()
	router.Extend(EventRules{
		&TestEvent{}: []string{"hello", "bye"},
	})

	handlers := router.Route(&TestEvent{})
	assert.Len(t, handlers, 2)
	assert.Contains(t, handlers, "bye")
}

type testQuery struct {
	QueryType
}

func (testQuery) Query() string {
	return "testQuery"
}

func (testQuery) Valid() error {
	return nil
}

func TestRoutesQueries(t *testing.T) {
	router := NewMessageRouter()
	router.Extend(QueryRules{
		testQuery{}: "hello",
	})

	handlers := router.Route(testQuery{})
	assert.Len(t, handlers, 1)
	assert.Equal(t, "hello", handlers[0])
}
