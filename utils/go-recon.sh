#!/bin/bash

#
# Author: D3Ext
# Github: https://github.com/D3Ext/go-recon
# Discord: @d3ext
# Twitter: @d3ext
# Blog: https://d3ext.github.io
#

# Assumming go-recon tools are installed

if [ "$(command -v gr-subdomains)" == "" ]; then
  echo -e "[-] go-recon tools not detected, ensure to have them installed"
  exit 0
fi

if [ "$1" == "" ]; then
  echo -e "[*] Usage: go-recon.sh example.com"
  exit 0
fi

domain="${1}"

mkdir "${domain}"

# find subdomains and save to file
echo -e "[*] Looking for subdomains..."
gr-subdomains -d "${domain}" -o "${domain}/subdomains.txt" -q

sleep 0.1

# probe active subdomains
echo -e "\n[*] Proving active subdomains..."
gr-probe -l "${domain}/subdomains.txt" -o "${domain}/active_subdomains.txt" -x -q

sleep 0.1

# fetching urls from WaybackMachine api
echo -e "\n[*] Retrieving urls from WaybackMachine..."
gr-urls -d "${domain}" -o "${domain}/urls.txt" -q

sleep 0.1

# filter unique urls and also use some especific filters
echo -e "\n[*] Filtering urls..."
gr-filter -l "${domain}/urls.txt" -o "${domain}/filtered_urls.txt" -q
gr-filter -l "${domain}/urls.txt" -o "${domain}/potential_vuln_urls.txt" -f "vuln,nocontent" -q
gr-filter -l "${domain}/urls.txt" -o "${domain}/potential_redirects_urls.txt" -f redirects -q

sleep 0.1

# find potential 403 bypasses
echo -e "\n[*] Finding potential 403 bypasses"
gr-probe -l "${domain}/subdomains.txt" -sc 403 -x -q | gr-403 -o "${domain}/403_bypass_urls.txt" -q

sleep 0.1

# try to find potential open redirects
echo -e "\n[*] Finding potential open redirects..."
gr-openredirects -l "${domain}/potential_redirects_urls.txt" -o "${domain}/open_redirects_urls.txt" -q

sleep 0.1

# look for valid s3 buckets
echo -e "\n[*] Looking for valid AWS S3 buckets..."
gr-aws -d "${domain}" -o "${domain}/s3_buckets.txt"

sleep 0.1

# discovering .js endpoints
echo -e "\n[*] Retrieving JS endpoints..."
gr-js -l "${domain}/active_subdomains.txt" -o "${domain}/js_endpoints.txt" -q

sleep 0.1

# looking for leaked secrets
echo -e "\n[*] Looking for leaked secrets..."
gr-secrets -l "${domain}/js_endpoints.txt" -o "${domain}/found_secrets.txt" -q

sleep 0.1

# getting DNS info
echo -e "\n[*] Getting DNS info..."
gr-dns -d "${domain}" -o "${domain}/dns_info.txt"

sleep 0.1

# sending WHOIS query
echo -e "\n[*] Sending WHOIS query..."
gr-whois -d "${domain}" -o "${domain}/whois_info.txt" -q

sleep 0.1

# identify WAFs
echo -e "\n[*] Identifying running WAFs..."
gr-waf -l "${domain}/active_subdomains.txt" -o "${domain}/wafs.txt"

sleep 0.1

echo -e "\n[+] Results stored under ${domain}/"
echo -e "[+] Subdomains written to ${domain}/subdomains.txt"
echo -e "[+] Active subdomains written to ${domain}/active_subdomains.txt"
echo -e "[+] Urls written to ${domain}/urls.txt"
echo -e "[+] Filtered urls written to ${domain}/filtered_urls.txt"
echo -e "[+] JS endpoints written to ${domain}/js_endpoints.txt"
echo -e "[+] Found open redirects written to ${domain}/open_redirects_urls.txt"
echo -e "[+] 403 bypassed urls written to ${domain}/403_bypass_urls.txt"
echo -e "[+] Discovered secrets written to ${domain}/found_secrets.txt"
echo -e "[+] WAFs written to ${domain}/wafs.txt"
echo -e "[+] S3 Buckets written to ${domain}/s3_buckets.txt"
echo -e "[+] DNS info written to ${domain}/dns_info.txt"
echo -e "[+] WHOIS info written to ${domain}/whois_info.txt"



