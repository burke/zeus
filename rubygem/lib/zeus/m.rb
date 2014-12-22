# This is very largely based on @qrush's M, but there are many modifications.

# we need to load all dependencies up front, because bundler will
# remove us from the load path soon.
require "rubygems"
require "zeus/m/test_collection"
require "zeus/m/test_method"

# the Gemfile may specify a version of method_source, but we also want to require it here.
# To avoid possible "you've activated X; gemfile specifies Y" errors, we actually scan
# Gemfile.lock for a specific version, and require exactly that version if present.
gemfile_lock = ROOT_PATH + "/Gemfile.lock"
if File.exists?(gemfile_lock)
  version = File.read(ROOT_PATH + "/Gemfile.lock").
    scan(/\bmethod_source\s*\(([\d\.]+)\)/).flatten[0]

  gem "method_source", version if version
end

require 'method_source'

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
    M::VERSION = "1.2.1" unless defined?(M::VERSION)

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
          require 'rake'
          Rake.application.init
          Rake.application.load_rakefile
          Rake.application.invoke_task("test")
          exit
        else
          parse_options! @argv

          # Parse out ARGV, it should be coming in in a format like `test/test_file.rb:9`
          _, line = @argv.first.split(':')
          @line ||= line.nil? ? nil : line.to_i

          @files = []
          @argv.each do |arg|
            add_file(arg)
          end
        end
      end

      def add_file(arg)
        file = arg.split(':').first
        if Dir.exist?(file)
          files = Dir.glob("#{file}/**/*test*.rb")
          @files.concat(files)
        else
          files = Dir.glob(file)
          files == [] and abort "Couldn't find test file '#{file}'!"
          @files.concat(files)
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

          opts.on '-n', '--name NAME', String, 'Name or pattern for test methods to run.' do |name|
            if name[0] == "/" && name[-1] == "/"
              @test_name = Regexp.new(name[1..-2])
            else
              @test_name = name
            end
          end

          opts.parse! argv
        end
      end

      def execute
        generate_tests_to_run

        test_arguments = build_test_arguments

        # directly run the tests from here and exit with the status of the tests passing or failing
        case framework
        when :minitest5
          nerf_test_unit_autorunner
          exit(Minitest.run(test_arguments) ? 0 : 1)
        when :minitest_old
          nerf_test_unit_autorunner
          exit(MiniTest::Unit.runner.run(test_arguments).to_i)
        when :testunit1, :testunit2
          exit Test::Unit::AutoRunner.run(false, nil, test_arguments)
        else
          not_supported
        end
      end

      def generate_tests_to_run
        # Locate tests to run that may be inside of this line. There could be more than one!
        all_tests = tests
        if @line
          @tests_to_run = all_tests.within(@line)
        end
      end

      def build_test_arguments
        if @line
          abort_with_no_test_found_by_line_number if @tests_to_run.empty?

          # assemble the regexp to run these tests,
          test_names = @tests_to_run.map(&:escaped_name).join('|')

          # set up the args needed for the runner
          ["-n", "/(#{test_names})/"]
        elsif user_specified_name?
          abort_with_no_test_found_by_name unless tests.contains?(@test_name)

          ["-n", test_name_to_s]
        else
          []
        end
      end

      def abort_with_no_test_found_by_line_number
        abort_with_valid_tests_msg "No tests found on line #{@line}. "
      end

      def abort_with_no_test_found_by_name
        abort_with_valid_tests_msg "No test name matches '#{test_name_to_s}'. "
      end

      def abort_with_valid_tests_msg message=""
        message << "Valid tests to run:\n\n"
        # For every test ordered by line number,
        # spit out the test name and line number where it starts,
        tests.by_line_number do |test|
          message << "#{sprintf("%0#{tests.column_size}s", test.escaped_name)}: zeus test #{@files[0]}:#{test.start_line}\n"
        end

        # fail like a good unix process should.
        abort message
      end

      def test_name_to_s
        @test_name.is_a?(Regexp)? "/#{@test_name.source}/" : @test_name
      end

      def user_specified_name?
        !@test_name.nil?
      end

      def framework
        @framework ||= begin
          if defined?(Minitest::Runnable)
            :minitest5
          elsif defined?(MiniTest)
            :minitest_old
          elsif defined?(Test)
            if Test::Unit::TestCase.respond_to?(:test_suites)
              :testunit2
            else
              :testunit1
            end
          end
        end
      end

      # Finds all test suites in this test file, with test methods included.
      def suites
        # Since we're not using `ruby -Itest -Ilib` to run the tests, we need to add this directory to the `LOAD_PATH`
        $:.unshift "./test", "./lib"

        if framework == :testunit1
          Test::Unit::TestCase.class_eval {
            @@test_suites = {}
            def self.inherited(klass)
              @@test_suites[klass] = true
            end
            def self.test_suites
              @@test_suites.keys
            end
            def self.test_methods
              public_instance_methods(true).grep(/^test/).map(&:to_s)
            end
          }
        end

        begin
          # Fire up the Ruby files. Let's hope they actually have tests.
          @files.each { |f| load f }
        rescue LoadError => e
          # Fail with a happier error message instead of spitting out a backtrace from this gem
          abort "Failed loading test file:\n#{e.message}"
        end

        # Figure out what test framework we're using
        case framework
        when :minitest5
          suites = Minitest::Runnable.runnables
        when :minitest_old
          suites = MiniTest::Unit::TestCase.test_suites
        when :testunit1, :testunit2
          suites = Test::Unit::TestCase.test_suites
        else
          not_supported
        end

        # Use some janky internal APIs to group test methods by test suite.
        suites.inject({}) do |suites, suite_class|
          # End up with a hash of suite class name to an array of test methods, so we can later find them and ignore empty test suites
          test_methods = case framework
          when :minitest5
            suite_class.runnable_methods
          else
            suite_class.test_methods
          end
          suites[suite_class] = test_methods if test_methods.size > 0
          suites
        end
      end

      # Shoves tests together in our custom container and collection classes.
      # Memoize it since it's unnecessary to do this more than one for a given file.
      def tests
        @tests ||= begin
          # With each suite and array of tests,
          # and with each test method present in this test file,
          # shove a new test method into this collection.
          suites.inject(TestCollection.new) do |collection, (suite_class, test_methods)|
            test_methods.each do |test_method|
              find_locations = (@files.size == 1 && @line)
              collection << TestMethod.create(suite_class, test_method, find_locations)
            end
            collection
          end
        end
      end

      def nerf_test_unit_autorunner
        return unless defined?(Test::Unit::Runner)
        if Test::Unit::Runner.class_variable_get("@@installed_at_exit")
          Test::Unit::Runner.class_variable_set("@@stop_auto_run", true)
        end
      end

      # Fail loudly if this isn't supported
      def not_supported
        abort "This test framework is not supported! Please open up an issue at https://github.com/qrush/m !"
      end
    end
  end
end
