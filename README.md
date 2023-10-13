<p align="center">
  <img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/transparent-banner.png" alt="Gopher"/>
  <h1 align="center">go-recon</h1>
  <h4 align="center">External recon toolkit</h4>
  <h6 align="center">Coded with 💙 by D3Ext</h6>
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
  <a href="#usage">Usage</a> •
  <a href="#contributing">Contributing</a>
</p>

<p align="center">
  <a href="https://github.com/D3Ext/go-recon/blob/main/SPANISH.md">README in Spanish</a>
</p>

# Introduction

This project started as various Golang scripts to automatically perform tedious processes and to perform external recon, between another bunch of things. Over time I polished the tools and finally decided to take it seriously, in this way I would also learn to use Golang channels and concurrency, so tools are fast and configurable.

This toolkit provides tools for different purposes while performing external recon. Most functions are also available and can be used through the official package API for your own tools. Feel free to contribute by reporting issues or discussing ideas.

See [Wiki](https://github.com/D3Ext/go-recon/wiki) for further info

# Tools

- ***gr-subdomains***: Enumerate subdomains of a domain using 8 different providers (passively)
- ***gr-urls***: Find URLs (endpoints) of a domain from different sources (Wayback, AlienVault)
- ***gr-probe***: Probe active subdomains and URLs (http and https) fastly, custom concurrency and more
- ***gr-403***: Try to bypass pages that return 403 status code (multiple techniques)
- ***gr-openredirects***: Fuzz for potential open redirects on given URLs using a payload/list of payloads
- ***gr-dns***: Retrieve DNS info from domains
- ***gr-aws***: Enumerate S3 buckets for given domain using permutations, verify bucket lists and much more
- ***gr-waf***: Identify which WAF is running on a domain
- ***gr-filter***: Remove useless URLs from list, apply filters, create custom filter patterns
- ***gr-replace***: Replace given keyword or parameter value with provided value from URLs of a list
- ***gr-secrets***: Search for API keys and leaked secrets in HTML and JS pages
- ***gr-crawl***: Fastly crawl urls for gathering URLs and JS endpoints, with custom depth and other options
- ***gr-whois***: Perform WHOIS query against domains

# Features

- Speed and concurrency
- Easy usage and configurable via CLI arguments
- Tools can be combined between them
- Multiple output formats (STDOUT, TXT, JSON, CSV)
- Input as URL, domains or STDIN
- Direct access to official package API
- Tested on Linux

# Installation

Compile and install from source code via Github:

```sh
git clone https://github.com/D3Ext/go-recon
cd go-recon
make
sudo make install
```

The binaries will be compiled and installed on PATH, so you just will have to execute it from CLI

```sh
$ gr-subdomains
```

## Extra

To install a set of custom filters/patterns and a Bash autocompletion script, you could execute the following command:

```sh
make extra
```

Then if you press TAB twice when using gr-subdomains or gr-filter, you will see the available providers and filters.

# Usage

All tools have similar usage and CLI parameters

> Example help panel
```
  __ _  ___        _ __ ___  ___ ___  _ __
 / _` |/ _ \ _____| '__/ _ \/ __/ _ \| '_ \
| (_| | (_) |_____| | |  __/ (_| (_) | | | |
 \__, |\___/      |_|  \___|\___\___/|_| |_|
  __/ |     by D3Ext
 |___/      v0.1

Usage of gr-subdomains:
  INPUT:
    -d, -domain string      domain to find its subdomains (i.e. example.com)
    -l, -list string        file containing a list of domains to find their subdomains (one domain per line)

  OUTPUT:
    -o, -output string          file to write subdomains into
    -oj, -output-json string    file to write subdomains into (JSON format)

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
    cat domain.txt | gr-subdomains -all
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
go get github.com/D3Ext/go-recon/pkg/go-recon
```

If you want to use ***go-recon*** in your own Golang code see [here](https://github.com/D3Ext/go-recon/tree/main/examples)

# TODO

- ~~Parameter to control used providers~~
- ~~CSV output~~
- More tools and features
- ~~Dockerfile~~
- ~~Changelog~~
- HTML results reports
- More optimization
- ~~Compare results with other tools such as **subfinder**, **gau**, **httprobe**...~~

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

# Contributing

See [CONTRIBUTING.md](https://github.com/D3Ext/go-recon/blob/main/CONTRIBUTING.md)

# Changelog

See [CHANGELOG.md](https://github.com/D3Ext/go-recon/blob/main/CHANGELOG.md)

# License

This project is under MIT license

Copyright © 2023, *D3Ext*

# Support

<a href="https://www.buymeacoffee.com/D3Ext" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-blue.png" alt="Buy Me A Coffee" style="height: 60px !important;width: 217px !important;" ></a>


