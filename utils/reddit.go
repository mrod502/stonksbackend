package utils

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/mrod502/logger"
	"github.com/shopspring/decimal"
)

type RedditCache struct {
	v map[string]T3Data
	sync.RWMutex
}

func (c *RedditCache) Get(k string) T3Data {
	c.RLock()
	defer c.RUnlock()
	return c.v[k]
}

func (c *RedditCache) Set(v T3Data) {
	c.Lock()
	defer c.Unlock()

	c.v[v.ID] = v
}

func (c *RedditCache) SetBulk(vals []T3Data) {
	c.Lock()
	defer c.Unlock()
	for _, v := range vals {
		c.v[v.ID] = v
	}
}
func (c *RedditCache) All() (d []T3Data) {

	d = make([]T3Data, 0, len(c.v))
	c.RLock()
	defer c.RUnlock()
	for _, v := range c.v {
		d = append(d, v)
	}
	return
}

func NewRedditCache() (c *RedditCache) {
	return &RedditCache{v: make(map[string]T3Data)}
}

const (
	ratelimitRemaining = "x-ratelimit-remaining"
	ratelimitReset     = "x-ratelimit-reset"
)

var (
	redditURL = "https://www.reddit.com/"
)

type BoardHome struct {
	Data RedditData `json:"data"`
}

type RedditData struct {
	After    string
	Before   string
	Children []Link
}

type Link struct {
	Kind string
	Data T3Data `json:"data"`
}
type T3Data struct {
	ID                string              `json:"id"`
	Created           float64             `json:"created"`
	CreatedUTC        float64             `json:"created_utc"`
	LinkFlairRichText []LinkFlairRichText `json:"link_flair_richtext"`
	Author            string              `json:"author"`
	Title             string              `json:"title"`
	Selftext          string              `json:"selftext"`
	Ups               int                 `json:"ups"`
	Downs             int                 `json:"downs"`
	UpvoteRatio       decimal.Decimal     `json:"upvote_ratio"`
	AllAwardings      []Awarding          `json:"all_awardings"`
	Body              string              `json:"body"`
	Subreddit         string              `json:"subreddit"`
	Symbols           []string
}

func (t T3Data) GetSymbols() T3Data {
	t.Symbols = []string{}
	return t
}

type Awarding struct {
	AwardType   string `json:"award_type"`
	CoinPrice   int    `json:"coin_price"`
	Description string `json:"desctiption"`
	Count       int    `json:"count"`
	Name        string `json:"name"`
	IsEnabled   bool   `json:"is_enabled"`
}

type LinkFlairRichText struct {
	E string `json:"e"`
	T string `json:"t"`
}

func (l Link) IsDailyDiscussion() bool {
	if len(l.Data.LinkFlairRichText) == 0 {
		return false
	}
	return l.Data.LinkFlairRichText[0].T == "Daily Discussion"
}
func (r RedditData) ChildIDs() (ids []string) {
	ids = make([]string, 0, len(r.Children))
	for _, val := range r.Children {
		ids = append(ids, val.Data.ID)
	}
	return ids
}

type RedditCommentResponse []RedditData

type ListingArray []RedditListing

type RedditListing struct {
	Kind string
	Data RedditData `json:"data"`
}

func (r RedditCommentResponse) AllChildren() (t []Link) {
	totalChildren := 0
	for _, v := range r {
		totalChildren += len(v.Children)
	}

	t = make([]Link, 0, totalChildren)

	for _, listing := range r {
		t = append(t, listing.Children...)
	}
	return
}

func (r ListingArray) AllChildren() (t []Link) {
	totalChildren := 0
	for _, v := range r {
		totalChildren += len(v.Data.Children)
	}

	t = make([]Link, 0, totalChildren)

	for _, listing := range r {
		t = append(t, listing.Data.Children...)
	}
	return
}

func (b BoardHome) Subreddit() string {
	if len(b.Data.Children) == 0 {
		return ""
	}
	return b.Data.Children[0].Data.Subreddit
}

func (b BoardHome) GetAllDiscussion() (t []T3Data, waitTime time.Duration, err error) {
	childIDs := b.Data.ChildIDs()

	for _, v := range childIDs {
		listings, remaining, reset, err := GetCommentListing(b.Subreddit(), v, "?sort=top")

		if err != nil {
			logger.Error("reddit", "GetAllDiscussion", err.Error())
		}
		for _, child := range listings.AllChildren() {
			t = append(t, child.Data)
		}
		time.Sleep(time.Second * (time.Duration(safeDivide(reset, remaining)) + 1))
		waitTime = time.Second * (time.Duration(safeDivide(reset, remaining)) + 1)
	}
	return
}

func GetCommentListing(board, id, opts string) (r ListingArray, remaining int64, reset int64, err error) {
	b, header, err := BrowserRequest(redditURL+fmt.Sprintf("/r/%s/comments/%s.json", board, id)+opts, "www.reddit.com")
	if err != nil {
		logger.Error("reddit", "GetCommentListing:1", err.Error())
	}
	remaining, _ = GetInt64HeaderVal(header, ratelimitRemaining)

	reset, _ = GetInt64HeaderVal(header, ratelimitReset)

	err = json.Unmarshal(b, &r)

	return
}

func GetBoard(boardName string) (t []T3Data, waitTime time.Duration, err error) {
	var board BoardHome
	b, _, err := BrowserRequest(redditURL + fmt.Sprintf("/r/%s.json", boardName))

	if err != nil {
		return
	}

	err = json.Unmarshal(b, &board)
	if err != nil {
		return
	}
	for _, v := range board.Data.Children {
		t = append(t, v.Data)
	}

	comments, waitTime, err := board.GetAllDiscussion()
	t = append(t, comments...)

	return
}

func safeDivide(n, d int64) int64 {
	if d == 0 {
		return 0
	}
	return n / d
}
