package scraper

import "github.com/playwright-community/playwright-go"

type PlaywrightYahooNewsScraper struct {
}

func (s *PlaywrightYahooNewsScraper) Scrape(page playwright.Page) (title, body string, err error) {

	return "", "", nil
}
