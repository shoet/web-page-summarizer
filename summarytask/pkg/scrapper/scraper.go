package scraper

import (
	"strings"

	"github.com/playwright-community/playwright-go"
)

type PlaywrightScraper interface {
	Scrape(page playwright.Page) (title, body string, err error)
}

func NewPlaywrightScraper(url string) (PlaywrightScraper, error) {
	switch {
	case strings.HasPrefix(url, "https://yahoo.com"):
		return &NormalScraper{}, nil // TODO
	case strings.HasPrefix(url, "https://note.com"):
		return &NormalScraper{}, nil // TODO
	default:
		return &NormalScraper{}, nil
	}
}
