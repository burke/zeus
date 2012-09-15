ROOT_PATH = File.expand_path(Dir.pwd)
ENV_PATH  = File.expand_path('config/environment',  ROOT_PATH)
BOOT_PATH = File.expand_path('config/boot',  ROOT_PATH)
APP_PATH  = File.expand_path('config/application',  ROOT_PATH)

require 'zeus'
require 'zeus/m'

module Zeus
  class Rails < Plan
    def after_fork
      reconnect_activerecord
      restart_girl_friday
    end

    def boot
      require BOOT_PATH
      # config/application.rb normally requires 'rails/all'.
      # Some 'alternative' ORMs such as Mongoid give instructions to switch this require
      # out for a list of railties, not including ActiveRecord.
      # We grep config/application.rb for all requires of rails/all or railties, and require them.
      rails_components = File.read(APP_PATH + ".rb").
        scan(/^\s*require\s*['"](.*railtie.*|rails\/all)['"]/).flatten

      rails_components == ["rails/all"] if rails_components == []
      rails_components.each do |component|
        require component
      end
    end

    def default_bundle
      Bundler.require(:default)
    end

    def development_environment
      Bundler.require(:development)
      ::Rails.env = ENV['RAILS_ENV'] = "development"
      require APP_PATH
      ::Rails.application.require_environment!
    end

    def prerake
      require 'rake'
      load 'Rakefile'
    end

    def rake
      Rake.application.run
    end

    def generate
      begin
        require 'rails/generators'
        ::Rails.application.load_generators
      rescue LoadError # Rails 3.0 doesn't require this block to be run, but 3.2+ does
      end
      require 'rails/commands/generate'
    end

    def runner
      require 'rails/commands/runner'
    end

    def console
      require 'rails/commands/console'
      ::Rails::Console.start(::Rails.application)
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
      $LOAD_PATH.unshift(ROOT_PATH) unless $LOAD_PATH.include?(ROOT_PATH)
      $LOAD_PATH.unshift(ROOT_PATH + "/lib") unless $LOAD_PATH.include?(ROOT_PATH + "/lib")

      if Dir.exist?(ROOT_PATH + "/test")
        test = File.join(ROOT_PATH, 'test')
        $LOAD_PATH.unshift(test) unless $LOAD_PATH.include?(test)
      end

      if Dir.exist?(ROOT_PATH + "/spec")
        spec = File.join(ROOT_PATH, 'spec')
        $LOAD_PATH.unshift(spec) unless $LOAD_PATH.include?(spec)
      end
    end

    def test_helper
      if File.exist?(ROOT_PATH + "/test/minitest_helper.rb")
        require 'minitest_helper'
      else
        require 'test_helper'
      end
    end

    def test
      Zeus::M.run(ARGV)
    end
    alias_method :testrb, :test # for compatibility with 0.10.x

    def spec_helper
      require 'spec_helper'
    end

    def rspec
      exit RSpec::Core::Runner.run(ARGV)
    end

    def cucumber_environment
      require 'cucumber/rspec/disable_option_parser'
      require 'cucumber/cli/main'
      cucumber_runtime = Cucumber::Runtime.new
    end

    def cucumber
      cucumber_main = Cucumber::Cli::Main.new(ARGV.dup)
      exit cucumber_main.execute!(cucumber_runtime)
    end

    private

    def restart_girl_friday
      return unless defined?(GirlFriday::WorkQueue)
      # The Actor is run in a thread, and threads don't persist post-fork.
      # We just need to restart each one in the newly-forked process.
      ObjectSpace.each_object(GirlFriday::WorkQueue) do |obj|
        obj.send(:start)
      end
    end

    def reconnect_activerecord
      ActiveRecord::Base.clear_all_connections! rescue nil
      ActiveRecord::Base.establish_connection   rescue nil
    end

  end
end

Zeus.plan ||= Zeus::Rails.new

