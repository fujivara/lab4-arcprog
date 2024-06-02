package integration

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	_ "net/http"
	"os"
	"testing"
	"time"
	_ "time"
)

var client = http.Client{
	Timeout: 3 * time.Second,
}

func TestLoadBalancerAlgorithm(t *testing.T) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		t.Skip("Integration test is not enabled")
	}

	loadBalancerAddress := "http://balancer:8090"

	numRequests := 10
	serverResponsesCount := make(map[string]int)

	for i := 0; i < numRequests; i++ {
		resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", loadBalancerAddress))
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()

		serverResponsesCount[resp.Header.Get("lb-from")]++
	}

	assert.Greater(t, len(serverResponsesCount), 1, "Responses should come from more than one server")
	for server, count := range serverResponsesCount {
		t.Logf("Server %s handled %d requests", server, count)
	}
}

func BenchmarkBalancer(b *testing.B) {
	b.ResetTimer()
	loadBalancerAddress := "http://balancer:8090/api/v1/some-data"

	req, err := http.NewRequest("GET", loadBalancerAddress, nil)
	if err != nil {
		b.Fatalf("Failed to create request: %v", err)
	}

	b.RunParallel(func(pb *testing.PB) {
		client := http.Client{}
		for pb.Next() {
			resp, err := client.Do(req)
			if err != nil {
				b.Fatalf("Failed to send request to load balancer: %v", err)
			}
			resp.Body.Close()
		}
	})
}
