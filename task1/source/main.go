package main

/*
Devlog
V1
1. an http tool to query the content of the page √
2. an html tool to check the content of the page. or consider use regex to get http √
V2
3. make an http server service to accpet the request √
4. accept url with custom destination url √
V3
5. considering script work as a server, mv log.fatal to log.Printf to avoid main func exit
6. add a logic for other website other than AKS
V4
7. support post method to give multiple urls in one time
8. use go routine to check all url in the same time, use channel to get the completed signal of each routine and use mutex lock to avoid different routine write to reply channel in same time(here is a bug considering using overall channel)
9. update unused path reply to give currently useful path
*/

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// use lock to avoid different routine compete response writer
var lock sync.Mutex

// get url and reply the html of target url
func GetUrlContent(url string) (body []byte, err error, ErrorState bool) {
	res, err := http.Get(url)
	if err != nil {
		//log.Fatal(err)
		log.Printf("%v\n", err)
		ErrorState = true
		return
	}
	defer res.Body.Close()
	body, err = io.ReadAll(res.Body)
	if res.StatusCode > 299 {
		log.Printf("Response failed with status code: %d and\nbody: %v\n", res.StatusCode, body)
		ErrorState = true
		return
	}
	if err != nil {
		log.Printf("%v\n", err)
		ErrorState = true
		return
	}
	return
}

// give error message when access to path other than url
func Helper(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "You are under path %q, currently there is no function under this path\nfunctinal path now: url\n", r.URL.Path)
}

// Print result which will be called by GerUrlInside
func PrintResult(w http.ResponseWriter, Urls string, start time.Time) {
	reg := regexp.MustCompile(`href="(.*?)"`) // for common url
	UrlContent, err, ErrorState := GetUrlContent(Urls)
	if !ErrorState {
		if Urls == "https://learn.microsoft.com/en-us/azure/aks" {
			reg = regexp.MustCompile(`<a class="has-external-link-indicator font-size-sm display-block" href="(.*?)" data-linktype="(.*?)">`)
		} // reg for AKS web
		if reg == nil {
			fmt.Println("MustCompile err")
			return
		}
		result := reg.FindAllStringSubmatch(string(UrlContent), -1) // get all href inside html
		secs := time.Since(start).Seconds()
		lock.Lock()
		defer lock.Unlock()
		fmt.Fprintf(w, "%.2fs  Result for Url %s:\n", secs, Urls)
		switch {
		case Urls == "https://learn.microsoft.com/en-us/azure/aks":
			for _, UrlFind := range result {
				switch UrlFind[2] {
				case "relative-path":
					fmt.Fprintf(w, "-  https://learn.microsoft.com/en-us/azure/aks/%s\n", UrlFind[1])
				case "absolute-path":
					fmt.Fprintf(w, "-  https://learn.microsoft.com/%s\n", UrlFind[1])
				case "external":
					fmt.Fprintf(w, "-  %s\n", UrlFind[1])
				}
			} // specific for AKS website since there are three kinds of href inside page
		default:
			regcheck := regexp.MustCompile(`\Ahttps?://.+`)
			for _, UrlFind := range result {
				if regcheck.MatchString(UrlFind[1]) {
					fmt.Fprintf(w, "-  %s\n", UrlFind[1])
				}
			}
		}
		fmt.Fprintf(w, "\n\n")
		return
	}
	fmt.Fprintf(w, "%v\n", err)
}

func FormatUrl(w http.ResponseWriter, query string, CCh chan<- bool) {
	var Urls string
	defer func() { CCh <- true }()
	start := time.Now()
	Regq := regexp.MustCompile(`\Ahttps?://.+`) // check if query valid
	switch {
	case query == "":
		Urls = "https://learn.microsoft.com/en-us/azure/aks"
	case Regq.MatchString(query):
		//	Urls = query[4:]
		Urls = query
	default:
		lock.Lock()
		fmt.Fprintf(w, "query \"%s\" is not valid\n, query should be a valid url start with http:// or https:// \n", query)
		lock.Unlock()
		return
	}
	PrintResult(w, Urls, start)
}

// main body of creeper
func GetUrlInside(w http.ResponseWriter, r *http.Request) {
	CCh := make(chan bool)
	query := r.URL.RawQuery
	fmt.Fprintf(w, " query is: %q\n", query)
	switch r.Method {
	case "GET":
		go FormatUrl(w, query, CCh)
		<-CCh
	case "POST":
		UrlData, _ := io.ReadAll(r.Body)
		UrlsGroup := regexp.MustCompile(",").Split(string(UrlData), -1)
		for _, i := range UrlsGroup {
			go FormatUrl(w, strings.Trim(i, " "), CCh)
		}
		for range UrlsGroup {
			<-CCh
		}
	default:
		fmt.Fprintf(w, "Method \"%s\" is not allowed, current support method: POST, GET\n", r.Method)
	}
}

func main() {
	http.HandleFunc("/url", GetUrlInside) //for creeper
	http.HandleFunc("/", Helper)
	log.Fatal(http.ListenAndServe("0.0.0.0:8848", nil))
}

