# Zeus

## What?

Zeus preloads your app so that your normal development tasks such as `console`, `server`, `generate`, and tests are faster.

[Mediocre screencast](http://burke.libbey.me/zeus.mov). Better one coming soon.

## Why?

Because waiting 25 seconds sucks, but waiting 0.4 seconds doesn't.

## Requirements

Pretty specific:

* OS X 10.7+
* Ruby 1.9+
* Backported GC from Ruby 2.0.

You can install the GC-patched ruby from [this gist](https://gist.github.com/1688857) or from RVM.

## Installation

Install the gem.

    gem install zeus

Run the project initializer from the root of your rails application:

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


## Contributing

Fork, Branch, Pull Request.

## Thanks...

* To [Jesse Storimer](http://github.com/jstorimer) for spin, part of the inspiration for this project
* To [Samuel Kadolph](http://github.com/samuelkadolph) for doing most of the cross-process pseudoterminal legwork.
* To [Stefan Penner](http://github.com/stefanpenner) for discussion and various contributions.
* To [Shopify](http://github.com/Shopify) for letting me spend (some of) my days working on this.

## Doesn't work for you?

Try these libraries instead:

* [spin](https://github.com/jstorimer/spin)
* [spork](https://github.com/sporkrb/spork)
* [guard](https://github.com/guard/guard)
