# Letsgophishing

![](img/gopherphishing.png)

This tool is my attempt to rewrite [phishfinder from cybercdh](https://github.com/cybercdh/phishfinder) in Go. This version make use of goroutines to make things faster.

Given a list of URLs, `phishfinder.go` will explore all the paths and open directories in order to find zip/txt/exe files and download them. This may be useful when threat hunting, since it is possible that these files contain the source code of the phishing kits, the logs of the victims or even malware.

Similarly to the original Python script, the tool will attempt to guess the name of the zip file, since it is often the case that the name is the same as the current URI folder:

```
https://malicious.com/aaa/bbb/malicious
https://malicious.com/aaa/bbb.zip
```

## Usage

```
$ go run phishfinder.go -h
  -i string
    	specify as input a file containing URLs; if not provided the tool will get the latest URLs from Phishtank
  -o string
    	specify output directory for the kits and the log files (default ".")
  -r int
    	number of goroutines to use (default 50)
```

**NOTE:** If the number of goroutines is too high, it will result in some errors, such as *'too many open files'*.

## Todo

- Use a static list for the extensions we want to look for
- Remove dependency from [goquery](https://github.com/PuerkitoBio/goquery)
