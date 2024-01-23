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
	case strings.Contains(url, "https://yahoo.com"):
		return &NoteScraper{}, nil
	case strings.Contains(url, "https://note.com"):
		return &NoteScraper{}, nil
	default:
		return &NormalScraper{}, nil
	}
}
