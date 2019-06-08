package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type urlCache struct {
	url map[string]bool
	mx  sync.Mutex
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, uc *urlCache, wg *sync.WaitGroup) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	defer wg.Done()
	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	uc.mx.Lock()
	if _, urlExists := uc.url[body]; !urlExists {
		if err != nil {
			fmt.Println(err)
			uc.url[body] = false //URL visited, but not found
			uc.mx.Unlock()
			return
		}
		uc.url[body] = true  //URL visited, but and found
		fmt.Printf("found: %s %q\n", url, body)
	}
	uc.mx.Unlock()
	for _, u := range urls {
		wg.Add(1)
		go Crawl(u, depth-1, fetcher, uc, wg)
	}
	return
}

func main() {
	// URLs cache used in order not to fetch URLs twice.
	var urlc urlCache
	// WaitGroup makes sure that program ends only after all
	// goroutines have finished.
	var wg sync.WaitGroup
	urlc.url = make(map[string]bool)
	
	wg.Add(1)
	go Crawl("https://golang.org/", 4, fetcher, &urlc, &wg)
	wg.Wait()
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
