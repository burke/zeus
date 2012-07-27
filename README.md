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
* Creates a bunch of sockets
* Uses an obscene number of file descriptors

## Installation

Install the gem.

    gem install zeus

Copy `examples/rails.rb` to `{your app}/.zeus.rb`

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

* Better message when no config file is found
* Try to detect what app we're in and use a stock config file
* zeus init command
* Handle client/server without requiring a unix socket for each acceptor (1 shared socket)
* Make the code less terrible
* Figure out how to run full test suites without multiple env loads
* Support other frameworks?
* Use fsevent instead of kqueue to reduce the obscene number of file descriptors.
* Support epoll on linux
* Once it's really stable, transparently start the process in the background rather than requiring zeus start.
* Handle connections for not-yet-started sockets
* Don't replace a socket with changed deps until the new one is ready

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Added some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request

## Thanks...

* To [Jesse Storimer](http://github.com/jstorimer) for spin, part of the inspiration for this project
* To [Samuel Kadolph](http://github.com/samuelkadolph) for doing most of the cross-process pseudoterminal legwork.
* To [Shopify](http://github.com/Shopify) for letting me spend (some of) my days working on this.

## Doesn't work for you?

Try these libraries instead:

* [spin](https://github.com/jstorimer/spin)
* [spork](https://github.com/sporkrb/spork)
* [guard](https://github.com/guard/guard)
