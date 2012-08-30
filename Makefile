MAKEFLAGS = -s

default: darwin

all: linux-386 linux-amd64 darwin manpages gem

manpages:
	cd man; /usr/bin/env rake

gem:
	cd rubygem; /usr/bin/env rake

linux-386: goversion
	cd cmd/zeus; CGO_ENABLED=0 GOOS=linux GOARCH=386 $(MAKE) -o zeus-linux-386

linux-amd64: goversion
	cd cmd/zeus; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(MAKE) -o zeus-linux-amd64

darwin: goversion
	cd cmd/zeus; CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(MAKE) -o zeus-darwin

goversion:
	cd zeusversion ; /usr/bin/env ruby ./genversion.rb

clean:
	cd cmd/zeus; $(MAKE) clean
	cd man; rake clean
	cd rubygem ; rake clean
	rm -rf zeusversion
