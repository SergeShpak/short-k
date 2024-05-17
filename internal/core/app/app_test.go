package app_test

import (
	"context"
	"testing"

	"shortik/internal/core/app"
	"shortik/internal/core/model"
)

func TestAppHappyPath(t *testing.T) {
	a := app.NewApp()

	const expectedFullURL = "http://example.com"
	ctx := context.Background()
	shortenURLResp, err := a.ShortenURL(ctx, model.ShortenURLRequest{
		URL: expectedFullURL,
	})
	if err != nil {
		t.Fatalf("failed to shorten URL: %v", err)
	}

	resp, err := a.GetFullURL(ctx, model.GetFullURLRequest{
		ShortURL: shortenURLResp.ShortURL,
	})
	if err != nil {
		t.Fatalf("failed to get full URL: %v", err)
	}

	if resp.URL != expectedFullURL {
		t.Fatalf("expected to get \"%s\", actually got \"%s\"", expectedFullURL, resp.URL)
	}
}
