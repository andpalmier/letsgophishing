package main

import (
    "bufio"
    "crypto/tls"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "net/url"
    "os"
    "path"
    "strconv"
    "strings"
    "sync"
    "time"
    "github.com/PuerkitoBio/goquery"
)

// for colored output
const (
    Green = "\033[1;36m%s\033[0m"
    Yellow = "\033[1;33m%s\033[0m"
    Red = "\033[1;31m%s\033[0m"
)


var kitsFolder string
var intput_file string
var Nroutines int

func main() {

    var outputDir string

    // PARSING ARGUMENTS
    flag.StringVar(&intput_file, "i", "", "specify as input a file containing URLs; if not provided the tool will get the latest URLs from Phishtank")
    flag.StringVar(&outputDir, "o", ".", "specify output directory for the kits and the log files")
    flag.IntVar(&Nroutines, "r", 50, "number of goroutines to use")

    if len(os.Args)>1 && (os.Args[1]=="-h" || os.Args[1]=="--help"){
	flag.PrintDefaults()
	os.Exit(1)
    }

    flag.Parse()

    // print alert if Nroutines is higher than 100
    if Nroutines > 100{
	fmt.Printf(Red,"[!] ALERT: setting a number of go routines too high will cause some error (such as 'too many open files')\n")
    }
    // define folder where to download the files
    kitsFolder = path.Join(outputDir,"kits")
    if _, err := os.Stat(kitsFolder); os.IsNotExist(err) {
	os.Mkdir(kitsFolder,0755)
    }

    // if a file containing URLs is given, use it
    if intput_file!="" {
	local_file()
    } else {
	// otherwise use phishtank
	phishtank()
    }
}

// using local file
func local_file()  {
    fmt.Println("[+] Local file given: "+intput_file)

    // try to open the file
    file, err := os.Open(intput_file)
    if err != nil {
	// error if file does not exist
	fmt.Printf(Red, "Error: failed opening file\n")
	log.Fatal(err)
    }

    // scan file line per line
    scanner := bufio.NewScanner(file)
    scanner.Split(bufio.ScanLines)

    // channel for concurrent tasks
    ch := make(chan string)
    wg := sync.WaitGroup{}

    // limit to 100 goroutines
    for r:=0;r<Nroutines;r++{
	wg.Add(1)
	// start the goroutine
	go task(ch, &wg)
    }

    // parse the json file
    for scanner.Scan(){
	urlstring := scanner.Text()
	// push url in the channel
	ch <- urlstring
    }

    // close the channel
    close(ch)
    // wait the waitgroup to finish
    wg.Wait()

}


// get list of ulrs from phishtank
func phishtank() {
    phishtank_url := "http://data.phishtank.com/data/online-valid.json"
    fmt.Println("[+] Parsing URLs from Phishtank")

    // create http client to get the json
    client := &http.Client{Timeout: 10 * time.Second}

    // prepare for the GET request
    req, err := http.NewRequest("GET", phishtank_url, nil)
    if err != nil {
	fmt.Printf(Red,"Error while contacting Phishtank URL")
	log.Fatal(err)
    }

    // perform the GET request
    resp, err := client.Do(req)
    if err != nil {
	fmt.Printf(Red,"Error with Phishtank response")
	log.Fatal(err)
    }
    defer resp.Body.Close()

    // parse resulting json, structure is here:
    // https://www.phishtank.com/developer_info.php
    var parsed []interface{}

    // save the body of the response as a json
    body, err := ioutil.ReadAll(resp.Body)
    jsonErr := json.Unmarshal(body, &parsed)
    if jsonErr != nil {
	log.Fatalf("Error parsing Json from Phishtank: ",jsonErr)
    }

    // channel for concurrent tasks
    ch := make(chan string)
    wg := sync.WaitGroup{}

    // limit to 100 goroutines
    for r:=0;r<Nroutines;r++{
	wg.Add(1)
	// start goroutine
	go task(ch, &wg)
    }

    // parse the json file
    for _,item := range parsed {
	// get instace
	entry,_ := item.(map[string]interface{})
	// get ur of the instance
	urlstring := fmt.Sprintf("%v", entry["url"])
	// push url in the channel
	ch <- urlstring
    }

    // close the channel
    close(ch)
    // wait the waitgroup to finish
    wg.Wait()
}

// content of the goroutine
func task(ch chan string, wg *sync.WaitGroup){
    // for every line in the channel
    for line := range ch{
	go_phishing(line)
    }
    // task is complete
    wg.Done()
}

// from a url, navigate the path searching for zip files
func go_phishing(urlstring string){

    // remove "/" if is the last char of url
    if urlstring[len(urlstring)-1:]=="/"{
	urlstring=urlstring[:len(urlstring)-1]
    }

    // parse url path into a list
    scheme,err := url.Parse(urlstring)
    if err != nil {
	fmt.Printf("Error parsing URL: %s : %s\n",urlstring,err)
	return
    }

    // from website.com/aaa/bbb/ccc
    // to [aaa,bbb,ccc]
    paths := strings.Split(scheme.Path,"/")

    // iterating the lenght of the path list
    for i:=0;i<len(paths)-1;i++{

	// create valid URL using the scheme and the path list
	phish_url := string(scheme.Scheme+"://"+scheme.Host+strings.Join(paths[:len(paths)-i],"/"))


	//fmt.Println("[+] Checking: " + phish_url)

	// check if is there a zip in the path
	guess_zip(phish_url)

	// check if there is a open directory in the path
	guess_opendir(phish_url)

    }
}

// check if directory listing is enabled
func guess_opendir(phish_url string){

    // skip bad certificate error
    tr := &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }

    // create http client for request
    client := &http.Client{Timeout: 2 * time.Second, Transport : tr}
    req, err := http.NewRequest("GET",phish_url, nil)
    if err != nil {
	//fmt.Printf("Error while contacting URL: %s : %s\n",phish_url,err)
	return
    }

    // perform get request
    client.Do(req)
    req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")
    resp, err := client.Do(req)
    if err != nil {
	//fmt.Printf("Error while contacting URL: %s : %s\n",phish_url,err)
	return
    }
    defer resp.Body.Close()
    // get body of the response for inspection
    doc, err := goquery.NewDocumentFromResponse(resp)
    if err != nil {
	//fmt.Printf("Error inspecting HTML Body, URL: %s : %s\n",phish_url,err)
	return
    }
    // check if "index of" is in title of the body of the HTML
    title := doc.Find("title").Text()
    if strings.Contains(title,"Index of"){

	// open directory found
	fmt.Printf("[!] Open directory found: "+phish_url+"\n")
	// check if logfile for open directory is created
	opendir_log, err := os.OpenFile("opendir.txt",os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
	    fmt.Printf("Error creating opendir_log file, %s\n",err)
	    return
	}

	defer opendir_log.Close()
	// write in the logfile for open directories
	if _, err := opendir_log.WriteString(phish_url+"\n"); err != nil {
	    fmt.Printf("Error appending to the opendir_log file: %s\n",err)
	    return
	}

	// get links from the HTML of the response
	doc.Find("a[href]").Each(func(index int, item *goquery.Selection){
	    href, _ := item.Attr("href")
	    //fmt.Printf(" link %s text %s \n", href, item.Text())

	    // possible kit or list of victims or malware
	    if strings.HasSuffix(href,".zip") || strings.HasSuffix(href,".txt") || strings.HasSuffix(href,".exe"){

		//download file
		download_file(phish_url+"/"+href)
	    }
	})
    }
}

// try adding '.zip' at the end to find zip
func guess_zip(phish_url string){

    guess_url := phish_url+".zip"

    // skip bad certificate error
    tr := &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }

    // try to do a GET request
    client := &http.Client{Timeout: 2 * time.Second, Transport : tr}
    req, err := http.NewRequest("GET", guess_url , nil)
    if err != nil {
	//fmt.Printf("Error while contacting URL: %s : %s \n",guess_url,err)
	return
    }
    req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")
    resp, err := client.Do(req)
    if err != nil {
	//fmt.Printf("Error while contacting URL %s: %s\n", guess_url, err)
	return
    }

    // check if it's a zip
    if strings.Contains(resp.Header.Get("Content-Type"),"zip"){
	//fmt.Println(resp.Header.Get("Content-Type"))
	download_file(guess_url)
    }
}

// download a file
func download_file(file_url string){

    // split url in array
    zip_url_split := strings.Split(file_url,"/")
    currentTime := time.Now()
    // get timestamp to append it for naming the file
    //timestamp := currentTime.Format("010206150405")
    timestamp := strconv.FormatInt(currentTime.Unix(),10)

    // name the file with the timestamp and the last part of the URL
    filename := timestamp + "_"+zip_url_split[len(zip_url_split)-1]
    fname_split := strings.Split(filename,".")
    // get extension of the file
    extension := fname_split[len(fname_split)-1]

    // skip bad certificate error
    tr := &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }

    // prepare request to download the file
    client := &http.Client{Timeout: 30 * time.Second, Transport:tr}
    req, err := http.NewRequest("GET",file_url, nil)
    if err != nil {
	fmt.Printf("Error while contacting URL %s: %s\n", file_url, err)
	return
    }
    client.Do(req)
    req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")
    // perform the GET request
    resp, err := client.Do(req)
    if err != nil {
	fmt.Printf("Error while contacting URL %s: %s\n", file_url, err)
	return
    }
    defer resp.Body.Close()

    // place the file in the right folder
    kitfile := path.Join(kitsFolder,filename)
    out, err := os.Create(kitfile)
    if err != nil {
	fmt.Printf(Red,"Error creating file "+ filename)
	fmt.Println(err)
	return
    }
    defer out.Close()

    // Write the body to file
    _, err = io.Copy(out, resp.Body)
    if err != nil {
	fmt.Printf("Error writing file %s: %s\n", filename, err)
	return
    } else {
	// yellow output if it is a txt or exe
	color := Yellow
	if extension == "zip" {
	    // green output if it is a zip
	    color = Green
	}
	fmt.Printf(color,"[!] "+extension+" on "+file_url+" has been downloaded!\n")

	// save in log that the file was downloaded
	zip_log, err := os.OpenFile("files.txt",os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
	    log.Fatalf("Error writing file_logs %s\n", err)
	}

	defer zip_log.Close()
	if _, err := zip_log.WriteString(filename+", "+ file_url+"\n"); err != nil {
	    log.Fatalf("Error appending to file_logs %s\n", err)
	}
    }
}
