require 'socket'

Zeus::Server.define! do
  stage :boot do

    action do
      ENV_PATH  = File.expand_path('../config/environment',  __FILE__)
      BOOT_PATH = File.expand_path('../config/boot',  __FILE__)
      APP_PATH  = File.expand_path('../config/application',  __FILE__)
      ROOT_PATH = File.expand_path('..',  __FILE__)

      require BOOT_PATH
      require 'rails/all'
    end

    stage :default_bundle do
      action { Bundler.require(:default) }

      stage :dev do
        action do
          Bundler.require(:development)
          ENV['RAILS_ENV'] = "development"
          require APP_PATH
          Rails.application.require_environment!
        end

        acceptor :generate, ".zeus.dev_generate.sock" do
          require 'rails/commands/generate'
        end
        acceptor :runner, ".zeus.dev_runner.sock" do
          require 'rails/commands/runner'
        end
        acceptor :console, ".zeus.dev_console.sock" do
          require 'rails/commands/console'
          Rails::Console.start(Rails.application)
        end

        acceptor :server, ".zeus.dev_server.sock" do
          require 'rails/commands/server'
          server = Rails::Server.new
          Dir.chdir(Rails.application.root)
          server.start
        end

        stage :prerake do
          action do
            require 'rake'
            load 'Rakefile'
          end

          acceptor :rake, ".zeus.dev_rake.sock" do
            Rake.application.run
          end

        end
      end

      stage :test do
        action do
          ENV['RAILS_ENV'] = "test"
          Bundler.require(:test)
          require APP_PATH
          Rails.application.require_environment!
        end

        acceptor :testrb, ".zeus.test_testrb.sock" do
          (r = Test::Unit::AutoRunner.new(true)).process_args(ARGV) or
            abort r.options.banner + " tests..."
          exit r.run
        end

      end

    end
  end
end
