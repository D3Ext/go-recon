#!/bin/bash

#
# Author: D3Ext
# Github: https://github.com/D3Ext/go-recon
# Discord: @d3ext
# Twitter: @d3ext
# Blog: https://d3ext.github.io
#

# Detect if go-recon tools are installed
if [ "$(command -v gr-subdomains)" == "" ]; then
  echo -e "[-] go-recon tools not detected, ensure to have them installed"
  exit 0
fi

if [ "$1" == "" ]; then
  echo -e "[*] Usage: go-recon.sh example.com"
  exit 0
fi

domain="${1}"

if [ "$(echo "${domain}" | grep ".")" == "" ]; then
  echo -e "[*] Usage: go-recon.sh example.com"
  echo -e "[-] Invalid domain format!"
  exit 0
fi

mkdir "${domain}"

# find subdomains and save to file
echo -e "[*] Looking for subdomains..."
gr-subdomains -d "${domain}" -all -o "${domain}/subdomains.txt" -q

sleep 0.1

# probe active subdomains
echo -e "\n[*] Proving active subdomains..."
gr-probe -l "${domain}/subdomains.txt" -o "${domain}/active_subdomains.txt" -skip -q

sleep 0.1

# fetching urls from WaybackMachine api
echo -e "\n[*] Retrieving urls from WaybackMachine..."
gr-urls -d "${domain}" -o "${domain}/urls.txt" -q

sleep 0.1

# filter unique urls and also use some especific filters
echo -e "\n[*] Filtering urls..."
gr-filter -l "${domain}/urls.txt" -o "${domain}/filtered_urls.txt"

mkdir "${domain}/vuln_urls"
gr-filter -l "${domain}/filtered_urls.txt" -o "${domain}/vuln_urls/potential_xss.txt" -f "xss,nocontent"
gr-filter -l "${domain}/filtered_urls.txt" -o "${domain}/vuln_urls/potential_redirects.txt" -f "redirects" -p FUZZ
gr-filter -l "${domain}/filtered_urls.txt" -o "${domain}/vuln_urls/potential_sqli.txt" -f "sqli,nocontent"
gr-filter -l "${domain}/filtered_urls.txt" -o "${domain}/vuln_urls/potential_ssti.txt" -f "ssti,nocontent"
gr-filter -l "${domain}/filtered_urls.txt" -o "${domain}/vuln_urls/potential_ssrf.txt" -f "ssrf,nocontent"

sleep 0.1

# find potential 403 bypasses
echo -e "\n[*] Finding potential 403 bypasses"
gr-probe -l "${domain}/subdomains.txt" -fc 403 -skip -q | gr-403 -o "${domain}/403_bypass_urls.txt" -skip -c

sleep 0.1

# try to find potential open redirects
echo -e "\n[*] Finding potential open redirects..."
gr-openredirects -l "${domain}/vuln_urls/potential_redirects.txt" -o "${domain}/vuln_urls/vuln_redirects.txt"

sleep 0.1

# look for valid s3 buckets
echo -e "\n[*] Looking for valid AWS S3 buckets..."
gr-aws -d "${domain}" -o "${domain}/s3_buckets.txt" -c

sleep 0.1

# discovering .js endpoints
echo -e "\n[*] Retrieving JS endpoints..."
gr-crawl -l "${domain}/active_subdomains.txt" -d 3 -js -o "${domain}/js_endpoints.txt"

sleep 0.1

echo -e "\n[*] Crawling subdomains..."
gr-crawl -l "${domain}/active_subdomains.txt" -o "${domain}/crawl_urls.txt"

sleep 0.1

# looking for leaked secrets
echo -e "\n[*] Looking for leaked secrets..."
gr-secrets -l "${domain}/js_endpoints.txt" -o "${domain}/found_secrets.txt"

sleep 0.1

# getting DNS info
echo -e "\n[*] Getting DNS info..."
gr-dns -d "${domain}" -o "${domain}/dns_info.json" -c

sleep 0.1

# sending WHOIS query
echo -e "\n[*] Sending WHOIS query..."
gr-whois -d "${domain}" -o "${domain}/whois_info.json" -c

sleep 0.1

# identify WAFs
echo -e "\n[*] Identifying running WAFs..."
gr-waf -l "${domain}/active_subdomains.txt" -o "${domain}/wafs.txt" -c

sleep 0.1

echo -e "\n[+] Results stored under ${domain}/"
echo -e "[+] Subdomains written to ${domain}/subdomains.txt"
echo -e "[+] Active subdomains written to ${domain}/active_subdomains.txt"
echo -e "[+] Urls written to ${domain}/urls.txt"
echo -e "[+] Filtered urls written to ${domain}/filtered_urls.txt"
echo -e "[+] JS endpoints written to ${domain}/js_endpoints.txt"
echo -e "[*] Crawling urls written to ${domain}/crawl_urls.txt"
echo -e "[*] Potential vuln urls written under ${domain}/vuln_urls/"
echo -e "[+] 403 bypassed urls written to ${domain}/403_bypass_urls.txt"
echo -e "[+] Discovered secrets written to ${domain}/found_secrets.txt"
echo -e "[+] Discovered running WAFs written to ${domain}/wafs.csv"
echo -e "[+] Existing S3 buckets written to ${domain}/s3_buckets.txt"
echo -e "[+] DNS info written to ${domain}/dns_info.json"
echo -e "[+] WHOIS info written to ${domain}/whois_info.json"



