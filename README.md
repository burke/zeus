# Zeus

Zeus preloads your Rails app so that your normal development tasks such as `console`, `server`, `generate`, and specs/tests take **less than one second**.

This screencast gives a quick overview of how to use zeus with Rails.

[![Watch the screencast!](http://s3.amazonaws.com/burkelibbey/vimeo-zeus.png)](http://vimeo.com/burkelibbey/zeus)

More technically speaking, Zeus is a language-agnostic application checkpointer for non-multithreaded applications. Currently only ruby is targeted, but explicit support for other languages is on the horizon.

## Requirements (for use with Rails)

* OS X 10.7+ *OR* Linux 2.6.13+
* Rails 3.0+ (Support for other versions is not difficult and is planned.)
* Ruby 1.9.3+ with backported GC from Ruby 2.0 *OR* Rubinius

You can install the GC-patched ruby from [this gist](https://gist.github.com/1688857) or from RVM. This is not actually 100% necessary, especially if you have a lot of memory. Feel free to give it a shot first without, but if you're suddenly out of RAM, switching to the GC-patched ruby will fix it.

## Installation

Install the gem.

    gem install zeus
    zeus init

Q: "I should put it in my `Gemfile`, right?"

A: You can, but running `bundle exec zeus` instead of `zeus` can add precious seconds to a command that otherwise would be quite a bit faster. Zeus was built to be run from outside of bundler.

## Usage

Start the server:

    zeus start

See a list of the available commands:

    zeus commands

Run some commands:

    zeus console
    zeus server
    zeus testrb test/unit/widget_test.rb
    zeus rspec spec/widget_spec.rb
    zeus generate model omg
    zeus rake -T
    zeus runner omg.rb

## Hacking

To add/modify commands, see [`docs/ruby/modifying.md`](/burke/zeus/tree/master/docs/ruby/modifying.md).

To get started hacking on Zeus itself, see [`docs/overview.md`](/burke/zeus/tree/master/docs/overview.md).
