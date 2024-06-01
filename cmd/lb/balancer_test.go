package balancer

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	_ "time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestGetServer(t *testing.T) {
	ServersPool = map[string]int{
		"server1:8080": 100,
		"server2:8080": 50,
		"server3:8080": 75,
	}

	expectedServer := "server2:8080"
	actualServer := GetServer()
	assert.Equal(t, expectedServer, actualServer, "getServer should return the server with the least traffic")
}

func TestHealth(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://server1:8080/health",
		httpmock.NewStringResponder(200, "OK"))

	httpmock.RegisterResponder("GET", "http://server2:8080/health",
		httpmock.NewStringResponder(500, "Internal Server Error"))

	assert.True(t, health("server1:8080"), "server1 should be healthy")
	assert.False(t, health("server2:8080"), "server2 should be unhealthy")
}

func TestForward(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mockResponse := httpmock.NewStringResponder(200, "Mocked server response")
	httpmock.RegisterResponder("GET", "http://server1:8080/test", mockResponse)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	err := Forward("server1:8080", w, req)
	assert.NoError(t, err, "forward should not return an error")

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "Mocked server response", string(body), "forward should return the correct response body")
	assert.Equal(t, 200, resp.StatusCode, "forward should return the correct status code")
}

func TestMainHandler(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	ServersPool = map[string]int{
		"server1:8080": 100,
		"server2:8080": 50,
		"server3:8080": 75,
	}

	httpmock.RegisterResponder("GET", "http://server2:8080/test",
		httpmock.NewStringResponder(200, "Server2 response"))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		serverName := GetServer()
		err := Forward(serverName, rw, r)
		if err != nil {
			return
		}
	})

	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "Server2 response", string(body), "Main handler should forward to the correct server")
	assert.Equal(t, 200, resp.StatusCode, "Main handler should return the correct status code")
}
