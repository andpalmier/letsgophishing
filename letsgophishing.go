/*

./letsgophishing -i input -o output -c 100

-i: path to input file containing 1 URL per line
-o: path to output file containing only suspicious URLs
-c: number of goroutines to create

TODO
- reduce duplicates

*/

package main

import (
	"bufio"
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
	configFile = "config.json"

	// decide this in another way?
	concurrency int

	// path to file containing list of URLs
	inputFile string

	// path to file containing output
	outputFile string

	// config data
	conf utils.Config

	// client to make requests
	client = &http.Client{Timeout: 3 * time.Second}

	// set of already written
	written = make(map[string]bool)
)

// gets url from urlsChan (coming from the file) and - if page is suspicious - push it into phishChan
func urlChecker(wg *sync.WaitGroup, urlsChan <-chan string, phishChan chan<- string) {
	// waitgroup should not be complete
	defer wg.Done()

	// set of already written
	analyzed := make(map[string]bool)

	// loop in the channel of URLs from file
	for url := range urlsChan {

		parsed := strings.Split(url, ".")
		length := len(parsed)

		// remove some strings for possible duplicates
		for _, rem := range conf.ToRemove {
			if len(parsed) > 1 && parsed[1] == rem {
				parsed = parsed[1:]
			}
		}

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
				title := strings.ToLower(utils.GetTitle("http://"+url+"/admin/", client))

				// check if title of the page in *url*/admin/ is a known phishing kit
				for _, kitname := range conf.KitsTitles {

					// FOUND!
					if strings.Contains(title, kitname) {
						color.Green.Printf("[!] Suspected phishing kit: %s at %s\n", title, url)

						// send to channel!
						phishChan <- url
					}
				}

				// find title of the page in URL
				title = strings.ToLower(utils.GetTitle("http://"+url, client))

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

	// channel for urls from file
	urlsChan := make(chan string)

	// channel where suspicious urls are pushed
	phishChan := make(chan string)

	// wait group used to sync goroutines
	var wg sync.WaitGroup

	// specify input file with i
	flag.StringVar(&inputFile, "i", "", "file containing URLs")

	// specify output file with o
	flag.StringVar(&outputFile, "o", "phishingurls.txt", "output file containing suspicious URLs")

	// specify goroutines number with c
	flag.IntVar(&concurrency, "c", 100, "number of goroutines")

	// parse flags
	flag.Parse()

	// open file
	if inputFile == "" {
		color.Red.Printf("Please provide an input file using: -i\n")
		os.Exit(1)
	}
	file, err := os.Open(inputFile)
	if err != nil {
		color.Red.Printf("Error opening the input file: %s\n", err)
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
		// create scanner to read line by line
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			url := scanner.Text()

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
	f, err := os.Create(outputFile)
	if err != nil {
		color.Red.Printf("Error creating output file: %s\n", err)
		os.Exit(1)
	}
	defer f.Close()

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
			_, err := f.WriteString(phishUrl + "\n")
			if err != nil {
				color.Red.Printf("Error writing to output file: %s\n", err)
				os.Exit(1)
			}
		}

	}
	fmt.Printf("\nSuspicious urls found: %d\n", counter)
}

/*
// given a URL, get the title of the HTML
func getTitle(url string) string {

	// make request to check title
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		//fmt.Printf("Error making the request: %s\n", err)
		return "Not found"
	}

	// use neutral and non blocking UA
	//req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	// iPhone UA
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148")

	// get response in HTML
	resp, err := client.Do(req)
	if err != nil {
		//fmt.Printf("Error getting the response: %s\n", err)
		return "Not found"
	}

	// get title of the page that was in the response
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		//fmt.Printf("Error getting the response: %s\n", err)
		return "Not found"
	}
	title := doc.Find("title").Text()
	return title
}


// parse config file
func parseConfig() Config {

	// open configuration file
	file, err := os.Open(configFile)
	if err != nil {
		color.Red.Printf("Error finding config file: %s\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// decode the json
	decoder := json.NewDecoder(file)
	conf := Config{}
	err = decoder.Decode(&conf)
	if err != nil {
		color.Red.Printf("Error parsing config file: %s\n", err)
		os.Exit(1)
	}

	// return the config struct
	return conf
}
*/
