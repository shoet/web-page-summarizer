package scraper

import (
	"fmt"
	"strings"

	"github.com/playwright-community/playwright-go"
)

type NormalScraper struct{}

func (s *NormalScraper) Scrape(page playwright.Page) (title, body string, err error) {
	titles, err := page.Locator("h1").All()
	if err != nil {
		return "", "", fmt.Errorf("could not get h1: %v", err)
	}
	titleBuilder := strings.Builder{}
	for _, t := range titles {
		text, err := t.TextContent()
		if err != nil {
			return "", "", fmt.Errorf("could not get text content: %v", err)
		}
		if _, err := titleBuilder.WriteString(text); err != nil {
			return "", "", fmt.Errorf("could not write string: %v", err)
		}
	}

	contents, err := page.Locator("p").All()
	if err != nil {
		return "", "", fmt.Errorf("could not get p: %v", err)
	}
	contentBuilder := strings.Builder{}
	for _, c := range contents {
		text, err := c.TextContent()
		if err != nil {
			return "", "", fmt.Errorf("could not get text content: %v", err)
		}
		if _, err := contentBuilder.WriteString(text); err != nil {
			return "", "", fmt.Errorf("could not write string: %v", err)
		}
	}
	return titleBuilder.String(), contentBuilder.String(), nil
}
