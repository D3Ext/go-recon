# Usage

This file contains some good examples about how to use the tools for external recon, some of them directly pipe output from one tool to another but you should also store the output in files so you don't have to repeat each process

- Find subdomains and probe active ones

```sh
gr-subdomains -d example.com -q | gr-probe -x -q
```

- Find secrets on JS files

```sh
cat urls.txt | gr-js -q | gr-secrets -c
```

- Discover and filter urls

```sh
gr-urls -d example.com -o urls.txt

# 4 different examples
cat urls.txt | gr-filter # remove duplicates, useless urls and more
cat urls.txt | gr-filter -w json,zip # look for .json and .zip files
cat urls.txt | gr-filter -f nocontent # remove urls with human content (blogs, stories, articles...)
cat urls.txt | gr-filter -f vuln # look for potential vulnerable parameters on urls
```

- Discover valid S3 buckets

```sh
gr-aws -d example.com
gr-aws -d example.com -p perms.txt # using custom permutations (one perm per line)
gr-aws -b buckets_to_check.txt # directly using a list of bucket names to check them
gr-subdomains -d example.com -q | gr-aws -c 
```

- Find open redirects

```sh
gr-urls -d example.com -q | gr-filter -f redirects,nocontent -q | gr-openredirects -c -o redirects.txt
```

- Find potential 403 bypasses

```sh
gr-subdomains -d example.com -q | gr-probe -x -q -sc 403 | gr-403 -c
```

- Find WAFs in mass

```sh
gr-subdomains -d hackerone.com -q | gr-probe -x -q | gr-waf -c
```

- Enumerate running technologies in mass

```sh
gr-subdomains -d hackerone.com -q | gr-probe -x -q | gr-tech -c
```


