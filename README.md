# Zeus [![Build Status](https://secure.travis-ci.org/burke/zeus.png?branch=master)](http://travis-ci.org/burke/zeus) [![Code Climate](https://codeclimate.com/badge.png)](https://codeclimate.com/github/burke/zeus)

## What is Zeus?

Zeus preloads your app so that your normal development tasks such as `console`, `server`, `generate`, and specs/tests take **less than one second**.

This screencast gives a quick overview of how to use zeus:

[![Watch the screencast!](http://s3.amazonaws.com/burkelibbey/vimeo-zeus.png)](http://vimeo.com/burkelibbey/zeus)

## Requirements

Pretty specific:

* OS X 10.7+
* Ruby 1.9+
* Rails 3.0+ (Support for other versions is not difficult and is planned.)
* Backported GC from Ruby 2.0.

You can install the GC-patched ruby from [this gist](https://gist.github.com/1688857) or from RVM.  This is not actually 100% necessary, especially if you have a lot of memory. Feel free to give it a shot first without, but if you're suddenly out of RAM, switching to the GC-patched ruby will fix it.

## Installation

Install the gem.

    gem install zeus

Q: "I should put it in my `Gemfile`, right?"

A: No. running `bundle exec zeus` instead of `zeus` can add precious seconds to a command that otherwise would take 200ms. Zeus was built to be run from outside of bundler.

## Usage

Start the server:

    zeus start

Run some commands:

    zeus console
    zeus server
    zeus testrb test/unit/widget_test.rb
    zeus rspec spec/widget_spec.rb
    zeus generate model omg
    zeus rake -T
    zeus runner omg.rb


## Contributing

Fork, Branch, Pull Request.

## Thanks...

* To [Stefan Penner](http://github.com/stefanpenner) for discussion and various contributions.
* To [Samuel Kadolph](http://github.com/samuelkadolph) for doing most of the cross-process pseudoterminal legwork.
* To [Jesse Storimer](http://github.com/jstorimer) for spin, part of the inspiration for this project
* To [Shopify](http://github.com/Shopify) for letting me spend (some of) my days working on this.

## Doesn't work for you?

Try these libraries instead:

* [spin](https://github.com/jstorimer/spin)
* [spork](https://github.com/sporkrb/spork)
* [guard](https://github.com/guard/guard)
