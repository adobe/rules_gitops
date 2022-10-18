package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/adobe/rules_gitops/testing/it_sidecar/client"
)

var setup client.K8STestSetup

func TestMain(m *testing.M) {
	setup = client.K8STestSetup{PortForwardServices: map[string]int{"helloworld": 8080}}
	setup.TestMain(m)
}

func TestSuccessfulReceival(t *testing.T) {
	localPort := setup.GetServiceLocalPort("helloworld")
	fmt.Printf("helloworld server is available at localport %d\n", localPort)
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d", localPort))
	if err != nil {
		t.Fatalf("request error %s", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("Unexpected status code %d, expectted 200", resp.StatusCode)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Hello World") {
		t.Error("Unexpected content returned:", string(body))
	}
}
