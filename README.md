# Zeus

## What?

Zeus preloads your app so that your normal development tasks such as `console`, `server`, `generate`, and tests are faster.

## Why?

Because waiting 25 seconds sucks, but waiting 0.4 seconds doesn't.

## When?

Not yet. Zeus is nowhere near production-ready yet. Use only if you really like broken things.

## Ugly bits

* Probably crashes a lot
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

* zeus init command
* Handle client/server without requiring a unix socket for each acceptor (1 shared socket)
* Make the code less terrible
* Figure out how to run full test suites without multiple env loads
* Support other frameworks?
* Use fsevents instead of kqueue to reduce the obscene number of file descriptors.
* Support epoll on linux

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Added some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
