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

## Version 0.2.0 TODO

* Make sure client connection requests are handled immediately (is this a problem? maybe not. Could chunk the select though)
* Zeus should "know" about all the acceptors it's handling before they boot.
* When a connection is requested to an unbooted acceptor, indicate that it's still loading.
* When a syntax error happens while a stage is booting, causing it to crash, connections to its descendent acceptors should print that message then terminate.
* After a syntax error causes failed load, a stage should re-attempt each time the file changes until it succeeds.
* command processes should not be killed when zeus reloads files. (should they be killed when zeus exits?)
* Make sure that when a command process's connection is dropped, it is killed

## Future TODO (roughly prioritized)

* Refactor, refactor, refactor...
* Support other frameworks?
* Figure out how to run full test suites without multiple env loads

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
