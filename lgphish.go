/*

cat urls | lgphish -o output -r 100 -c config.json

-o: path to output file containing only suspicious URLs
-r: number of goroutines to create
-c: config file

*/

package main

import (
    "flag"
    "fmt"
    "github.com/gookit/color"
    "letsgophishing/utils"
    "net/http"
    "os"
    "strings"
    "sync"
    "time"
)

var (
    // config file name
    configFile string
    // decide this in another way?
    concurrency int

    // path to file containing output
    outputFile string

    // config data
    conf utils.Config

    // client to make requests
    client = &http.Client{Timeout: 3 * time.Second}

    // set of already written
    written = make(map[string]bool)
)

// gets url from urlsChan (coming from stdin) and - if page is suspicious - push it into phishChan
func urlChecker(wg *sync.WaitGroup, urlsChan <-chan string, phishChan chan<- string) {
    // waitgroup should not be complete
    defer wg.Done()

    // set of already written
    analyzed := make(map[string]bool)

    // loop in the channel of URLs from stdin
    for url := range urlsChan {

	parsed := strings.Split(url, ".")
	length := len(parsed)

	// iterate over the subdomains
	// ex. icloud.fakeapple.au.co --> au.co, fakeapple.au.co...
	for i := 0; i < length-1; i++ {

	    // from list of subdomains create a URL
	    url := strings.Join(parsed[i:], ".")

	    // if it was not analyzed
	    if !analyzed[url] {

		// add to list of analyzed urls
		analyzed[url] = true

		//fmt.Println(url)

		// get title of the page in lowercase
		title, err := utils.GetTitle("http://"+url+"/admin/", client)
		if err != nil {
		    // title not found
		    continue
		}
		title = strings.ToLower(title)

		// check if title of the page in *url*/admin/ is a known phishing kit
		for _, kitname := range conf.KitsTitles {

		    // FOUND!
		    if strings.Contains(title, kitname) {
			color.Green.Printf("[!] Suspected phishing kit: %s at %s\n", title, url)

			// send to channel!
			phishChan <- url+"/admin/"
		    }
		}

		// find title of the page in URL
		title, err = utils.GetTitle("http://"+url, client)
		if err != nil {
		    // title not found
		    continue
		}
		title = strings.ToLower(title)

		// check if title of the page is suspicious
		for _, susp := range conf.SuspiciousTitles {

		    // FOUND!
		    if strings.Contains(title, susp) {
			color.Yellow.Printf("[!] Suspicious title found: %s at %s\n", title, url)
			// send to channel!
			phishChan <- url
		    }
		}
	    }
	}
    }

}

func main() {

    // channel for urls from stdin
    urlsChan := make(chan string)

    // channel where suspicious urls are pushed
    phishChan := make(chan string)

    // wait group used to sync goroutines
    var wg sync.WaitGroup

    // specify output file with o
    flag.StringVar(&outputFile, "o", "phishingurls.txt", "output file containing suspicious URLs")

    // specify config file with c
    flag.StringVar(&configFile, "c", "config.json", "config file")

    // specify goroutines number with r
    flag.IntVar(&concurrency, "r", 100, "number of goroutines")

    // parse flags
    flag.Parse()

    // get input from stdin
    input, err := utils.GetInput()
    if err != nil {
	// raise error if no input is provided
	color.Red.Printf("Error getting user input: %s\n", err)
	os.Exit(1)
    }

    // parse json config
    conf = utils.ParseConfig(configFile)

    // concurrency is set
    for i := 0; i < concurrency; i++ {
	wg.Add(1)

	// start goroutine that will wait for urls to be pushed in urlsChan
	go urlChecker(&wg, urlsChan, phishChan)
    }

    // push urls in urlsChan
    go func() {

	// get url from list
	for _, url := range input {

	    // clean URL
	    if strings.HasPrefix(url, "https://") {
	    url = url[8:]
	} else if strings.HasPrefix(url, "http://") {
	url = url[7:]
    }

    // push url in urlsChan (will be pulled from urlChecker)
    urlsChan <- url
}

// close channel because we wrote everything we need
close(urlsChan)
	}()

	// start goroutine to monitor if phishChan is complete
	go func() {
	    wg.Wait()
	    close(phishChan)
	}()

	// create result file
	phishf, err := os.Create(outputFile)
	if err != nil {
	    color.Red.Printf("Error creating output file: %s\n", err)
	    os.Exit(1)
	}

	defer phishf.Close()

	counter := 0
	// loop in URLs in phishChan
	for phishUrl := range phishChan {
	    // if not already in output file
	    if !written[phishUrl] {
		// write in output file
		written[phishUrl] = true

		// increase counter
		counter++

		// write them in a file
		_, err := phishf.WriteString(phishUrl + "\n")
		if err != nil {
		    color.Red.Printf("Error writing to output file: %s\n", err)
		    os.Exit(1)
		}
	    }

	}

	fmt.Printf("\nSuspicious urls found: %d\n", counter)
    }
