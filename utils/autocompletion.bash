#!/usr/bin/env bash

complete -W "$(gr-subdomains -lp -q)" gr-subdomains
complete -W "$(gr-filter -lf -q)" gr-filter
