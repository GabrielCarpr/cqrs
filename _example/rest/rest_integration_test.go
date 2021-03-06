package rest_test

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"example/internal/tester"
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/stretchr/testify/suite"
	adapter "github.com/GabrielCarpr/cqrs/ports/rest"
	"example/rest"
)

func TestRestAdapter(t *testing.T) {
	suite.Run(t, new(RestTest))
}

type RestTest struct {
	suite.Suite
	tester.Integration

	server *adapter.Server
}
func (s *RestTest) SetupTest() {
	s.Integration.SetupTest()

	s.server = rest.Rest(s.Bus(), tester.GetTestConfig())
}

func (s *RestTest) TestLoginViaRest() {
	resp := httptest.NewRecorder()
	body := map[string]interface{}{
		"email": "me@gabrielcarpreau.com",
		"password": "password123",
	}
	json, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/rest/v1/auth/login", bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")

	s.server.Router.ServeHTTP(resp, req)

	s.Equal(200, resp.Code)
}

func (s *RestTest) loggedIn(nextResp http.ResponseWriter, next *http.Request) {
	resp := httptest.NewRecorder()
	body := map[string]interface{}{
		"email": "me@gabrielcarpreau.com",
		"password": "password123",
	}
	jsonInput, _ := json.Marshal(body)
	logIn := httptest.NewRequest("POST", "/rest/v1/auth/login", bytes.NewBuffer(jsonInput))
	logIn.Header.Set("Content-Type", "application/json")

	s.server.Router.ServeHTTP(resp, logIn)
	s.Require().Equal(200, resp.Code)
	var respJson map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &respJson)
	s.Require().NoError(err)
	token := respJson["access_token"].(string)

	next.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	s.server.Router.ServeHTTP(nextResp, next)
}

func (s *RestTest) TestGetsUsers() {
	resp := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/rest/v1/users/", nil)

	s.loggedIn(resp, req)

	s.Require().Equal(200, resp.Code)

	var body map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &body)
	s.Len(body["data"].([]interface{}), 1)
	s.Equal(float64(1), body["metadata"].(map[string]interface{})["count"].(float64))
}
