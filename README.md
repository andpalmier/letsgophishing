# LetsGoPhishing

<p align="center">
  <img alt="gopherphishing" src="https://github.com/andpalmier/letsgophishing/blob/master/gopherphishing.png?raw=true" />
  <p align="center">
    <a href="https://github.com/andpalmier/letsgophishing/blob/master/LICENSE"><img alt="Software License" src="https://img.shields.io/badge/license-GPL3-brightgreen.svg?style=flat-square"></a>
    <a href="https://goreportcard.com/report/github.com/andpalmier/letsgophishing"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/andpalmier/letsgophishing?style=flat-square"></a>
    <a href="https://twitter.com/intent/follow?screen_name=andpalmier"><img src="https://img.shields.io/twitter/follow/andpalmier?style=social&logo=twitter" alt="follow on Twitter"></a>
  </p>
</p>

This tool was written to inspect a list of URLs and check if they are hosting phishing pages by looking at the `title` tag in the retrieved page and at `URL/admin/` (where usually are located the panels of known phishing kits, eg. 16shop). `letsgophishing` makes use of the goroutines and channels to parallelize the requests.

## Usage

Build the executable with `go build lgphish.go`. Then:

```
$ cat urls | lgphish -o output -r 100 -c config.json

-o: path to output file containing only suspicious URLs
-r: number of goroutines to create
-c: path to config file (json format)
```

## Config.json

An example of config file is provided in `config.json`; you can create your own config and specify the path with `-c`.

The config file allows to specify:

- `SuspiciousTitles`: if the `title` attribute of the HTML page at the specified URL contains one of the string in this array, the URL will be considered as suspicious.
- `KitsTitles`: if the `title` attribute of the HTML page at `<specified_URL>/admin/` contains one of the string in this array, the URL could host a phishing kit.
