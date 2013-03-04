ROOT_PATH = File.expand_path(Dir.pwd)
ENV_PATH  = File.expand_path('config/environment',  ROOT_PATH)
BOOT_PATH = File.expand_path('config/boot',  ROOT_PATH)
APP_PATH  = File.expand_path('config/application',  ROOT_PATH)

require 'zeus'

def gem_is_bundled?(gem)
  gemfile_lock_contents = File.read(ROOT_PATH + "/Gemfile.lock")
  gemfile_lock_contents.scan(/\b#{gem} \(([^=~><]+?)\)/).flatten.first
end

if version = gem_is_bundled?('method_source')
  gem 'method_source', version
end

require 'zeus/m'

module Zeus
  class Rails < Plan
    def deprecated
      puts "Zeus 0.11.0 changed zeus.json. You'll have to rm zeus.json && zeus init."
    end
    alias_method :spec_helper, :deprecated
    alias_method :testrb,      :deprecated
    alias_method :rspec,       :deprecated


    def after_fork
      reconnect_activerecord
      restart_girl_friday
      reconnect_redis
    end

    def _monkeypatch_rake
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
      if defined?(Pry) && IRB == Pry
        require "pry"
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
      if File.exists?(ROOT_PATH + "/spec/spec_helper.rb")
        require 'spec_helper'
      elsif File.exist?(ROOT_PATH + "/test/minitest_helper.rb")
        require 'minitest_helper'
      else
        require 'test_helper'
      end
    end

    def test
      if spec_file?(ARGV) && defined?(RSpec)
        # disable autorun in case the user left it in spec_helper.rb
        RSpec::Core::Runner.disable_autorun!
        exit RSpec::Core::Runner.run(ARGV)
      else
        Zeus::M.run(ARGV)
      end
    end

    def cucumber_environment
      require 'cucumber/rspec/disable_option_parser'
      require 'cucumber/cli/main'
      @cucumber_runtime = Cucumber::Runtime.new
    end

    def cucumber
      cucumber_main = Cucumber::Cli::Main.new(ARGV.dup)
      exit_code = cucumber_main.execute!(@cucumber_runtime)
      if exit_code.nil?
        exit 0
      else
        exit exit_code
      end
    end


    private

    SPEC_DIR_REGEXP = %r"(^|/)spec"
    SPEC_FILE_REGEXP = /.+_spec\.rb$/
    def spec_file? argv
      last_arg = argv[-1]
      last_arg.match (Regexp.union(SPEC_DIR_REGEXP, SPEC_FILE_REGEXP))
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
