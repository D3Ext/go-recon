# Usage

This file contains some good examples about how to use the tools for external recon, some of them directly pipe output from one tool to another but you should also store the output in files so you don't have to repeat each process. If you have any question contact me via Discord and I'll answer you in less than 24h: ***@d3ext***

### How to use go-recon

#### Main parameters

All tools use CLI parameters with similar names between then, for example most tools allow single use when working with an unique domain via `-d` or `-domain` parameter (i.e. `gr-subdomains -d example.com`) and also allow multiple use when working with multiple domains via `-l` or `-list` parameter (i.e. `gr-subdomains -l domains.txt`) which expect the path to a file which is used as input (one domain/url per line)

#### STDIN

If you prefer to use the tools using STDIN (as a lot of people do) you also can, tools will autodetect that STDIN contains data and will take it as input so you could do something like this:

```sh
echo "example.com" | gr-subdomains
```

#### STDOUT

Tools have multiple output formats: STDOUT, TXT, JSON and CSV. The results will always be printed via STDOUT but you can also use `-o` (or `-output`) to save plain text output to a file (TXT) or `-oj` to save the results to a file in JSON format or `-oc` to do it again but this time in CSV format. Here you have a couple of examples when discovering subdomains:

> TXT example
```
support.hackerone.com
gslink.hackerone.com
design.hackerone.com
docs.hackerone.com
3d.hackerone.com
events.hackerone.com
resources.hackerone.com
hackerone.com
api.hackerone.com
www.hackerone.com
mta-sts.forwarding.hackerone.com
b.ns.hackerone.com
a.ns.hackerone.com
mta-sts.hackerone.com
mta-sts.managed.hackerone.com
links.hackerone.com
go.hackerone.com
info.hackerone.com
zendesk2.hackerone.com
zendesk3.hackerone.com
zendesk4.hackerone.com
zendesk1.hackerone.com
forwarding.hackerone.com
```

> JSON example
```
{
  "subdomains": [
    "support.hackerone.com",
    "gslink.hackerone.com",
    "design.hackerone.com",
    "docs.hackerone.com",
    "3d.hackerone.com",
    "events.hackerone.com",
    "resources.hackerone.com",
    "hackerone.com",
    "api.hackerone.com",
    "www.hackerone.com",
    "mta-sts.forwarding.hackerone.com",
    "b.ns.hackerone.com",
    "a.ns.hackerone.com",
    "mta-sts.hackerone.com",
    "mta-sts.managed.hackerone.com",
    "links.hackerone.com",
    "go.hackerone.com",
    "info.hackerone.com",
    "zendesk2.hackerone.com",
    "zendesk3.hackerone.com",
    "zendesk4.hackerone.com",
    "zendesk1.hackerone.com",
    "forwarding.hackerone.com"
  ],
  "length": 23,
  "time": "5.79s"
}
```

Anyways if you want to pipe STDOUT directly to other tools you could use `-q` (or `-quiet`) parameter which stands for "quiet" so in this way only the results will be printed without printing banner, logging info and other stuff.

> Example
```
gr-subdomains -d hackerone.com -quiet | gr-probe
```

### Examples and tool chains

#### Find subdomains and probe active ones

Check if domains are up, for example, if `example.com` is up, it will print https://example.com and http://example.com but if you only want to print the https one make sure to especify `-skip` parameter

```sh
gr-subdomains -d example.com -q | gr-probe -skip -q
```

#### Crawl a list of urls

***gr-crawl*** is a fast crawler with some options which can be truly helpful during target reconnaisance. By default it will crawl with a depth of 2 and with 10 workers, but it can be configured to accomplish your tasks.

> Only crawl urls inside path
```sh
gr-crawl -u https://example.com -path
```

> Only crawl JS files (useful)
```
gr-crawl -u https://example.com -js
```

#### Find secrets on JS files

***gr-secrets*** will look for all kind of leaked secrets on given urls by using regular expressions. It should be used with a list of HTML or JS endpoints.

A full example of automated use would be something like this:

```sh
echo "domain.com" | gr-subdomains -q | gr-probe -skip -q | gr-crawl -depth 1 -js -path -q | gr-secrets -c
```

#### Remove duplicates and useless urls with custom parameters

Some useful filters are preloaded on the tool itself they're `nocontent`, `hasparams`, `noparams`, `hasextension`, `noextension`. Other filters/patterns will be instaled under `~/.config/go-recon/patterns/`. They can be used with `-f` or `-filter` parameter.

```sh
gr-filter -l urls.txt # default filter (remove duplicates and useless urls)
gr-filter -l urls.txt -w json,zip # look for .json and .zip files (whitelist)
gr-filter -l urls.txt -b asp # exclude .asp extensions (blacklist)
gr-filter -l urls.txt -f nocontent # remove urls with human content (blogs, stories, articles...)
gr-filter -l urls.txt -params FUZZ # replace urls parameters values with given value (i.e. FUZZ)
```

#### Filter for specific vulns/patterns with custom templates

gr-replace use custom grep-based templates which are stored under `~/.config/go-recon/patterns/` by default go-recon comes with `xss.json`, `redirects.json`, `sqli.json`, `lfi.json`, `ssti.json`, `ssrf.json`, `rce.json`, `idor.json`, `takeovers.json`, `base64`, `ip`, `jwt`, `sqli_errors`, `s3_buckets`. See [utils/patterns/](https://github.com/D3Ext/go-recon/tree/main/utils/patterns) for all templates

This is an example template:

```json
{
  "description": "filter for parameters potentially vulnerable to LFI (Local File Inclusion)",
  "flags": "-iE",
  "patterns": [
    "file=",
    "document=",
    "folder=",
    "root=",
    "path=",
    "pg=",
    "style=",
    "pdf=",
    "template=",
    "php_path=",
    "doc=",
    "page=",
    "name=",
    "cat=",
    "dir=",
    "action=",
    "board=",
    "date=",
    "detail=",
    "download=",
    "prefix=",
    "include=",
    "inc=",
    "locate=",
    "show=",
    "site=",
    "type=",
    "view=",
    "content=",
    "layout=",
    "mod=",
    "conf=",
    "url="
  ]
}
```

Here are some good examples to take advantage of the templates:

```sh
gr-urls -d example.com -o urls.txt

gr-filter -l urls.txt -f xss # look for XSS vuln parameters on urls
gr-filter -l urls.txt -f sqli,nocontent # look for SQLi vuln parameters without human content
gr-filter -l urls.txt -f ssti -params FUZZ # look for SSTI vuln parameters and change param values to "FUZZ"
```

It can also be used with any other kind of output, just think about it like a grep on steroids which can do a lot of things.

```sh
curl -s -X GET "http://example.com" | gr-filter -f s3_buckets -c
```

#### Find open redirects

***gr-openredirects*** allows you to find open redirects in-mass or against an specific target. It can be configured to work in a variety of situations.

By default it will use a list of 149 payloads which is more useful against an unique target, but you can also use just 1 payload to test urls much faster.

Here are some examples:

> Common usage
```sh
gr-openredirects -u http://example.com/?param=FUZZ -c
```

> Different param placeholder
```sh
gr-openredirects -u http://example.com/?param=TEST -k TEST
```

> Only use 1 payload
```sh
gr-openredirects -u http://example.com/?param=FUZZ -skip
```

#### Discover valid S3 buckets

The tool uses official AWS S3 SDK for Golang. I highly recommend you to enable color on output via `-c` parameter since this way output will be much more readable. This tool has some different uses:

Check if a list of generated bucket names based on domain name exists, get bucket region, try to list their ACLs, list objects and more. For example if provided domain is `example.com`, this and more bucket names will be checked and enumerated:

```
a-example.com
a-example
example-a.com
example.com
example
account-example.com
example-account.com
account-example
example-a
example-com
example-account
admin-example.com
example-admin.com
admin-example
example-admin
administration-example.com

...
```

Execute `gr-aws -d example.com` for single domain or `gr-aws -l domains.txt` for a list of them. The amount of generated permutations can be configured via `-level` flag, it allows `1`, `2`, `3`, `4` and `5` as values. (i.e. `gr-aws -d example.com -level 2`)

</br>

This tool has other flag (`-bl` or `-bucket-list`), to directly check a list of bucket names. This way is really useful if you also work with other tools or any other thing. If you have a file with bucket names like this:

```
bucket-com
hackerone-files
github.com
admin-example.com

...
```

You can check if they exists and enumerate them like this: `gr-aws -bl buckets.txt`

Here are some extra generic examples:

```sh
gr-aws -d example.com -o found_buckets.txt -c # write existing buckets to file
gr-aws -d example.com -p perms.txt # using custom permutations (one perm per line)
gr-aws -bl buckets_to_check.txt -proxy http://127.0.0.1:8080 # using a proxy
gr-subdomains -d example.com -quiet | gr-aws -level 1 -c # Check common buckets for subdomains
```

#### Find open redirects

```sh
gr-urls -d example.com -q | gr-filter -f redirects,nocontent -params FUZZ -q | gr-openredirects -c -k FUZZ -o redirects.txt
```

#### Find potential 403 bypasses

```sh
gr-subdomains -d example.com -q | gr-probe -skip -fc 403 -q | gr-403 -skip -c
```

#### Find WAFs in mass

```sh
gr-subdomains -d hackerone.com -q | gr-probe -skip -q | gr-waf -hide -c
```

#### Enumerate running technologies in mass

```sh
gr-subdomains -d hackerone.com -q | gr-probe -techs -skip -c
```


