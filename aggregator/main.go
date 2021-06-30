package main

import (
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/deta/deta-go"
	"zgo.at/gadget"
	"zgo.at/isbot"
)

const key = "a014dfsf_B48gbpLSoKo9RRiseBbTScx4J341vWqD"

var (
	fReport = flag.String("report", "report.html", "Output file to write the report to")
)

func main() {
	flag.Parse()

	db, err := loadDB(key, "requests")
	if err != nil {
		panic(err)
	}

	report, err := db.aggregate(30, "www.larus.se")
	if err != nil {
		panic(err)
	}
	// pretty.Println(report)

	if err := db.writeReport(*fReport, report); err != nil {
		panic(err)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

type Record struct {
	Key       string
	Timestamp int64
	Timezone  string
	Useragent string
	Referrer  string
	Host      string
	Path      string
}

func (r *Record) Time() time.Time {
	return time.Unix(r.Timestamp, 0)
}

func (r *Record) Date() string {
	return r.Time().Format("2006-01-02")
}

func (r *Record) Page() string {
	return path.Dir(r.Path)
}

func (r *Record) ReferrerHost() string {
	u, err := url.Parse(r.Referrer)
	if err != nil {
		log.Printf("Error parsing referrer (%s): %s\n", r.Referrer, err)
		return "error"
	}
	return u.Host
}

func (r *Record) IsBot() bool {
	return isbot.Is(isbot.UserAgent(r.Useragent))
}

func (r *Record) Browser() string {
	return gadget.Parse(r.Useragent).BrowserName
}

////////////////////////////////////////////////////////////////////////////////////////////////////

//go:embed template.html
var tmpl string
var tmplFuncs = template.FuncMap{
	"limit": func(n int, items []*CounterItem) []*CounterItem {
		l := len(items)
		if n < 0 || n > l {
			n = l
		}
		return items[:n]
	},
}

type DB struct {
	_db  *deta.Base
	tmpl *template.Template
}

func loadDB(key, name string) (*DB, error) {
	t := template.Must(template.New("").Funcs(tmplFuncs).Parse(tmpl))
	d, err := deta.New(key)
	if err != nil {
		return nil, err
	}
	db, err := d.NewBase(name)
	if err != nil {
		return nil, err
	}
	return &DB{db, t}, nil
}

func (db *DB) writeReport(path string, report *Report) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return db.tmpl.Execute(f, report)
}

func (db *DB) aggregate(days int, domain string) (*Report, error) {
	ts := time.Now().AddDate(0, 0, -days).UTC().Unix()
	var records []*Record
	query := &deta.FetchInput{
		Q: deta.Query{
			{"timestamp?gt": ts},
		},
		Limit: 10,
		Dest:  &records,
	}

	report := newReport(domain)
	for {
		last, err := db._db.Fetch(query)
		if err != nil {
			return nil, err
		}
		for _, r := range records {
			report.Add(r)
		}
		if last == "" {
			break
		}
		query.LastKey = last
	}
	report.Stop()
	return report, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

type CounterItem struct {
	Key  string
	Hits int
}

func (c CounterItem) String() string {
	return fmt.Sprintf("%s: %d", c.Key, c.Hits)
}

type Counter struct {
	c []*CounterItem
}

func (cl *Counter) String() string {
	var ret []string
	for _, c := range cl.c {
		ret = append(ret, c.String())
	}
	return "[" + strings.Join(ret, ",") + "]"
}

func (cl *Counter) Len() int {
	return len(cl.c)
}

func (cl *Counter) Hits() int {
	var count int
	for _, c := range cl.c {
		count += c.Hits
	}
	return count
}

func (cl *Counter) Max() int {
	max := 0
	for _, c := range cl.c {
		if c.Hits > max {
			max = c.Hits
		}
	}
	return max
}

func (cl *Counter) Add(key string) {
	for _, c := range cl.c {
		if c.Key == key {
			c.Hits++
			return
		}
	}
	cl.c = append(cl.c, &CounterItem{
		Key:  key,
		Hits: 1,
	})
}

func (cl *Counter) ItemsByHits() []*CounterItem {
	sort.Slice(cl.c, func(i, j int) bool {
		return cl.c[i].Hits > cl.c[j].Hits
	})
	return cl.c
}

func (cl *Counter) ItemsByValue() []*CounterItem {
	sort.Slice(cl.c, func(i, j int) bool {
		return cl.c[i].Key < cl.c[j].Key
	})
	return cl.c
}

type NormalizedItem struct {
	Key    string
	Hits   int
	Height int
	X      int
	Y      int
}

func (cl *Counter) Normalize(width, height, marginWidth, marginHeight, max int) []NormalizedItem {
	const perc float64 = 100
	itemHeight := float64(height-marginHeight*2) / perc
	itemWidth := float64(width) / float64(cl.Len())
	var norm []NormalizedItem
	for i, c := range cl.ItemsByValue() {
		n := float64(c.Hits) / float64(max) * perc // Normalized value in percent
		h := math.Round(itemHeight * n)
		w := math.Round(itemWidth * float64(i))
		norm = append(norm, NormalizedItem{
			Key:    c.Key,
			Hits:   c.Hits,
			Height: int(h),
			X:      int(w) + marginWidth,
			Y:      height - int(h) - marginHeight,
		})
	}
	return norm
}

type YLabel struct {
	Value int
	Y     int
}

func (cl *Counter) YLabels(height, margin int) []YLabel {
	max := float64(cl.Max())
	const lines float64 = 10
	lineHeight := float64(height-margin*2) / lines
	var labels []YLabel
	for i := 0.0; i <= lines; i += 1.0 {
		h := math.Round(lineHeight * i)
		label := math.Round(max / lines * i)
		labels = append(labels, YLabel{
			Value: int(label),
			Y:     height - int(h) - margin,
		})
	}
	return labels
}

////////////////////////////////////////////////////////////////////////////////////////////////////

type Report struct {
	Host      string
	Duration  time.Duration
	Timestamp time.Time

	ViewsPerDay *Counter
	Pages       *Counter

	VisitsPerDay *Counter
	Referrers    *Counter

	Timezones *Counter
	Browsers  *Counter
	Bots      *Counter
}

func newReport(host string) *Report {
	now := time.Now()
	return &Report{
		Host:         host,
		Timestamp:    now,
		ViewsPerDay:  &Counter{},
		VisitsPerDay: &Counter{},
		Referrers:    &Counter{},
		Pages:        &Counter{},
		Timezones:    &Counter{},
		Browsers:     &Counter{},
		Bots:         &Counter{},
	}
}

func (r *Report) Stop() {
	if r.Duration != 0 {
		return
	}
	r.Duration = time.Since(r.Timestamp)
}

func (r *Report) Add(record *Record) {
	if record.Host != r.Host {
		return
	} else if record.IsBot() {
		r.Bots.Add(record.Useragent)
		return
	}
	r.ViewsPerDay.Add(record.Date())
	r.Pages.Add(record.Page())
	r.Browsers.Add(record.Browser())
	r.Timezones.Add(record.Timezone)
	ref := record.ReferrerHost()
	if ref != record.Host {
		r.VisitsPerDay.Add(record.Date())
		if ref != "" {
			r.Referrers.Add(ref)
		}
	}
}
