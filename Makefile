PACKAGE=github.com/burke/zeus
VERSION=$(shell cat VERSION)
GEM=rubygem/pkg/zeus-$(VERSION).gem

.PHONY: default all clean binaries compileBinaries fmt install
default: all

all: fmt binaries man/build $(GEM)

binaries: build/zeus-linux-386 build/zeus-linux-amd64 build/zeus-darwin-amd64

fmt:
	find . -name '*.go' | xargs -t -I@ go fmt @

man/build: Gemfile.lock
	cd man && bundle exec rake

rubygem/pkg/%: \
	rubygem/build/fsevents-wrapper \
	rubygem/man \
	rubygem/examples \
	rubygem/lib/zeus/version.rb \
	rubygem/build \
	Gemfile.lock
	cd rubygem && bundle exec rake

rubygem/build/fsevents-wrapper: ext/fsevents/build/Release/fsevents-wrapper
	mkdir -p $(@D)
	cp $< $@

rubygem/man: man/build
	mkdir -p $@
	cp -R $< $@

rubygem/build: binaries
	mkdir -p $@
	cp -R build/zeus-* $@

rubygem/examples: examples
	rm -rf $@
	cp -r $< $@

ext/fsevents/build/Release/fsevents-wrapper:
	cd ext/fsevents && xcodebuild

build/zeus-%: go/zeusversion/zeusversion.go compileBinaries
	@:
compileBinaries:
	gox -osarch="linux/386 linux/amd64 darwin/amd64" \
		-output="build/zeus-{{.OS}}-{{.Arch}}" \
		$(PACKAGE)/go/cmd/zeus

go/zeusversion/zeusversion.go:
	mkdir -p $(@D)
	@echo 'package zeusversion\n\nconst VERSION string = "$(VERSION)"' > $@
rubygem/lib/zeus/version.rb:
	mkdir -p $(@D)
	@echo 'module Zeus\n  VERSION = "$(VERSION)"\nend' > $@


install: $(GEM)
	gem install $< --no-ri --no-rdoc

Gemfile.lock: Gemfile
	bundle check || bundle install

clean:
	rm -rf ext/fsevents/build man/build go/zeusversion build
	rm -rf rubygem/{man,build,pkg,examples,lib/zeus/version.rb,ext/fsevents-wrapper,MANIFEST}




.PHONY: dev_bootstrap
dev_bootstrap:
	go get ./...
	bundle -v || gem install bundler --no-rdoc --no-ri
	bundle install
	go get github.com/mitchellh/gox
	gox -build-toolchain
