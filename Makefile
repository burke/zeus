MAKEFLAGS = -s

default: darwin

all: linux-386 linux-amd64 darwin

linux-386:
	cd cmd/zeus; CGO_ENABLED=0 GOOS=linux GOARCH=386 $(MAKE) -o zeus-linux-386

linux-amd64:
	cd cmd/zeus; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(MAKE) -o zeus-linux-amd64

darwin:
	cd cmd/zeus; CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(MAKE) -o zeus-darwin

clean:
	cd cmd/zeus; $(MAKE) clean
