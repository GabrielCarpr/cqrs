package rest_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/ports/rest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type TestCmd struct {
	bus.CommandType

	TestVal   string
	TestNum   int
	TestThang string `cqrs:"test_thang"`
	TestList  []string
}

func TestBind(t *testing.T) {
	suite.Run(t, new(BindTest))
}

type BindTest struct {
	suite.Suite
}

func (s BindTest) TestBindsCommandToJson() {
	cmd := TestCmd{}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	body, _ := json.Marshal(map[string]interface{}{
		"TestVal":    "test",
		"testnum":    13,
		"test_thang": "lol",
		"testlist":   []string{"hello", "there"},
	})
	c.Request = httptest.NewRequest("POST", "/v3/test", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	err := rest.Bind(c, &cmd)
	s.NoError(err)
	s.Equal("test", cmd.TestVal)
	s.Equal(13, cmd.TestNum)
	s.Equal("lol", cmd.TestThang)
	s.Equal([]string{"hello", "there"}, cmd.TestList)
}

func (s BindTest) TestBindsCommandToQuery() {
	cmd := TestCmd{}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(
		"GET",
		"/v3/test?testval=test&testnum=13&test_thang=lol&testlist=hello&testlist=there",
		nil,
	)

	err := rest.Bind(c, &cmd)
	s.NoError(err)
	s.Equal("test", cmd.TestVal)
	s.Equal(13, cmd.TestNum)
	s.Equal("lol", cmd.TestThang)
	s.Equal([]string{"hello", "there"}, cmd.TestList)
}

func (s BindTest) TestBindsCommandToForm() {
	cmd := TestCmd{}

	form := url.Values{}
	form.Add("testval", "test")
	form.Add("testnum", "13")
	form.Add("test_thang", "lol")
	form.Add("testlist", "hello")
	form.Add("testlist", "there")

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(
		"POST",
		"/v3/test",
		bytes.NewBuffer([]byte(form.Encode())),
	)
	c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	err := rest.Bind(c, &cmd)
	s.NoError(err)
	s.Equal("test", cmd.TestVal)
	s.Equal(13, cmd.TestNum)
	s.Equal("lol", cmd.TestThang)
	s.Len(cmd.TestList, 2)
	s.Equal("hello", cmd.TestList[0])
}

func (s BindTest) TestBindsCommandToURI() {
	cmd := TestCmd{}
	run := false

	resp := httptest.NewRecorder()
	c, eng := gin.CreateTestContext(resp)
	eng.POST("/v3/:testval/:testnum/:test_thang", func(c *gin.Context) {
		err := rest.Bind(c, &cmd)
		s.NoError(err)
		s.Equal("test", cmd.TestVal)
		s.Equal(13, cmd.TestNum)
		s.Equal("lol", cmd.TestThang)
		run = true
	})
	c.Request = httptest.NewRequest(
		"POST",
		"/v3/test/13/lol",
		nil,
	)

	eng.ServeHTTP(resp, c.Request)
	s.True(run)
}

func (s BindTest) TestAllForRest() {
	cmd := TestCmd{}
	run := false

	resp := httptest.NewRecorder()
	c, eng := gin.CreateTestContext(resp)
	eng.POST("/v3/:testval/:testnum/:test_thang", func(c *gin.Context) {
		err := rest.Bind(c, &cmd)
		s.NoError(err)
		s.Equal("test", cmd.TestVal)
		s.Equal(13, cmd.TestNum)
		s.Equal("hi", cmd.TestThang)
		run = true
	})
	body := map[string]interface{}{
		"testlist":   []string{"Hello", "world"},
		"test_thang": "lol",
	}
	b, _ := json.Marshal(body)
	c.Request = httptest.NewRequest(
		"POST",
		"/v3/test/13/hi",
		bytes.NewBuffer(b),
	)
	c.Request.Header.Add("Content-Type", "application/json")

	eng.ServeHTTP(resp, c.Request)
	s.True(run)
}
