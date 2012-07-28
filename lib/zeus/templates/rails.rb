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
          Rails.env = ENV['RAILS_ENV'] = "development"
          require APP_PATH
          Rails.application.require_environment!
        end

        command :generate, :g do
          require 'rails/commands/generate'
        end

        command :runner, :r do
          require 'rails/commands/runner'
        end

        command :console, :c do
          require 'rails/commands/console'
          Rails::Console.start(Rails.application)
        end

        command :server, :s do
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

          command :rake do
            Rake.application.run
          end

        end
      end

      # stage :test do
      #   action do
      #     Rails.env = ENV['RAILS_ENV'] = "test"
      #     Bundler.require(:test)
      #     require APP_PATH
      #     Rails.application.require_environment!
      #   end

      #   command :testrb do
      #     (r = Test::Unit::AutoRunner.new(true)).process_args(ARGV) or
      #       abort r.options.banner + " tests..."
      #     exit r.run
      #   end

      # end

    end
  end
end

