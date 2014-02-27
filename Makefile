VERSIONFILE=go/zeusversion/zeusversion.go
GEM:=rubygem/pkg/zeus-$(shell cat VERSION).gem

.PHONY: default all clean binaries compileBinaries fmt install
default: all

all: fmt binaries man/build $(GEM)

binaries: build/zeus-linux-386 build/zeus-linux-amd64 build/zeus-darwin-amd64

fmt:
	find . -name '*.go' | xargs -t -I@ go fmt @

man/build: Gemfile.lock
	cd man && bundle exec rake

rubygem/pkg/%: \
	rubygem/ext/fsevents-wrapper/fsevents-wrapper \
	rubygem/man \
	rubygem/examples \
	rubygem/lib/zeus/version.rb \
	Gemfile.lock
	cd rubygem && bundle exec rake

rubygem/ext/fsevents-wrapper/fsevents-wrapper: ext/fsevents/build/Release/fsevents-wrapper
	mkdir -p $(@D)
	cp $< $@

rubygem/lib/zeus/version.rb:
	cd $(@D) && ./genversion.sh

rubygem/man: man/build
	mkdir -p $@
	cp -R $< $@

rubygem/build: binaries
	mkdir -p $@
	cp -R build/zeus-* $@

rubygem/examples: examples
	cp -r $< $@

ext/fsevents/build/Release/fsevents-wrapper:
	cd ext/fsevents && xcodebuild

build/zeus-%: $(VERSIONFILE) compileBinaries
	:
compileBinaries:
	gox -osarch="linux/386 linux/amd64 darwin/amd64" \
		-output="build/bin/zeus-{{.OS}}-{{.Arch}}" \
		github.com/burke/zeus/go/cmd/zeus

$(VERSIONFILE):
	cd $(@D) && ./genversion.sh

install: $(GEM)
	gem install $< --no-ri --no-rdoc

Gemfile.lock: Gemfile
	bundle check || bundle install

clean:
	rm -rf ext/fsevents/build man/build $(VERSIONFILE) rubygem/lib/zeus/version.rb rubygem/pkg




.PHONY: dev_bootstrap
dev_bootstrap:
	go get ./...
	bundle -v || gem install bundler --no-rdoc --no-ri
	bundle install
	go get github.com/mitchellh/gox
	gox -build-toolchain
