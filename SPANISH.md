<p align="center">
  <img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/transparent-banner.png" alt="Gopher"/>
  <h1 align="center">go-recon</h1>
  <h4 align="center">External recon toolkit</h4>
  <h6 align="center">Hecho con 💙 por D3Ext</h6>

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
  <a href="#introducción">Introducción</a> •
  <a href="#herramientas">Herramientas</a> •
  <a href="#instalación">Instalación</a> •
  <a href="#uso">Uso</a> •
  <a href="#contribuir">Contribuir</a>
</p>

# Introducción

Este proyecto comenzó como varios scripts en Golang para llevar a cabo diferentes procesos tediosos de forma automatica, para realizar el recon de forma externa, entre otras cosas mas. Con el tiempo fui puliendo las herramientas y finalmente decidí tomarmelo en serio, además de esta forma aprendería a usar los canales y la concurrencia en Golang, consiguiendo que las herramientas sean veloces y configurables.

Este toolkit proporciona diferentes herramientas para llevar a cabo reconocimiento externo. La mayoria de las funciones también estan disponibles para ser usadas, a traves de la API oficial del paquete. Sientete libre de contribuir reportando errores o dando ideas.

Mira la [Wiki](https://github.com/D3Ext/go-recon/wiki) para mas info

# Herramientas

- ***gr-subdomains***: Enumera los subdominios de un dominio mediante 8 proveedores diferentes (de forma pasiva)
- ***gr-urls***: Encuentra URLs de un dominio de diferentes fuentes (Wayback, AlienVault)
- ***gr-probe***: Prueba subdominios y urls activas (http y https) de forma veloz, concurrencia configurable y otras funciones
- ***gr-403***: Intenta "bypassear" paginas que devuelven codigo de estado 403 (forbidden)
- ***gr-openredirects***: Fuzzea por posibles open redirects en las URLs proporcionadas
- ***gr-dns***: Consigue información DNS de dominios
- ***gr-aws***: Enumera buckets S3 para un dominio/s mediante permutaciones, comprueba listas de buckets y mucho mas
- ***gr-waf***: Identifica si es posible el WAF que esta corriendo en una URL
- ***gr-filter***: Elimina URLs duplicadas e inutiles de una lista, aplica filtros, crea patrones personalizados y filtra output
- ***gr-replace***: Reemplaza palabras clave o el valor de un parametro por el valor proporcionado, en una lista de urls
- ***gr-secrets***: Busca API keys y secretos en paginas HTML y JS
- ***gr-crawl***: Realiza crawling de forma rápida para enumerar URLs y endpoints JS, con profundidad personalizada y otras opciones
- ***gr-whois***: Realiza consultas WHOIS a dominios

# Funciones

- Rapidez y concurrencia
- Fácil de usar y configurable mediante argumentos CLI
- Las herramientas se pueden combinar entre ellas
- Multiples formatos de output (STDOUT, TXT, JSON, CSV)
- Input como URL, dominio o STDIN
- Acceso directo a la API oficial del paquete
- Probado en Linux

# Instalación

Compila e instala el codigo fuente mediante Github:

```sh
git clone https://github.com/D3Ext/go-recon
cd go-recon
make
sudo make install
```

Los binarios será compilados e instalados en el PATH, por lo que solo tendrías que ejecutar las herramientas desde la CLI

```sh
$ gr-subdomains
```

## Extra

Si quieres instalar una lista de filtros/patrones personalizados y un script de autocompletado en Bash, ejecuta el siguiente comando:

```sh
make extra
```

Luego si presionas TAB dos veces al usar gr-subdomains o gr-filter, veras los proveedores y filtros disponibles

# Uso

Todas las herramientas se usan de forma similar y con los mismos parametros CLI

> Ejemplo de panel de ayuda
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

Mira [aqui](https://github.com/D3Ext/go-recon/blob/main/USAGE.md) para ver ideas y ejemplos reales como usar ***go-recon*** para reconocimiento externo

# Demo

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo1.png">

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo2.png">

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo3.png">

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo4.png">

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo5.png">

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo6.png">

<img src="https://raw.githubusercontent.com/D3Ext/go-recon/main/static/demo7.png">

# API

Instala el paquete de Golang oficial ***go-recon*** de esta forma:

```sh
go get github.com/D3Ext/go-recon/pkg/go-recon
```

Si quieres usar ***go-recon*** en tu propio codigo en Golang mira [aquí](https://github.com/D3Ext/go-recon/tree/main/examples)

# TODO

- ~~Parametro para controlar los proveedores (providers) utilizados~~
- ~~CSV output~~
- Mas herramientas y funciones
- ~~Dockerfile~~
- ~~Changelog~~
- Reportes de resultados en formato HTML
- Mas optimización
- ~~Comparar resultados con otras herramientas como **subfinder**, **gau**, **httprobe**...~~

# Referencias

Inspirado y motivado por herramientas increibles como estas:

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

Mira [CONTRIBUTING.md](https://github.com/D3Ext/go-recon/blob/main/CONTRIBUTING.md)

# Changelog

Mira [CHANGELOG.md](https://github.com/D3Ext/go-recon/blob/main/CHANGELOG.md)

# Licencia

Este proyecto está bajo licencia MIT

Copyright © 2023, *D3Ext*

# Soporte

<a href="https://www.buymeacoffee.com/D3Ext" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-blue.png" alt="Buy Me A Coffee" style="height: 60px !important;width: 217px !important;" ></a>


