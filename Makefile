PACKAGE=github.com/burke/zeus
VERSION=$(shell cat VERSION)
GEM=rubygem/pkg/zeus-$(VERSION).gem
GEM_LINUX=rubygem-linux/pkg/zeus-$(VERSION).gem
GEMPATH=$(realpath rubygem)
VAGRANT_PLUGIN=vagrant/pkg/vagrant-zeus-$(VERSION).gem
VAGRANT_PLUGIN_LINUX=vagrant-linux/pkg/vagrant-zeus-$(VERSION).gem

CXX=g++
CXXFLAGS=-O3 -g -Wall

ifeq ($(shell uname -s),Darwin)
	TEST_LISTENER_DEP=ext/fsevents/build/Release/fsevents-wrapper
else
	TEST_LISTENER_DEP=ext/inotify-wrapper/inotify-wrapper
endif

.PHONY: default all clean binaries compileBinaries fmt install
default: all

all: fmt binaries man/build $(GEM) $(VAGRANT_PLUGIN)

binaries: build/zeus-linux-386 build/zeus-linux-amd64 build/zeus-darwin-amd64

linux: fmt linuxBinaries man/build $(GEM_LINUX) $(VAGRANT_PLUGIN_LINUX)

linuxBinaries: build-linux

fmt:
	find . -name '*.go' | xargs -t -I@ go fmt @

test-go: go/zeusversion/zeusversion.go $(TEST_LISTENER_DEP)
	ZEUS_LISTENER_BINARY=$(realpath $(TEST_LISTENER_DEP)) ZEUS_TEST_GEMPATH=$(GEMPATH) GO15VENDOREXPERIMENT=1 govendor test +local

man/build: Gemfile.lock
	cd man && ../bin/rake

rubygem-linux/pkg/%: \
	rubygem/man \
	rubygem/examples \
	rubygem/lib/zeus/version.rb \
	rubygem/build \
	rubygem/ext \
	Gemfile.lock
	cd rubygem && bundle install && bin/rake

rubygem/pkg/%: \
	rubygem/build/fsevents-wrapper \
	rubygem/man \
	rubygem/examples \
	rubygem/lib/zeus/version.rb \
	rubygem/build \
	rubygem/ext \
	Gemfile.lock
	cd rubygem && bundle install && bin/rake

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

rubygem/ext: \
    ext/inotify-wrapper/inotify-wrapper.cpp \
    ext/inotify-wrapper/extconf.rb \
    ext/file-listener/file-listener.cpp \
    ext/file-listener/extconf.rb
	rm -rf $@
	mkdir -p $@
	cp -r ext/inotify-wrapper ext/file-listener $@

vagrant/pkg/%: \
	vagrant/build/fsevents-wrapper \
	vagrant/lib/vagrant-zeus/version.rb \
	vagrant/ext \
	Gemfile.lock
	cd vagrant && bundle install && bundle exec rake

vagrant-linux/pkg/%: \
	vagrant/lib/vagrant-zeus/version.rb \
	vagrant/ext \
	Gemfile.lock
	cd vagrant && bundle install && bundle exec rake

vagrant/build/fsevents-wrapper: ext/fsevents/build/Release/fsevents-wrapper
	mkdir -p $(@D)
	cp $< $@

vagrant/ext: \
    ext/inotify-wrapper/inotify-wrapper.cpp \
    ext/inotify-wrapper/extconf.rb
	rm -rf $@
	mkdir -p $@
	cp -r ext/inotify-wrapper $@

ext/fsevents/build/Release/fsevents-wrapper:
	cd ext/fsevents && xcodebuild

# This compilation is only used for testing under linux and is
# not packaged into the final gem.
ext/inotify-wrapper/inotify-wrapper: ext/inotify-wrapper/inotify-wrapper.o
	$(CXX) $(CXXFLAGS) $< -o $@

%.o: %.cpp
	$(CXX) $(CXXFLAGS) -c $< -o $@

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

go/zeusversion/zeusversion.go:
	mkdir -p $(@D)
	@echo 'package zeusversion\n\nconst VERSION string = "$(VERSION)"' > $@
rubygem/lib/zeus/version.rb:
	mkdir -p $(@D)
	@echo 'module Zeus\n  VERSION = "$(VERSION)"\nend' > $@
vagrant/lib/vagrant-zeus/version.rb:
	mkdir -p $(@D)
	@echo 'module VagrantPlugins\n  module Zeus\n    VERSION = "$(VERSION)"\n  end\nend' > $@


install: $(GEM)
	gem install $< --no-ri --no-rdoc

Gemfile.lock: Gemfile
	bundle check || bundle install

clean:
	rm -rf ext/fsevents/build man/build go/zeusversion build
	rm -rf rubygem/{man,build,pkg,examples,ext,lib/zeus/version.rb,MANIFEST}
	rm -rf vagrant/{build,pkg,ext,lib/vagrant-zeus/version.rb,MANIFEST}




.PHONY: dev_bootstrap
dev_bootstrap: go/zeusversion/zeusversion.go
	bundle -v || gem install bundler --no-rdoc --no-ri
	bundle install
	go get github.com/mitchellh/gox github.com/kardianos/govendor
