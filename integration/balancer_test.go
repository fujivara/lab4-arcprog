package integration

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	_ "time"

	"github.com/jarcoal/httpmock"
	"github.com/roman-mazur/architecture-practice-4-template/cmd/lb"
	"github.com/stretchr/testify/assert"
)

func TestLoadBalancingDistribution(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	balancer.ServersPool = map[string]int{
		"server1:8080": 0,
		"server2:8080": 0,
		"server3:8080": 0,
	}

	httpmock.RegisterResponder("GET", "http://server1:8080/test",
		httpmock.NewStringResponder(200, "Server1 response"))
	httpmock.RegisterResponder("GET", "http://server2:8080/test",
		httpmock.NewStringResponder(200, "Server2 response"))
	httpmock.RegisterResponder("GET", "http://server3:8080/test",
		httpmock.NewStringResponder(200, "Server3 response"))

	req := httptest.NewRequest("GET", "/test", nil)
	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		serverName := balancer.GetServer()
		err := balancer.Forward(serverName, rw, r)
		if err != nil {
			return
		}
	})

	numRequests := 10
	serverResponses := make(map[string]int)

	for i := 0; i < numRequests; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		serverResponses[string(body)]++
	}

	assert.Greater(t, len(serverResponses), 1, "Responses should come from more than one server")
	for server, count := range serverResponses {
		t.Logf("Server %s handled %d requests", server, count)
	}
}

func BenchmarkBalancer(b *testing.B) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	balancer.ServersPool = map[string]int{
		"server1:8080": 0,
		"server2:8080": 0,
		"server3:8080": 0,
	}

	httpmock.RegisterResponder("GET", "http://server1:8080/test",
		httpmock.NewStringResponder(200, "Server1 response"))
	httpmock.RegisterResponder("GET", "http://server2:8080/test",
		httpmock.NewStringResponder(200, "Server2 response"))
	httpmock.RegisterResponder("GET", "http://server3:8080/test",
		httpmock.NewStringResponder(200, "Server3 response"))

	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		serverName := balancer.GetServer()
		err := balancer.Forward(serverName, rw, r)
		if err != nil {
			return
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		_, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}
