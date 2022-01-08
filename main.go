package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	//"github.com/actions-go/toolkit/core"

	"github.com/gocolly/colly"
)

const (
	MAX_LIMIT int = 56
)

type Author_ struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ResharedStatus_ struct {
	Author Author_ `json:"Author"`
}

type Card_ struct {
	ID             string          `json:"id"`
	URL            string          `json:"url"`
	Title          string          `json:"title"`
	Activity       string          `json:"activity"`
	OwnerURI       string          `json:"owner_uri"`
	OwnerName      string          `json:"owner_name"`
}

type Status_ struct {
	ID             string          `json:"id"`
	Card           Card_           `json:"card,omitempty"`
	Text           string          `json:"text,omitempty"`
	Activity       string          `json:"activity"`
	SharingURL     string          `json:"sharing_url"`
	ResharedStatus ResharedStatus_ `json:"reshared_status,omitempty"`
}

type Item struct {
	Status Status_ `json:"status"`
}

type Result struct {
	Count int    `json:"count"`
	Items []Item `json:"items"`
}

func WriteToFile(path string, items []string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		log.Fatal(err)
	}

	f.Close()

	f, err = os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	stop := false
	w := bufio.NewWriter(f)
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if !stop && line == "<!-- DOUBAN-ACTIVITIES:START -->" {
			fmt.Fprintln(w, line)
			stop = true
		}

		if line == "<!-- DOUBAN-ACTIVITIES:END -->" {
			stop = false
			for _, item := range items {
				fmt.Fprintln(w, item)
			}
		}

		if !stop {
			fmt.Fprintln(w, line)
		}

	}
	if err = w.Flush(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	uid := os.Getenv("INPUT_UID")
	if uid == "" {
		return
	}
	count, err := strconv.Atoi(os.Getenv("INPUT_MAX_COUNT"))
	if err != nil {
		count = 5
	}
	path := os.Getenv("INPUT_README_PATH")

	items := []string{}
	url := fmt.Sprintf("https://m.douban.com/rexxar/api/v2/status/user_timeline/%s?for_mobile=1", uid)
	RefererURL := fmt.Sprintf("https://m.douban.com/people/%s/statuses", uid)
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"),
	)

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", RefererURL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})

	c.OnResponse(func(r *colly.Response) {
		rs := &Result{}
		if err := json.Unmarshal(r.Body, rs); err != nil {
			log.Fatal(err)
			return
		}
		var (
			text                 []rune
			url, title, activity string
		)

		for index, item := range rs.Items {
			if index+1 > count {
				break
			}
			ResharedStatus := item.Status.ResharedStatus
			url = item.Status.SharingURL
			activity = item.Status.Activity
			var item_ string
			if ResharedStatus != (ResharedStatus_{}) {
				item_ = fmt.Sprintf("- [%s %s 的动态](%s)", activity, ResharedStatus.Author.Name, url)
			} else {
				if activity == "说" { // 广播
					text = []rune(item.Status.Text)
					if len(text) > MAX_LIMIT {
						activity = fmt.Sprintf("说: %s...", string(text[:MAX_LIMIT]))
					} else {
						activity = fmt.Sprintf("说: %s", item.Status.Text)
					}
				} else if strings.HasPrefix(activity, "收藏") { // 收藏
					card := item.Status.Card
					owner_id := card.OwnerURI[strings.LastIndex(card.OwnerURI,"/")+1:]
					origin_url := card.URL
					origin_title := "广播"
					if card.Activity == "说" {
						origin_url = fmt.Sprintf("https://www.douban.com/people/%s/status/%s/",
							owner_id, card.ID)
					}
					title := item.Status.Text[strings.LastIndex(item.Status.Text, " ")+1:]
					item_ = fmt.Sprintf("- [收藏 %s的%s](%s) 到 豆列[%s](%s)", card.OwnerName,
						origin_title, origin_url, title, url)
				} else if strings.HasPrefix(activity, "写了") { // 读书笔记
					text = []rune(activity)
					item_ = fmt.Sprintf("- [写了 %s %s 的读书笔记](%s)",
						string(text[4:len(text)-5]), item.Status.Card.Title, url)
				} else if strings.HasPrefix(activity, "转发") { // 转发
					act := []rune(activity)
					if len(act) == 2 {
						title = item.Status.Card.Title
					} else {
						title = string(act[2:])
					}
					item_ = fmt.Sprintf("- [%s %s](%s)", string(act[:2]), title, url)
				} else { // 书影音
					activity = string([]rune(activity)[:2])
				}
				if item_ == "" {
					item_ = fmt.Sprintf("- [%s %s](%s)", activity, item.Status.Card.Title, url)
				}
			}
			items = append(items, item_)
		}
	})

	c.OnScraped(func(r *colly.Response) {
		WriteToFile(path, items)
	})

	c.Visit(url)
}
