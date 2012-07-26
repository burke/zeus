# Zeus

WARNING: Nowhere near production-ready yet. Use only if you really like broken things.

TODO: Write a gem description

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

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Added some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
