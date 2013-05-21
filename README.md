# Zeus
[![Build Status](https://travis-ci.org/burke/zeus.png?branch=master)](https://travis-ci.org/burke/zeus)

Zeus preloads your Rails app so that your normal development tasks such as `console`, `server`, `generate`, and specs/tests take **less than one second**.

This screencast gives a quick overview of how to use zeus with Rails.

[![Watch the screencast!](http://s3.amazonaws.com/burkelibbey/vimeo-zeus.png)](http://vimeo.com/burkelibbey/zeus)

Zeus is also covered in [RailsCasts episode 412](http://railscasts.com/episodes/412-fast-rails-commands)

More technically speaking, Zeus is a language-agnostic application checkpointer for non-multithreaded applications. Currently only ruby is targeted, but explicit support for other languages is on the horizon.

## Requirements (for use with Rails)

* OS X 10.7+ *OR* Linux 2.6.13+
* Rails 3.x
* Ruby 1.9.3+ with backported GC from Ruby 2.0 *OR* Rubinius

You can install the GC-patched ruby from [this gist for rbenv](https://gist.github.com/1688857) or [this gist for RVM](https://gist.github.com/4136373). This is not actually 100% necessary, especially if you have a lot of memory. Feel free to give it a shot first without, but if you're suddenly out of RAM, switching to the GC-patched ruby will fix it.

*Please note*: Zeus requires your project to be running on a file system that supports FSEvents or inotify.

## Installation

Install the gem.

    gem install zeus

Q: "I should put it in my `Gemfile`, right?"

A: No. You can, but running `bundle exec zeus` instead of `zeus` adds precious seconds to commands that otherwise would be quite a bit faster. Zeus was built to be run from outside of bundler.

## Usage

Start the server:

    zeus start

The server will print a list of available commands.

Run some commands in another shell:

    zeus console
    zeus server
    zeus test test/unit/widget_test.rb
    zeus test spec/widget_spec.rb
    zeus generate model omg
    zeus rake -T
    zeus runner omg.rb

## Related gems

* [Spork](https://github.com/sporkrb/spork) - a [DRb server](http://www.ruby-doc.org/stdlib-1.9.3/libdoc/drb/rdoc/DRb.html) that forks before each run to ensure a clean testing state
* [Commands](https://github.com/rails/commands) - a persistent console that runs Rails commands without reloading the env
* [Spring](https://github.com/jonleighton/spring) - like Zeus but in pure Ruby, totally automatic, alpha and limited compatibility

If you're switching from Spork, be sure to [read the wiki page on Spork](https://github.com/burke/zeus/wiki/Spork).


## Hacking

To add/modify commands, see [`docs/ruby/modifying.md`](docs/ruby/modifying.md).

To get started hacking on Zeus itself, see [`docs/overview.md`](docs/overview.md).

See also the handy contribution guide at [`contributing.md`](contributing.md).

## Alternative plans

The default plan bundled with zeus only supports Rails 3.x. There is a project (currently WIP) to provide Rails 2.3 support at https://github.com/tyler-smith/zeus-rails23.

