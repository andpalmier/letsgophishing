# Letsgophishing

<p align="center">
  <img alt="goransom" src="https://github.com/andpalmier/letsgophishing/blob/master/gopherphishing.png?raw=true" />
  <p align="center">
    <a href="https://github.com/andpalmier/letsgophishing/blob/master/LICENSE"><img alt="Software License" src="https://img.shields.io/badge/license-GPL3-brightgreen.svg?style=flat-square"></a>
    <a href="https://goreportcard.com/report/github.com/andpalmier/goransom"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/andpalmier/goransom?style=flat-square"></a>
    <a href="https://twitter.com/intent/follow?screen_name=andpalmier"><img src="https://img.shields.io/twitter/follow/andpalmier?style=social&logo=twitter" alt="follow on Twitter"></a>
  </p>
</p>

This tool was written to inspect a huge list of URLs and check if they are hosting phishing pages by looking at the `title` tag in the retrieved page and at `URL/admin/` (where usually are located the panels of known phishing kits, eg. 16shop). `letsgophishing` makes use of the goroutines and channels to parallelize the requests.

## Usage

Build the executable with `go build letsgophishing.go`. Then:

```
$ ./letsgophishing -i inputFile -o outputFile -c 100


  -i string
    	specify as input a file containing one URL per line (required)
  -o string
    	specify output directory for the kits and the log files (default "phishingulrs.txt")
  -c int
    	number of goroutines to use (default 100)
```

