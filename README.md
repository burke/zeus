# Zeus

## What?

Zeus preloads your app so that your normal development tasks such as `console`, `server`, `generate`, and tests are faster.

[Mediocre screencast](http://burke.libbey.me/zeus.mov). Better one coming soon.

## Why?

Because waiting 25 seconds sucks, but waiting 0.4 seconds doesn't.

## When?

Soon? You can use Zeus now, but don't expect it to be perfect. I'm working hard on it.

## Ugly bits

* Not battle-tested yet

## Installation

Install the gem.

    gem install zeus

Run the project initializer.

    zeus init

## Upgrading from initial release

Since zeus is super-brand-new, the config file format changed already.

    gem install zeus
    rm .zeus.rb
    zeus init

## Usage

Start the server:

    zeus start

Run some commands:

    zeus console
    zeus server
    zeus testrb -Itest -I. test/unit/omg_test.rb
    zeus generate model omg
    zeus rake -T
    zeus runner omg.rb

## TODO (roughly prioritized)

* Make sure client connection requests are handled immediately
* Acceptors booting should not be dependent on passing all loaded features to the file monitor
* Route all logging output through Zeus.ui
* Handle connections for not-yet-started sockets
* Refactor, refactor, refactor...
* Support other frameworks?
* Figure out how to run full test suites without multiple env loads
* Don't replace a socket with changed deps until the new one is ready

## Ideas (not quite TODOs)

* (maybe) Start the preloader as a daemon transparently when any command is run, then wait for it to finish
* Support inotify on linux

## Contributing

Fork, Branch, Pull Request.

## Thanks...

* To [Jesse Storimer](http://github.com/jstorimer) for spin, part of the inspiration for this project
* To [Samuel Kadolph](http://github.com/samuelkadolph) for doing most of the cross-process pseudoterminal legwork.
* To [Shopify](http://github.com/Shopify) for letting me spend (some of) my days working on this.

## Doesn't work for you?

Try these libraries instead:

* [spin](https://github.com/jstorimer/spin)
* [spork](https://github.com/sporkrb/spork)
* [guard](https://github.com/guard/guard)
