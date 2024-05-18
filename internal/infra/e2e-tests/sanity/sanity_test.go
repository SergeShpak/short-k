//go:build e2e_tests
// +build e2e_tests

package sanity

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"

	"shortik/internal/infra/e2e-tests/client"
)

var cfg config

func TestMain(m *testing.M) {
	var err error
	cfg, err = getConfig()
	if err != nil {
		log.Printf("failed to get service configuration: %v", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestSanity(t *testing.T) {
	rawClient, err := client.NewClient(cfg.ServiceAddr)
	if err != nil {
		t.Errorf("failed to initialize a new raw client: %v", err)
		return
	}
	rawClient.Client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	c, err := client.NewClientWithResponses(cfg.ServiceAddr)
	if err != nil {
		t.Errorf("failed to initialize a new client: %v", err)
		return
	}
	c.ClientInterface = rawClient

	ctx := context.Background()

	getNonExistentSlugResp, err := c.GetSlugWithResponse(ctx, "nonExistentSlug")
	if err != nil {
		t.Errorf("get non existent slug failed: %v", err)
		return
	}

	if getNonExistentSlugResp.StatusCode() != http.StatusNotFound {
		t.Errorf(
			"expected to receive the status code %d when requesting a non-existen slug, got %d",
			http.StatusNotFound,
			getNonExistentSlugResp.StatusCode(),
		)
		return
	}

	var urlToShorten string = "http://example.com"
	postURLResp, err := c.PostWithResponse(ctx, client.PostJSONRequestBody{
		Url: &urlToShorten,
	})
	if err != nil {
		t.Errorf("post URL failed: %v", err)
		return
	}

	if postURLResp.StatusCode() != http.StatusCreated {
		t.Errorf(
			"expected to receive the status code %d when shortening a URL, got %d",
			http.StatusCreated,
			postURLResp.StatusCode(),
		)
		return
	}

	shortenedURL := *postURLResp.JSON201.ShortenedUrl

	getReq, err := http.NewRequest("GET", shortenedURL, nil)
	if err != nil {
		t.Errorf("failed to create a get URL request: %v", err)
		return
	}

	resp, err := rawClient.Client.Do(getReq)
	if err != nil {
		t.Errorf("failed to get the full URL: %v", err)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf(
			"expected to receive the status code %d when getting a full URL, got %d",
			http.StatusTemporaryRedirect,
			resp.StatusCode,
		)
		return
	}
	locationHeader := resp.Header.Get("Location")
	if locationHeader != urlToShorten {
		t.Errorf("expected location header to be %q, got %q", urlToShorten, locationHeader)
		return
	}
}
