PACKAGE=github.com/burke/zeus
VERSION=$(shell cat VERSION)
GEM=rubygem/pkg/zeus-$(VERSION).gem
GEMPATH=$(realpath rubygem)
VAGRANT_PLUGIN=vagrant/pkg/vagrant-zeus-$(VERSION).gem
VAGRANT_WRAPPERS=$(wildcard vagrant/ext/inotify-wrapper/*)
BINARIES=zeus-linux-386 zeus-linux-amd64
GO_SRC=$(shell find go -name '*.go')
GEM_SRC=$(shell find rubygem -name '*.rb')
VAGRANT_SRC=$(shell find vagrant -name '*.rb')

ifeq ($(shell uname -s),Darwin)
	VAGRANT_WRAPPERS += vagrant/build/fsevents-wrapper
	BINARIES += zeus-darwin-amd64
endif

CXX=g++
CXXFLAGS=-O3 -g -Wall

.PHONY: default all clean fmt test test-go test-rubygem install mod-download bundler
default: all

all: test fmt man/build $(GEM) $(VAGRANT_PLUGIN)

mod-download:
	go mod download

fmt:
	go fmt ./...

test: test-go test-rubygem

test-go: go/zeusversion/zeusversion.go rubygem/lib/zeus/version.rb mod-download
	ZEUS_TEST_GEMPATH="$(GEMPATH)" go test ./...

test-rubygem: rubygem/lib/zeus/version.rb rubygem/Gemfile.lock
	cd rubygem && bin/rspec

man/build: Gemfile.lock
	cd man && ../bin/rake

rubygem/pkg/%: \
	rubygem/man \
	rubygem/examples \
	rubygem/lib/zeus/version.rb \
	rubygem/Gemfile.lock \
	$(GEM_SRC) \
	$(addprefix rubygem/build/,$(BINARIES))
	cd rubygem && bundle install && bin/rake

rubygem/man: man/build
	mkdir -p $@
	cp -R $< $@

rubygem/examples: examples
	rm -rf $@
	cp -r $< $@

vagrant/pkg/%: vagrant/lib/vagrant-zeus/version.rb $(VAGRANT_WRAPPERS) $(VAGRANT_SRC) vagrant/Gemfile.lock
	cd vagrant && bundle install && bundle exec rake

vagrant/build/fsevents-wrapper: vagrant/ext/fsevents/build/Release/fsevents-wrapper
	mkdir -p $(@D)
	cp $< $@

vagrant/ext/fsevents/build/Release/fsevents-wrapper: vagrant/ext/fsevents/main.m
	cd vagrant/ext/fsevents && xcodebuild

rubygem/build/zeus-%: go/zeusversion/zeusversion.go install-gox $(GO_SRC)
	mkdir -p rubygem/build
	gox -osarch="$(subst -,/,$*)" \
		-output="rubygem/build/zeus-{{.OS}}-{{.Arch}}" \
		$(PACKAGE)/go/cmd/zeus

go/zeusversion/zeusversion.go: VERSION
	mkdir -p $(@D)
	@echo 'package zeusversion\n\nconst VERSION string = "$(VERSION)$(GO_VERSION_SUFFIX)"' > $@
rubygem/lib/zeus/version.rb: VERSION
	mkdir -p $(@D)
	@echo 'module Zeus\n  VERSION = "$(VERSION)"\nend' > $@
vagrant/lib/vagrant-zeus/version.rb: VERSION
	mkdir -p $(@D)
	@echo 'module VagrantPlugins\n  module Zeus\n    VERSION = "$(VERSION)"\n  end\nend' > $@


install: $(GEM)
	gem install $< --no-ri --no-rdoc

%/Gemfile.lock: $*Gemfile bundler
	cd $* && bundle check || bundle install

Gemfile.lock: Gemfile bundler
	bundle check || bundle install

install-gox:
	go install github.com/mitchellh/gox@latest

clean:
	rm -rf vagrant/ext/fsevents/build man/build go/zeusversion
	rm -rf rubygem/{man,build,pkg,examples,lib/zeus/version.rb,MANIFEST}
	rm -rf vagrant/{build,pkg,lib/vagrant-zeus/version.rb,MANIFEST}

bundler:
	bundle -v || gem install bundler --no-rdoc --no-ri
