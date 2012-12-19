# Reporting issues

I'd love it if you could use this template for bug reports, but it's not necessary:

```
## Description of Problem

One or two sentences giving an overview of the issue.

## System details

* **`uname -a`**: 

* **`ruby -v`**: 

* **`go version`**: (only if hacking on the go code)

## Steps to Reproduce

1) `zeus start` in a new rails project

2) `zeus ponies`

## Observed Behavior

* Ponies die

## Expected Behavior

* Ponies survive
```

# Hacking on Zeus

## Step 1: Prerequisites

You'll need ruby 1.9+ installed and `rake`, `bundler` and `ronn` gems are required to build the gem and manpages (`gem install rake bundler ronn`).

You will also need a working Go toolchain. You can get one from http://golang.org/doc/install.

To be able to fully-build Zeus, you will also need to install a couple additional Go runtimes. See http://dave.cheney.net/2012/09/08/an-introduction-to-cross-compilation-with-go. Zeus currently builds for darwin/amd64, linux/amd64, and linux/386.

## Step 2: Paths, etc.

You should check out this repository into `$GOPATH/src/github.com/burke/zeus`. Often `$GOPATH` will be set to `~/go`, but this is configurable. If you've just installed Go, you'll have to set this up yourself in your shell config. The Go site [has documentation on this](http://golang.org/doc/code.html).

## Step 3: Dependencies

cd into the zeus project directory and run `./dev-bootstrap.sh`. This will fetch and compile a couple libraries zeus uses for terminal interaction and such.

## Step 4: Building

### Context: How zeus is structured

The core of zeus is a single go program that acts as the coordinating process (master, e.g. `zeus start`), or the client (called per-command, e.g. `zeus client`). This code is cross-compiled for a handful of different architectures and bundled with a ruby gem. The ruby gem contains all the shim code necessary to boot a rails app under the control of the master process.

### Building

Just run `make`, basically. The `Makefile` should be an easy read. `make darwin` and related tasks will just build the go binaries to `./build`, while `make gem` will build the rubygem from the files currently in `./build`. `make all` does the whole shebang, but takes a few seconds to complete.

## Step 5: Contributing

Fork, branch, pullrequest! I'm sometimes really bad about responding to these in a timely fashion. Feel free to harass me on email or twitter if I'm not getting back to you.

## Questions?

If this doesn't work out for you, hit me up on twitter at @burkelibbey or email at burke@libbey.me

