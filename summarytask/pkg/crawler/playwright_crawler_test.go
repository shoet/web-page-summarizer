package crawler_test

import (
	"testing"

	"github.com/shoet/web-page-summarizer-task/pkg/crawler"
)

func Test_PlaywrightClient_FetchPage(t *testing.T) {

	config := &crawler.PlaywrightClientConfig{
		SkipInstallBrowsers:     false,
		BrowserLaunchTimeoutSec: 30,
	}
	playwrightClient, closer, err := crawler.NewPlaywrightClient(config)
	if err != nil {
		t.Fatalf("could not create playwright client: %v", err)
	}
	t.Cleanup(func() {
		if err := closer(); err != nil {
			t.Fatalf("could not close playwright client: %v", err)
		}
	})

	_, err = playwrightClient.FetchPage("https://example.com")
	if err != nil {
		t.Fatalf("could not fetch page: %v", err)
	}

}
