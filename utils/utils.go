package utils

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/gookit/color"
	"net/http"
	"os"
)

// config file struct
type Config struct {
	SuspiciousTitles []string
	KitsTitles       []string
	ToRemove         []string
}

// given a URL, get the title of the HTML
func GetTitle(url string, client *http.Client) string {

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
func ParseConfig(configFile string) Config {

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
