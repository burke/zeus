def find_rails_path(root_path)
  paths = %w(spec/dummy test/dummy .)
  paths.find { |path| File.exists?(File.expand_path(path, root_path)) }
end

ROOT_PATH = File.expand_path(Dir.pwd)
RAILS_PATH = find_rails_path(ROOT_PATH)
ENV_PATH  = File.expand_path('config/environment',  RAILS_PATH)
BOOT_PATH = File.expand_path('config/boot',  RAILS_PATH)
APP_PATH  = File.expand_path('config/application',  RAILS_PATH) unless defined? APP_PATH

require 'zeus'

def gem_is_bundled?(gem)
  gemfile_lock_contents = File.read(ROOT_PATH + "/Gemfile.lock")
  gemfile_lock_contents.scan(/^\s*#{gem} \(([^=~><]+?)\)/).flatten.first
end

if version = gem_is_bundled?('method_source')
  gem 'method_source', version
end

require 'zeus/m'

module Zeus
  class Rails < Plan
    def after_fork
      reconnect_activerecord
      restart_girl_friday
      reconnect_redis
    end

    def _monkeypatch_rake
      if version = gem_is_bundled?('rake')
        gem 'rake', version
      end
      require 'rake/testtask'
      Rake::TestTask.class_eval {

        # Create the tasks defined by this task lib.
        def define
          desc "Run tests" + (@name==:test ? "" : " for #{@name}")
          task @name do
            # ruby "#{ruby_opts_string} #{run_code} #{file_list_string} #{option_list}"
            rails_env = ENV['RAILS_ENV']
            rubyopt = ENV['RUBYOPT']
            ENV['RAILS_ENV'] = nil
            ENV['RUBYOPT'] = nil # bundler sets this to require bundler :|
            puts "zeus test #{file_list_string}"
            ret = system "zeus test #{file_list_string}"
            ENV['RAILS_ENV'] = rails_env
            ENV['RUBYOPT'] = rubyopt
            ret
          end
          self
        end

        alias_method :_original_define, :define

        def self.inherited(klass)
          return unless klass.name == "TestTaskWithoutDescription"
          klass.class_eval {
            def self.method_added(sym)
              class_eval do
                if !@rails_hack_reversed
                  @rails_hack_reversed = true
                  alias_method :define, :_original_define
                  def desc(*)
                  end
                end
              end
            end
          }
        end
      }
    end

    def boot
      _monkeypatch_rake
      $LOAD_PATH.unshift "./lib"

      require BOOT_PATH
      # config/application.rb normally requires 'rails/all'.
      # Some 'alternative' ORMs such as Mongoid give instructions to switch this require
      # out for a list of railties, not including ActiveRecord.
      # We grep config/application.rb for all requires of rails/all or railties, and require them.
      rails_components = File.read(APP_PATH + ".rb").
        scan(/^\s*require\s*['"](.*railtie.*|rails\/all)['"]/).flatten

      rails_components = ["rails/all"] if rails_components == []
      rails_components.each do |component|
        require component
      end
    end

    def default_bundle
      Bundler.require(:default)
      Zeus::LoadTracking.add_feature('./Gemfile.lock')
    end

    def development_environment
      Bundler.require(:development)
      ::Rails.env = ENV['RAILS_ENV'] = "development"
      require APP_PATH
      ::Rails.application.require_environment!
    end

    def prerake
      require 'rake'
    end

    def rake
      Rake.application.run
    end

    def generate
      load_rails_generators
      require 'rails/commands/generate'
    end

    def destroy
      load_rails_generators
      require 'rails/commands/destroy'
    end

    def runner
      require 'rails/commands/runner'
    end

    def console
      require 'rails/commands/console'

      if defined?(Pry)
        # Adding Rails Console helpers to Pry.
        if (3..4).include?(::Rails::VERSION::MAJOR)
          require 'rails/console/app'
          require 'rails/console/helpers'
          TOPLEVEL_BINDING.eval('self').extend ::Rails::ConsoleMethods
        end

        Pry.start
      else
        ::Rails::Console.start(::Rails.application)
      end
    end

    def dbconsole
      require 'rails/commands/dbconsole'

      meth = ::Rails::DBConsole.method(:start)

      # `Rails::DBConsole.start` has been changed to load faster in
      # https://github.com/rails/rails/commit/346bb018499cde6699fcce6c68dd7e9be45c75e1
      #
      # This will work with both versions.
      if meth.arity.zero?
        ::Rails::DBConsole.start
      else
        ::Rails::DBConsole.start(::Rails.application)
      end
    end

    def server
      require 'rails/commands/server'
      server = ::Rails::Server.new
      Dir.chdir(::Rails.application.root)
      server.start
    end

    def test_environment
      Bundler.require(:test)

      ::Rails.env = ENV['RAILS_ENV'] = 'test'
      require APP_PATH

      $rails_rake_task = 'yup' # lie to skip eager loading
      ::Rails.application.require_environment!
      $rails_rake_task = nil

      $LOAD_PATH.unshift ".", "./lib", "./test", "./spec"
    end

    def test_helper
      if ENV['RAILS_TEST_HELPER']
        require ENV['RAILS_TEST_HELPER']
      else
        if File.exists?(ROOT_PATH + "/spec/rails_helper.rb")
          # RSpec >= 3.0+
          require 'rails_helper'
        elsif File.exists?(ROOT_PATH + "/spec/spec_helper.rb")
          # RSpec < 3.0
          require 'spec_helper'
        elsif File.exists?(ROOT_PATH + "/test/minitest_helper.rb")
          require 'minitest_helper'
        else
          require 'test_helper'
        end
      end
    end

    def test(argv=ARGV)
      # if there are two test frameworks and one of them is RSpec,
      # then "zeus test/rspec/testrb" without arguments runs the
      # RSpec suite by default.
      if using_rspec?(argv)
        ARGV.replace(argv)
        if RSpec::Core::Runner.respond_to?(:invoke)
          RSpec::Core::Runner.invoke
        else
          RSpec::Core::Runner.run(argv)
        end
      else
        require 'minitest/autorun' if using_minitest?
        # Minitest and old Test::Unit
        Zeus::M.run(argv)
      end
    end

    private

    def using_rspec?(argv)
      defined?(RSpec) && (argv.empty? || spec_file?(argv))
    end

    def using_minitest?
      defined?(:MiniTest) || defined?(:Minitest)
    end

    SPEC_DIR_REGEXP = %r"(^|/)spec"
    SPEC_FILE_REGEXP = /.+_spec\.rb$/

    def spec_file?(argv)
      argv.any? do |arg|
        arg.match(Regexp.union(SPEC_DIR_REGEXP, SPEC_FILE_REGEXP))
      end
    end

    def restart_girl_friday
      return unless defined?(GirlFriday::WorkQueue)
      # The Actor is run in a thread, and threads don't persist post-fork.
      # We just need to restart each one in the newly-forked process.
      ObjectSpace.each_object(GirlFriday::WorkQueue) do |obj|
        obj.send(:start)
      end
    end

    def reconnect_activerecord
      return unless defined?(ActiveRecord::Base)
      begin
        ActiveRecord::Base.clear_all_connections!
        ActiveRecord::Base.establish_connection
        if ActiveRecord::Base.respond_to?(:shared_connection)
          ActiveRecord::Base.shared_connection = ActiveRecord::Base.retrieve_connection
        end
      rescue ActiveRecord::AdapterNotSpecified
      end
    end

    def reconnect_redis
      return unless defined?(Redis::Client)
      ObjectSpace.each_object(Redis::Client) do |client|
        client.connect rescue nil
      end
    end

    def load_rails_generators
      require 'rails/generators'
      ::Rails.application.load_generators
    rescue LoadError # Rails 3.0 doesn't require this block to be run, but 3.2+ does
    end

  end
end

Zeus.plan ||= Zeus::Rails.new
