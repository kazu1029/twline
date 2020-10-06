package get

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

const (
	docElem   = "article > div > div > div > div.css-1dbjc4n.r-18u37iz > div.css-1dbjc4n.r-1iusvr4.r-16y2uox.r-1777fci.r-1mi0q7o > div > div > div > span"
	tweetElem = "section > div > div > div > div > div > article > div > div"
)

func GetTimeline(urls []string) {
	ctx := context.Background()
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.WindowSize(1200, 3000),
	)
	allocCtx, _ := chromedp.NewExecAllocator(ctx, opts...)

	var wg sync.WaitGroup
	wg.Add(len(urls))
	for _, url := range urls {
		go func(u string) {
			var html string
			defer wg.Done()
			cc, cancel := chromedp.NewContext(allocCtx)
			defer cancel()
			err := chromedp.Run(cc, scrape(u, &html))
			if err != nil {
				log.Fatal(err)
			}

			readFromHTML(html)
		}(url)
	}

	wg.Wait()
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
		text := s.Find(docElem).Text()
		fmt.Println(text)
	})
}

func scrape(url string, str *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitVisible(tweetElem),
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
