# Zeus

[![Join the chat at https://gitter.im/zeus-application-preloader/Lobby](https://badges.gitter.im/zeus-application-preloader/Lobby.svg)](https://gitter.im/zeus-application-preloader/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Build Status](https://travis-ci.org/burke/zeus.svg?branch=master)](https://travis-ci.org/burke/zeus)

Zeus preloads your Rails app so that your normal development tasks such as `console`, `server`, `generate`, and specs/tests take **less than one second**.

This screencast gives a quick overview of how to use zeus with Rails.

[![Watch the screencast!](http://s3.amazonaws.com/burkelibbey/vimeo-zeus.png)](http://vimeo.com/burkelibbey/zeus)

Zeus is also covered in [RailsCasts episode 412](http://railscasts.com/episodes/412-fast-rails-commands).

More generally, Zeus is a language-agnostic application checkpointer for non-multithreaded applications. Currently only ruby is targeted, but explicit support for other languages is possible.


## Requirements (for use with Rails)

* OS X 10.7+ *OR* Linux 2.6.13+
* Rails 3.x or 4.x
* Compatible Ruby installation
  * Ruby 2.0+
  * Rubinius

**Please note**: Zeus requires your project to be running on a file system that supports FSEvents or inotify. This means no NFS, CIFS, Samba, or VBox/VMWare shared folders.


## Installation

Install the gem.

    gem install zeus

Q: "I should put it in my `Gemfile`, right?"

A: No. You can, but running `bundle exec zeus` instead of `zeus` adds precious seconds to commands that otherwise would be quite a bit faster. Zeus was built to be run from outside of bundler.

#### IMPORTANT

It is common to see tests running twice when starting out with Zeus. If you see your tests/specs running twice, you should try disabling `require 'rspec/autotest'` and `require 'rspec/autorun'` (for RSpec), or `require 'minitest/autorun'` (for Minitest). (see [#134](https://github.com/burke/zeus/issues/134) for more information).


## Rails Set up 

In your app's directory initialize zeus:

    zeus init

which will create two files

    custom_plan.rb
    zeus.json

Force the test environment: In `test/test_helper.rb' (or `spec/spec_helper.rb` if using RSpec), change:


    ENV['RAILS_ENV'] ||= 'test'

To:

    ENV['RAILS_ENV'] = 'test'

_More helpful set up hints at [this very nice guide put together by Thoughtbot](https://robots.thoughtbot.com/improving-rails-boot-time-with-zeus)_

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

## Limitations

You need to restart zeus if you make changes to various initialization files. Examples of these files include:

 * FactoryGirl factories
 * RSpec support files

## Related gems

* [Spork](https://github.com/sporkrb/spork) - a [DRb server](http://www.ruby-doc.org/stdlib-1.9.3/libdoc/drb/rdoc/DRb.html) that forks before each run to ensure a clean testing state
* [Commands](https://github.com/rails/commands) - a persistent console that runs Rails commands without reloading the env
* [Spring](https://github.com/rails/spring) - like Zeus but in pure Ruby, totally automatic, and included in Rails 4.1+.

If you're switching from Spork, be sure to [read the wiki page on Spork](https://github.com/burke/zeus/wiki/Spork).


## Customizing Zeus Commands

To add/modify commands, see [`docs/ruby/modifying.md`](docs/ruby/modifying.md).


## Contributing

To get started hacking on Zeus itself, see [`docs/overview.md`](docs/overview.md).

See also the handy contribution guide at [`contributing.md`](contributing.md).


## Rails 2.3 Support

The default plan bundled with zeus only supports Rails 3.x and 4.x. There is a project to provide Rails 2.3 support [here](https://github.com/tyler-smith/zeus-rails23), however it has not been updated in some time.

## Legacy Ruby Versions

Due to official Ruby support lapsing in Feb 2015, Ruby 1.9.3 and below are no longer officially supported. Zeus may work with them but we are not actively supporting them at this time. The last release tested with these versions is [0.15.4](https://github.com/burke/zeus/tree/v0.15.4).

