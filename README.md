<p align="center">
  <img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/banner.png" width="130" heigth="60" alt="Gopher"/>
  <h1 align="center">go-recon</h1>
  <h4 align="center">Bug Bounty and external recon toolkit</h4>
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

# Introduction

This project started over a year ago as a simple Python script to find subdomains using SSL certs transparency and other APIs, then I decided to implement a bunch of functions to enumerate more things but it was really slow and wasn't stable so I decided to split up the project for the different tasks into multiple tools, this time in Golang and much fast. Moreover, I decided to develop this project to learn how to use concurrency and channels.

This toolkit provides tools for different purposes while performing external recon. Most functions are also available and can be used through the official package API for your own tools.

README in spanish [here](https://github.com/D3Ext/go-recon/blob/main/SPANISH.md)

# Tools

- ***gr-subdomains*** Find all subdomains of a domain (passive)
- ***gr-urls*** Find domain URLs from different sources
- ***gr-probe*** Probe active subdomains (http and https)
- ***gr-js*** Extract JS endpoints from URLs
- ***gr-403*** Try to bypass pages that return 403 status code (multiple techniques)
- ***gr-openredirects*** Fuzz for possible open redirects on given URLs
- ***gr-dns*** Retrieve DNS info from domains
- ***gr-aws*** Find S3 buckets for given domain (Work In Progress)
- ***gr-waf*** Identify which WAF is running on a domain
- ***gr-tech*** Identify technologies running on a URL (similar to wappalyzer)
- ***gr-filter*** Remove duplicated URLs, useless URLs (images, css...) and more from a list of endpoints
- ***gr-replace*** Replace given keyword or parameter value with provided value
- ***gr-secrets*** Search for API keys and leaked secrets in html and js pages
- ***gr-crawl*** Fastly crawl urls for gathering URLs, with custom depth and more options
- ***gr-whois*** Perform WHOIS query against domains

# Features

- Speed and concurrency
- Configurable via CLI arguments
- Easy usage
- Tools can be combined between them
- Output results to file in txt or JSON format
- Supports STDIN and STDOUT
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

Install official ***go-recon*** Golang package like this:

```sh
go get github.com/D3Ext/go-recon/pkg/go-recon
```

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

Usage of gr-secrets:
    -u)       url to search for secrets in (i.e. https://example.com/script.js)
    -l)       file containing a list of JS endpoints to search for secrets (one url per line)
    -r)       custom regex to search for (i.e. apikey=secret[a-z]+)
    -lr)      file containing a list of custom regex to search for (one regex per line)
    -w)       number of concurrent workers (default=15)
    -o)       output file to write secrets into
    -a)       user agent to include on requests (default=none)
    -c)       print colors on output
    -t)       milliseconds to wait before each request timeout (default=5000)
    -q)       don't print banner, only output
    -h)       print help panel

Examples:
    gr-secrets -u https://example.com -o secrets.txt -c
    gr-secrets -l js.txt -w 10 -t 6000
    gr-secrets -u https://example.com -lr regex.txt
    cat js.txt | gr-secrets -r "secret=api_[A-Z]+"
```

See [here](https://github.com/D3Ext/go-recon/blob/main/USAGE.md) for real usage examples during external recon

# Demo

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo1.png">

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo2.png">

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo3.png">

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo4.png">

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo5.png">

# API

If you want to use ***go-recon*** in your own Golang code see [here](https://github.com/D3Ext/go-recon/tree/main/examples)

# TODO

- More tools and features
- ~~Dockerfile~~
- HTML results reports
- More optimization
- Compare results with other tools such as **subfinder**, **gau**, **httprobe**...

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

# License

This project is under MIT license

Copyright © 2023, *D3Ext*

# Support

<a href="https://www.buymeacoffee.com/D3Ext" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-blue.png" alt="Buy Me A Coffee" style="height: 60px !important;width: 217px !important;" ></a>


