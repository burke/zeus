# Zeus

WARNING: Nowhere near production-ready yet. Use only if you really like broken things.

TODO: Write a gem description

## Installation

Install the gem.

    gem install zeus

Copy `examples/rails.rb` to `{your app}/.zeus.rb`

## Usage

1. Start the server:

    zeus start

2. Run some commands:

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
