package crawler

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type PageCrawler struct {
	browser *rod.Browser
}

type PageCrawlerInput struct {
	BrowserPath string
}

func NewPageCrawler(input *PageCrawlerInput) (*PageCrawler, error) {
	if input.BrowserPath == "" {
		return nil, fmt.Errorf("Browser path is empty")
	}
	browser, err := BuildBrowser(input.BrowserPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to build browser: %w", err)
	}
	return &PageCrawler{browser: browser}, nil
}

func (f *PageCrawler) FetchContents(url string) (string, string, error) {
	page, err := f.FetchPage(url)
	if err != nil {
		return "", "", fmt.Errorf("Failed to fetch page: %w", err)
	}
	title, content, err := ScrapBody(page)
	if err != nil {
		return "", "", fmt.Errorf("Failed to scrap body: %w", err)
	}
	return title, content, nil
}

func (f *PageCrawler) FetchPage(url string) (*rod.Page, error) {
	p, err := f.browser.Page(proto.TargetCreateTarget{URL: url})
	if err != nil {
		return nil, fmt.Errorf("Failed to create page: %w", err)
	}
	if err := p.WaitLoad(); err != nil {
		return nil, fmt.Errorf("Failed to wait load: %w", err)
	}
	return p, nil
}

func BuildBrowser(browserPath string) (*rod.Browser, error) {
	u := launcher.New().Bin(browserPath).NoSandbox(true).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()
	fmt.Printf("Start browser: %s\n", u)
	return browser, nil
}

func ScrapBody(page *rod.Page) (string, string, error) {
	titles, err := page.Elements("h1")
	if err != nil {
		return "", "", fmt.Errorf("failed to get h1: %w", err)
	}

	titleBuilder := strings.Builder{}
	for _, t := range titles {
		titleBuilder.WriteString(t.MustText())
	}

	paragraphs, err := page.Elements("p")
	if err != nil {
		return "", "", fmt.Errorf("failed to get p: %w", err)
	}

	contentBuilder := strings.Builder{}
	for _, p := range paragraphs {
		contentBuilder.WriteString(p.MustText() + "\n")
	}

	return titleBuilder.String(), contentBuilder.String(), nil
}
