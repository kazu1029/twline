package get

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

const (
	usernameElem = "article > div > div > div > div.css-1dbjc4n.r-18u37iz > div.css-1dbjc4n.r-1iusvr4.r-16y2uox.r-1777fci.r-1mi0q7o > div > div > div > div.css-1dbjc4n.r-1d09ksm.r-18u37iz.r-1wbh5a2 > div.css-1dbjc4n.r-1wbh5a2.r-dnmrzs > a > div > div.css-1dbjc4n.r-1awozwy.r-18u37iz.r-dnmrzs"
	timeElem     = "article > div > div > div > div.css-1dbjc4n.r-18u37iz > div.css-1dbjc4n.r-1iusvr4.r-16y2uox.r-1777fci.r-1mi0q7o > div > div > div > div.css-1dbjc4n.r-1d09ksm.r-18u37iz.r-1wbh5a2 > a > time"
	rtsElem      = "article > div > div > div > div.css-1dbjc4n.r-18u37iz > div.css-1dbjc4n.r-1iusvr4.r-16y2uox.r-1777fci.r-1mi0q7o > div > div.css-1dbjc4n.r-18u37iz.r-1wtj0ep.r-156q2ks.r-1mdbhws > div > div"
	favsElem     = "article > div > div > div > div.css-1dbjc4n.r-18u37iz > div.css-1dbjc4n.r-1iusvr4.r-16y2uox.r-1777fci.r-1mi0q7o > div > div.css-1dbjc4n.r-18u37iz.r-1wtj0ep.r-156q2ks.r-1mdbhws > div > div"
	docElem      = "article > div > div > div > div.css-1dbjc4n.r-18u37iz > div.css-1dbjc4n.r-1iusvr4.r-16y2uox.r-1777fci.r-1mi0q7o > div > div > div > span"
	tweetElem    = "section > div > div > div > div > div > article > div > div"
	baseURL      = "https://twitter.com/"
)

type tweet struct {
	ID        int      `json:"id"`
	Body      string   `json:"body"`
	User      string   `json:"user"`
	Images    []string `json:"images"`
	Links     []string `json:"links"`
	Rts       string   `json:"rts"`
	Favs      string   `json:"favs"`
	TweetedAt string   `json:"tweeted_at"`
}

// Timeline fetchs tweets from assigned targets arg.
// output arg is to specify output source.
func Timeline(targets []string, output string) {
	ctx := context.Background()
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.WindowSize(1200, 3000),
	)
	allocCtx, _ := chromedp.NewExecAllocator(ctx, opts...)

	var wg sync.WaitGroup
	wg.Add(len(targets))
	for _, target := range targets {
		go func(target string) {
			var html string
			defer wg.Done()
			cc, cancel := chromedp.NewContext(allocCtx)
			defer cancel()
			err := chromedp.Run(cc, scrape(target, &html))
			if err != nil {
				log.Fatal(err)
			}

			tweets := readFromHTML(html)
			switch output {
			case "csv":
				outputCSV(target, tweets)
			default:
				fmt.Println(tweets)
			}
		}(target)
	}

	wg.Wait()
}

func outputCSV(filename string, tweets []tweet) {
	nowStr := strconv.FormatInt(time.Now().Unix(), 10)
	file, err := os.OpenFile("tmp/"+filename+"-"+nowStr+".csv", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	rows := genRows(tweets)
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			log.Fatal(err)
		}
	}
	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
}

func genRows(src interface{}) [][]string {
	sl := toSlice(src)
	rows := make([][]string, 1)

	for n, d := range sl {
		rows = append(rows, []string{})
		v := reflect.ValueOf(d)
		for i := 0; i < v.NumField(); i++ {
			if n == 0 {
				colName := strings.ToLower(v.Type().Field(i).Name)
				rows[0] = append(rows[0], colName)
			}
			rows[len(rows)-1] = append(rows[len(rows)-1], fmt.Sprint(v.Field(i).Interface()))
		}
	}
	return rows
}

func toSlice(src interface{}) []interface{} {
	ret := []interface{}{}
	if v := reflect.ValueOf(src); v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			ret = append(ret, v.Index(i).Interface())
		}
	} else {
		ret = append(ret, v.Interface())
	}

	return ret
}

func readFromHTML(html string) []tweet {
	reader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Fatal(err)
	}

	var tweets []tweet
	tSelections := doc.Find(tweetElem)
	tSelections.Each(func(i int, s *goquery.Selection) {
		user := s.Find(usernameElem).Text()
		timeStr, _ := s.Find(timeElem).Attr("datetime")
		text := s.Find(docElem).Text()
		var images []string
		s.Find("img").Each(func(j int, t *goquery.Selection) {
			src, _ := t.Attr("src")
			images = append(images, src)
		})
		var links []string
		s.Find("a").Each(func(j int, t *goquery.Selection) {
			title, _ := t.Attr("title")
			if strings.HasPrefix(title, "https://") {
				links = append(links, title)
			}
		})

		rtsStr, _ := s.Find(rtsElem).Attr("aria-label")
		fmt.Printf("rtsStr: %v\n", rtsStr)
		rtsAry := strings.Split(rtsStr, " ")
		favsStr, _ := s.Find(favsElem).Attr("aria-label")
		fmt.Printf("favsStr: %v\n", favsStr)
		favsAry := strings.Split(favsStr, " ")
		tweet := tweet{
			ID:        i,
			Body:      text,
			User:      user,
			Images:    images,
			Links:     links,
			Rts:       rtsAry[0],
			Favs:      favsAry[0],
			TweetedAt: timeStr,
		}
		tweets = append(tweets, tweet)
	})

	return tweets
}

func scrape(url string, str *string) chromedp.Tasks {
	u := baseURL + url
	return chromedp.Tasks{
		chromedp.Navigate(u),
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
