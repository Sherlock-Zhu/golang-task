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
*/

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
)

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
	fmt.Fprintf(w, "You are under path %q, currently there is no function under this path\n", r.URL.Path)
}

// main body of creeper
func GetUrlInside(w http.ResponseWriter, r *http.Request) {
	var Urls string
	reg := regexp.MustCompile(`href="(.*?)"`) // for common url
	query := r.URL.RawQuery
	fmt.Fprintf(w, " %q\n", query)
	Regq := regexp.MustCompile(`\Aurl=https?://.+`) // check if query valid
	switch {
	case query == "":
		Urls = "https://learn.microsoft.com/en-us/azure/aks"
	case Regq.MatchString(query):
		Urls = query[4:]
		fmt.Printf("%s\n", Urls)
	default:
		fmt.Fprintf(w, "query \"%s\" is not valid\n", query)
		return
	}
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
		switch {
		case Urls == "https://learn.microsoft.com/en-us/azure/aks":
			for _, UrlFind := range result {
				switch UrlFind[2] {
				case "relative-path":
					fmt.Fprintf(w, "https://learn.microsoft.com/en-us/azure/aks/%s\n", UrlFind[1])
				case "absolute-path":
					fmt.Fprintf(w, "https://learn.microsoft.com/%s\n", UrlFind[1])
				case "external":
					fmt.Fprintf(w, "%s\n", UrlFind[1])
				}
			} // specific for AKS website since there are three kinds of href inside page
		default:
			for _, UrlFind := range result {
				fmt.Fprintf(w, "%s\n", UrlFind[1])
			}
		}
		return
	}
	fmt.Fprintf(w, "%v\n", err)
}

func main() {
	http.HandleFunc("/url", GetUrlInside) //for creeper
	http.HandleFunc("/", Helper)
	log.Fatal(http.ListenAndServe("0.0.0.0:8848", nil))
}

