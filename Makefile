CC = go

all:
	mkdir build
	export CGO_ENABLED=0
	$(CC) build -o build/gr-subdomains tools/gr-subdomains.go
	$(CC) build -o build/gr-aws tools/gr-aws.go
	$(CC) build -o build/gr-urls tools/gr-urls.go
	$(CC) build -o build/gr-secrets tools/gr-secrets.go
	$(CC) build -o build/gr-waf tools/gr-waf.go
	$(CC) build -o build/gr-probe tools/gr-probe.go
	$(CC) build -o build/gr-dns tools/gr-dns.go
	$(CC) build -o build/gr-openredirects tools/gr-openredirects.go
	$(CC) build -o build/gr-crawl tools/gr-crawl.go
	$(CC) build -o build/gr-whois tools/gr-whois.go
	$(CC) build -o build/gr-replace tools/gr-replace.go
	$(CC) build -o build/gr-403 tools/gr-403.go
	$(CC) build -o build/gr-filter tools/gr-filter.go

install:
	install -m 0755 build/gr-subdomains $(DESTDIR)/usr/bin/gr-subdomains
	install -m 0755 build/gr-aws $(DESTDIR)/usr/bin/gr-aws
	install -m 0755 build/gr-urls $(DESTDIR)/usr/bin/gr-urls
	install -m 0755 build/gr-secrets $(DESTDIR)/usr/bin/gr-secrets
	install -m 0755 build/gr-waf $(DESTDIR)/usr/bin/gr-waf
	install -m 0755 build/gr-probe $(DESTDIR)/usr/bin/gr-probe
	install -m 0755 build/gr-dns $(DESTDIR)/usr/bin/gr-dns
	install -m 0755 build/gr-openredirects $(DESTDIR)/usr/bin/gr-openredirects
	install -m 0755 build/gr-crawl $(DESTDIR)/usr/bin/gr-crawl
	install -m 0755 build/gr-whois $(DESTDIR)/usr/bin/gr-whois
	install -m 0755 build/gr-403 $(DESTDIR)/usr/bin/gr-403
	install -m 0755 build/gr-filter $(DESTDIR)/usr/bin/gr-filter
	install -m 0755 build/gr-replace $(DESTDIR)/usr/bin/gr-replace

extra:
	mkdir -p ~/.config/go-recon
	cp -r utils/patterns/ ~/.config/go-recon/
	cp utils/autocompletion.bash ~/.config/go-recon/
	echo "source ~/.config/go-recon/autocompletion.bash" >> ~/.bashrc

clean:
	rm -rf build
