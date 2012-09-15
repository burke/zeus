# This is basically just a vendorization of @qrush's `m`.
module Zeus
  #`m`  stands for metal, which is a better test/unit test runner that can run
  #tests by line number.
  #
  #[![m ci](https://secure.travis-ci.org/qrush/m.png)](http://travis-ci.org/qrush/m)
  #
  #![Rush is a heavy metal band. Look it up on Wikipedia.](https://raw.github.com/qrush/m/master/rush.jpg)
  #
  #<sub>[Rush at the Bristol Colston Hall May 1979](http://www.flickr.com/photos/8507625@N02/3468299995/)</sub>
  ### Install
  #
  ### Usage
  #
  #Basically, I was sick of using the `-n` flag to grab one test to run. Instead, I
  #prefer how RSpec's test runner allows tests to be run by line number.
  #
  #Given this file:
  #
  #     $ cat -n test/example_test.rb
  #      1	require 'test/unit'
  #      2
  #      3	class ExampleTest < Test::Unit::TestCase
  #      4	  def test_apple
  #      5	    assert_equal 1, 1
  #      6	  end
  #      7
  #      8	  def test_banana
  #      9	    assert_equal 1, 1
  #     10	  end
  #     11	end
  #
  #You can run a test by line number, using format `m TEST_FILE:LINE_NUMBER_OF_TEST`:
  #
  #     $ m test/example_test.rb:4
  #     Run options: -n /test_apple/
  #
  #     # Running tests:
  #
  #     .
  #
  #     Finished tests in 0.000525s, 1904.7619 tests/s, 1904.7619 assertions/s.
  #
  #     1 tests, 1 assertions, 0 failures, 0 errors, 0 skips
  #
  #Hit the wrong line number? No problem, `m` helps you out:
  #
  #     $ m test/example_test.rb:2
  #     No tests found on line 2. Valid tests to run:
  #
  #      test_apple: m test/examples/test_unit_example_test.rb:4
  #     test_banana: m test/examples/test_unit_example_test.rb:8
  #
  #Want to run the whole test? Just leave off the line number.
  #
  #     $ m test/example_test.rb
  #     Run options:
  #
  #     # Running tests:
  #
  #     ..
  #
  #     Finished tests in 0.001293s, 1546.7904 tests/s, 3093.5808 assertions/s.
  #
  #     1 tests, 2 assertions, 0 failures, 0 errors, 0 skips
  #
  #### Supports
  #
  #`m` works with a few Ruby test frameworks:
  #
  #* `Test::Unit`
  #* `ActiveSupport::TestCase`
  #* `MiniTest::Unit::TestCase`
  #
  ### License
  #
  #This gem is MIT licensed, please see `LICENSE` for more information.

  ### M, your metal test runner
  # Maybe this gem should have a longer name? Metal?
  module M
    VERSION = "1.2.1" unless defined?(VERSION)

    # Accept arguments coming from bin/m and run tests.
    def self.run(argv)
      Runner.new(argv).run
    end

    ### Runner is in charge of running your tests.
    # Instead of slamming all of this junk in an `M` class, it's here instead.
    class Runner
      def initialize(argv)
        @argv = argv
      end

      # There's two steps to running our tests:
      # 1. Parsing the given input for the tests we need to find (or groups of tests)
      # 2. Run those tests we found that match what you wanted
      def run
        parse
        execute
      end

      private

      def parse
        # With no arguments,
        if @argv.empty?
          # Just shell out to `rake test`.
          exec "rake test"
        else
          parse_options! @argv

          # Parse out ARGV, it should be coming in in a format like `test/test_file.rb:9`
          @file, line = @argv.first.split(':')
          @line ||= line.to_i

          # If this file is a directory, not a file, run the tests inside of this directory
          if Dir.exist?(@file)
            # Make a new rake test task with a hopefully unique name, and run every test looking file in it
            require "rake/testtask"
            Rake::TestTask.new(:m_custom) do |t|
              t.libs << 'test'
              t.pattern = "#{@file}/*test*.rb"
            end
            # Invoke the rake task and exit, hopefully it'll work!
            Rake::Task['m_custom'].invoke
            exit
          end
        end
      end

      def parse_options!(argv)
        require 'optparse'

        OptionParser.new do |opts|
          opts.banner  = 'Options:'
          opts.version = M::VERSION

          opts.on '-h', '--help', 'Display this help.' do
            puts "Usage: m [OPTIONS] [FILES]\n\n", opts
            exit
          end

          opts.on '--version', 'Display the version.' do
            puts "m #{M::VERSION}"
            exit
          end

          opts.on '-l', '--line LINE', Integer, 'Line number for file.' do |line|
            @line = line
          end

          opts.parse! argv
        end
      end

      def execute
        # Locate tests to run that may be inside of this line. There could be more than one!
        tests_to_run = tests.within(@line)

        # If we found any tests,
        if tests_to_run.size > 0
          # assemble the regexp to run these tests,
          test_names = tests_to_run.map(&:name).join('|')

          # set up the args needed for the runner
          test_arguments = ["-n", "/(#{test_names})/"]

          # directly run the tests from here and exit with the status of the tests passing or failing
          if defined?(MiniTest)
            exit MiniTest::Unit.runner.run test_arguments
          elsif defined?(Test)
            exit Test::Unit::AutoRunner.run(false, nil, test_arguments)
          else
            not_supported
          end
        else
          # Otherwise we found no tests on this line, so you need to pick one.
          message = "No tests found on line #{@line}. Valid tests to run:\n\n"

          # For every test ordered by line number,
          # spit out the test name and line number where it starts,
          tests.by_line_number do |test|
            message << "#{sprintf("%0#{tests.column_size}s", test.name)}: m #{@file}:#{test.start_line}\n"
          end

          # fail like a good unix process should.
          abort message
        end
      end

      # Finds all test suites in this test file, with test methods included.
      def suites
        # Since we're not using `ruby -Itest -Ilib` to run the tests, we need to add this directory to the `LOAD_PATH`
        $:.unshift "./test", "./lib"

        begin
          # Fire up this Ruby file. Let's hope it actually has tests.
          load @file
        rescue LoadError => e
          # Fail with a happier error message instead of spitting out a backtrace from this gem
          abort "Failed loading test file:\n#{e.message}"
        end

        # Figure out what test framework we're using
        if defined?(MiniTest)
          suites = MiniTest::Unit::TestCase.test_suites
        elsif defined?(Test)
          suites = Test::Unit::TestCase.test_suites
        else
          not_supported
        end

        # Use some janky internal APIs to group test methods by test suite.
        suites.inject({}) do |suites, suite_class|
          # End up with a hash of suite class name to an array of test methods, so we can later find them and ignore empty test suites
          suites[suite_class] = suite_class.test_methods if suite_class.test_methods.size > 0
          suites
        end
      end

      # Shoves tests together in our custom container and collection classes.
      # Memoize it since it's unnecessary to do this more than one for a given file.
      def tests
        @tests ||= begin
          require "zeus/m/test_collection"
          require "zeus/m/test_method"
          # With each suite and array of tests,
          # and with each test method present in this test file,
          # shove a new test method into this collection.
          suites.inject(TestCollection.new) do |collection, (suite_class, test_methods)|
            test_methods.each do |test_method|
              collection << TestMethod.create(suite_class, test_method)
            end
            collection
          end
        end
      end

      # Fail loudly if this isn't supported
      def not_supported
        abort "This test framework is not supported! Please open up an issue at https://github.com/qrush/m !"
      end
    end
  end
end
