<p align="center">
  <img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/transparent-banner.png" alt="Gopher"/>
  <h1 align="center">go-recon</h1>
  <h4 align="center">External recon toolkit</h4>
</p>

<p align="center">

  <a href="https://opensource.org/licenses/MIT">
    <img src="https://img.shields.io/badge/license-MIT-_red.svg">
  </a>

  <a href="https://github.com/D3Ext/go-recon/blob/main/CHANGELOG.md">
    <img src="https://img.shields.io/badge/maintained%3F-yes-brightgreen.svg">
  </a>

  <a href="https://github.com/D3Ext/go-recon/issues">
    <img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat">
  </a>

  <a href="https://goreportcard.com/report/github.com/D3Ext/go-recon">
    <img src="https://goreportcard.com/badge/github.com/D3Ext/go-recon" alt="go report card">
  </a>

</p>

<p align="center">
  <a href="#introduction">Introduction</a> •
  <a href="#tools">Tools</a> •
  <a href="#installation">Installation</a> •
  <a href="#usage">Usage</a>
</p>

# Introduction

This project started as various Golang scripts to automatically perform tedious processes while performing external recon, between another bunch of things. Over the time I polished the scripts and finally decided to create much more versatile tools, in this way I would also learn to use Golang channels and concurrency.

This toolkit provides tools for different purposes (enum and exploitation) while performing external recon. Most functions are also available and can be used through the official package API for your own tools. Feel free to contribute by reporting issues or discussing ideas.

# General Features

This are some of the most notable features of this suite:

- Speed and concurrency
- Easy and malleable usage via CLI arguments
- Tools are designed to be combined between them
- Designed for Bug Bounty and external recon
- Multiple output formats (STDOUT, TXT, JSON, CSV)
- Take input as CLI arguments or directly from STDIN
- Direct access to official package API
- Coded in Golang to provide the best performance

# Tools

Every tool starts with "gr" as acronym of ***GoRecon*** in order to distinct their names from other tools

- ***gr-subdomains***: Enumerate subdomains of a domain using 8 different providers (passively)
- ***gr-urls***: Find URLs (endpoints) of a domain from different sources (Wayback, AlienVault)
- ***gr-probe***: Probe active subdomains and URLs (http and https) fastly, with custom concurrency and more
- ***gr-403***: Try to bypass pages that return 403 status code (multiple techniques)
- ***gr-openredirects***: Fuzz for potential open redirects on given URLs using a payload/list of payloads
- ***gr-aws***: Enumerate S3 buckets for given domain using permutations, verify bucket lists and much more
- ***gr-waf***: Identify which WAF is running on target using multiple payloads
- ***gr-filter***: Remove useless URLs from list using inteligent filtering, create custom filter patterns
- ***gr-replace***: Replace given keyword or parameter value with provided value from URLs of a list
- ***gr-secrets***: Search for API keys and leaked secrets in HTML and JS pages
- ***gr-crawl***: Fastly crawl urls for gathering URLs and JS endpoints, with custom depth and other config options
- ***gr-dns***: Retrieve DNS info from domains
- ***gr-whois***: Perform WHOIS query against domains

# Installation

Compile and install from source code via Github:

```sh
$ git clone https://github.com/D3Ext/go-recon
$ cd go-recon
$ make
$ sudo make install
```

The binaries will be compiled and installed on PATH, so you just will have to execute it from CLI

```sh
$ gr-subdomains
```

## Extra

To install a set of custom filters/patterns, you should execute the following command:

```sh
$ make extra
```

Then you can use them with `gr-filter`

# Usage

All tools have similar usage and CLI parameters to make recon easier

> Example help panel
```
  __ _  ___        _ __ ___  ___ ___  _ __
 / _` |/ _ \ _____| '__/ _ \/ __/ _ \| '_ \
| (_| | (_) |_____| | |  __/ (_| (_) | | | |
 \__, |\___/      |_|  \___|\___\___/|_| |_|
  __/ |     by D3Ext
 |___/      v0.2

Usage of gr-subdomains:
  INPUT:
    -d, -domain string      domain to find its subdomains (i.e. example.com)
    -l, -list string        file containing a list of domains to find their subdomains (one domain per line)

  OUTPUT:
    -o, -output string          file to write subdomains into (TXT format)
    -oj, -output-json string    file to write subdomains into (JSON format)
    -oc, -output-csv string     file to write subdomains into (CSV format)

  PROVIDERS:
    -all                      use all available providers to discover subdomains (slower than default)
    -p, -providers string[]   providers to use for subdomain discovery (separated by comma)
    -lp, -list-providers      list available providers

  CONFIG:
    -proxy string         proxy to send requests through (i.e. http://127.0.0.1:8080)
    -t, -timeout int      milliseconds to wait before each request timeout (default=5000)
    -c, -color            print colors on output
    -q, -quiet            print neither banner nor logging, only print output

  DEBUG:
    -version      show go-recon version
    -h, -help     print help panel

Examples:
    gr-subdomains -d example.com -o subdomains.txt -c
    gr-subdomains -l domains.txt -p crt,hackertarget -t 8000
    cat domain.txt | gr-subdomains -all -q
    cat domain.txt | gr-subdomains -p anubis -oj subdomains.json -c
```

See [here](https://github.com/D3Ext/go-recon/blob/main/USAGE.md) for ideas and real examples about how to use ***go-recon*** for external reconnaisance

# Demo

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo1.png">

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo2.png">

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo3.png">

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo4.png">

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo5.png">

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo6.png">

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo7.png">

# API

Install official ***go-recon*** Golang package like this:

```sh
$ go get github.com/D3Ext/go-recon/pkg/go-recon
```

If you want to use ***go-recon*** in your own Golang code see [here](https://github.com/D3Ext/go-recon/tree/main/examples)

# TODO

- More tools and features
- More optimization
- Email accounts enumeration
- ~Custom headers support (only on tools that send requests directly to the target)~
- ~Little fixes~
- ~CSV output supported by default on every tool~
- ~CLI parameter to configure custom user agents~
- ~More filtering patterns~
- ~More vulnerabilities payloads~
- ~WAF detection improved~

# References

Inspired and motivated by some awesome tools like this:

```
https://github.com/lc/gau
https://github.com/lc/subjs
https://github.com/tomnomnom/httprobe
https://github.com/projectdiscovery/subfinder
https://github.com/tomnomnom/waybackurls
https://github.com/projectdiscovery/nuclei
https://github.com/tomnomnom/qsreplace
https://github.com/hakluke/hakrawler
https://github.com/gocolly/colly/
https://github.com/d3mondev/puredns
https://github.com/blacklanternsecurity/bbot
https://github.com/s0md3v/uro
https://github.com/nytr0gen/deduplicate
https://github.com/smaranchand/bucky
https://github.com/projectdiscovery/interactsh
https://github.com/swisskyrepo/PayloadsAllTheThings
https://github.com/1ndianl33t/Gf-Patterns
https://github.com/r3curs1v3-pr0xy/sub404
https://github.com/devanshbatham/ParamSpider
https://github.com/m4ll0k/SecretFinder
https://github.com/MrEmpy/mantra
https://github.com/iamj0ker/bypass-403
https://github.com/edoardottt/favirecon
https://github.com/hueristiq/xs3scann3r
```

# License

This project is under MIT license

Copyright © 2024, *D3Ext*



