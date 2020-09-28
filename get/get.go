package get

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

const (
	docElem   = "article > div > div > div > div.css-1dbjc4n.r-18u37iz > div.css-1dbjc4n.r-1iusvr4.r-16y2uox.r-1777fci.r-1mi0q7o > div > div > div > span"
	tweetElem = "section > div > div > div > div > div > article > div > div"
	waitElem  = "article"
)

func GetTimeline(urls []string) {
	ctx := context.Background()
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.WindowSize(1200, 3000),
	)
	allocCtx, _ := chromedp.NewExecAllocator(ctx, opts...)
	cc, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	for _, url := range urls {
		var html string
		err := chromedp.Run(cc, scrape(url, &html))
		if err != nil {
			log.Fatal(err)
		}
		readFromHTML(html)
	}
}

func readFromHTML(html string) {
	reader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Fatal(err)
	}

	tSelections := doc.Find(tweetElem)
	tSelections.Each(func(i int, s *goquery.Selection) {
		// docHtml, err := s.Html()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// fmt.Println("docHtml: ", docHtml)
		fmt.Println("doc text: ", s.Find(docElem).Text())
	})
}

func scrape(url string, str *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitVisible(waitElem),
		chromedp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			*str, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			return err
		}),
	}
}
