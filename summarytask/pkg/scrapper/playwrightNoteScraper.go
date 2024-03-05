package scraper

import "github.com/playwright-community/playwright-go"

type NoteScraper struct{}

func (s *NoteScraper) Scrape(page playwright.Page) (title, body string, err error) {
	return "", "", nil
}
