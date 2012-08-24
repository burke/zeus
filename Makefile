MAKEFLAGS = -s

all:
	cd cmd/zeus; $(MAKE)

clean:
	cd cmd/zeus; $(MAKE) clean
