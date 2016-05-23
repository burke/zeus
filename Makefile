PACKAGE=github.com/burke/zeus
VERSION=$(shell cat VERSION)
GEM=rubygem/pkg/zeus-$(VERSION).gem
GEM_LINUX=rubygem-linux/pkg/zeus-$(VERSION).gem
GEMPATH=$(realpath rubygem)
VAGRANT_PLUGIN=vagrant/pkg/vagrant-zeus-$(VERSION).gem
VAGRANT_PLUGIN_LINUX=vagrant-linux/pkg/vagrant-zeus-$(VERSION).gem

CXX=g++
CXXFLAGS=-O3 -g -Wall

.PHONY: default all clean binaries compileBinaries fmt install
default: all

all: fmt binaries man/build $(GEM) $(VAGRANT_PLUGIN)

binaries: build/zeus-linux-386 build/zeus-linux-amd64 build/zeus-darwin-amd64

linux: fmt linuxBinaries man/build $(GEM_LINUX) $(VAGRANT_PLUGIN_LINUX)

linuxBinaries: build-linux

fmt:
	govendor fmt +local

test-go: go/zeusversion/zeusversion.go
	ZEUS_TEST_GEMPATH=$(GEMPATH) GO15VENDOREXPERIMENT=1 govendor test +local

man/build: Gemfile.lock
	cd man && ../bin/rake

rubygem-linux/pkg/%: \
	rubygem/man \
	rubygem/examples \
	rubygem/lib/zeus/version.rb \
	rubygem/build \
	Gemfile.lock
	cd rubygem && bundle install && bin/rake

rubygem/pkg/%: \
	rubygem/man \
	rubygem/examples \
	rubygem/lib/zeus/version.rb \
	rubygem/build \
	Gemfile.lock
	cd rubygem && bundle install && bin/rake

rubygem/man: man/build
	mkdir -p $@
	cp -R $< $@

rubygem/build: binaries
	mkdir -p $@
	cp -R build/zeus-* $@

rubygem/examples: examples
	rm -rf $@
	cp -r $< $@

vagrant/pkg/%: \
	vagrant/lib/vagrant-zeus/version.rb \
	vagrant/build/fsevents-wrapper \
	$(wildcard vagrant/ext/inotify-wrapper/*) \
	Gemfile.lock
	cd vagrant && bundle install && bundle exec rake

vagrant-linux/pkg/%: \
	vagrant/lib/vagrant-zeus/version.rb \
	$(wildcard vagrant/ext/inotify-wrapper/*) \
	Gemfile.lock
	cd vagrant && bundle install && bundle exec rake

vagrant/build/fsevents-wrapper: vagrant/ext/fsevents/build/Release/fsevents-wrapper
	mkdir -p $(@D)
	cp $< $@

vagrant/ext/fsevents/build/Release/fsevents-wrapper: vagrant/ext/fsevents/main.m
	cd vagrant/ext/fsevents && xcodebuild

build/zeus-%: go/zeusversion/zeusversion.go compileBinaries
	@:
compileBinaries:
	GO15VENDOREXPERIMENT=1 gox -osarch="linux/386 linux/amd64 darwin/amd64" \
		-output="build/zeus-{{.OS}}-{{.Arch}}" \
		$(PACKAGE)/go/cmd/zeus

build-linux: go/zeusversion/zeusversion.go compileLinuxBinaries
	@:
compileLinuxBinaries:
	GO15VENDOREXPERIMENT=1 gox -osarch="linux/386 linux/amd64" \
		-output="build/zeus-{{.OS}}-{{.Arch}}" \
		$(PACKAGE)/go/cmd/zeus

go/zeusversion/zeusversion.go: VERSION
	mkdir -p $(@D)
	@echo 'package zeusversion\n\nconst VERSION string = "$(VERSION)"' > $@
rubygem/lib/zeus/version.rb: VERSION
	mkdir -p $(@D)
	@echo 'module Zeus\n  VERSION = "$(VERSION)"\nend' > $@
vagrant/lib/vagrant-zeus/version.rb: VERSION
	mkdir -p $(@D)
	@echo 'module VagrantPlugins\n  module Zeus\n    VERSION = "$(VERSION)"\n  end\nend' > $@


install: $(GEM)
	gem install $< --no-ri --no-rdoc

Gemfile.lock: Gemfile
	bundle check || bundle install

clean:
	rm -rf vagrant/ext/fsevents/build man/build go/zeusversion build
	rm -rf rubygem/{man,build,pkg,examples,lib/zeus/version.rb,MANIFEST}
	rm -rf vagrant/{build,pkg,lib/vagrant-zeus/version.rb,MANIFEST}




.PHONY: dev_bootstrap
dev_bootstrap: go/zeusversion/zeusversion.go
	bundle -v || gem install bundler --no-rdoc --no-ri
	bundle install
	go get github.com/mitchellh/gox github.com/kardianos/govendor
