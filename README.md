# Zeus

## What?

Zeus preloads your app so that your normal development tasks such as `console`, `server`, `generate`, and tests are faster.

## Why?

Because waiting 25 seconds sucks, but waiting 0.4 seconds doesn't.

## When?

Not yet. Zeus is nowhere near production-ready yet. Use only if you really like broken things.

## Installation

Install the gem.

    git clone git://github.com/burke/zeus.git
    cd zeus
    gem build zeus.gemspec
    gem install zeus-*.gem

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

* Kill process when files are detected to have changed
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
