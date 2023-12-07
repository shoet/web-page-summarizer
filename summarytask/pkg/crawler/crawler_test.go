package crawler

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/shoet/webpagesummary/testutil"
)

func Test_FetchPage(t *testing.T) {
	server := testutil.BuildLocalServer()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		if err := server.Serve(l); err != http.ErrServerClosed {
			t.Fatalf("HTTP server ListenAndServe: %v", err)
		}
		close(idleConnsClosed)
	}()

	url := "http://" + l.Addr().String()

	sut, err := NewPageCrawler(&PageCrawlerInput{
		BrowserPath: "/opt/homebrew/bin/chromium", // TODO local
	})
	if err != nil {
		t.Fatalf("failed to create PageCrawler: %v", err)
	}

	title, content, err := sut.FetchContents(url)
	if err != nil {
		t.Fatalf("failed to fetch contents: %v", err)
	}

	if title != "TestPage h1" {
		t.Fatalf("title is not expected: %v", title)
	}
	if strings.Trim(content, "\n") != "TestPage p" {
		t.Fatalf("content is not expected: %v", content)
	}

	if err := server.Shutdown(context.Background()); err != nil {
		t.Fatalf("HTTP server Shutdown: %v", err)
	}

	<-idleConnsClosed
}
