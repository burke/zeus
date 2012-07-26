require 'socket'

forkpoint boot: ->{
  ENV_PATH  = File.expand_path('../config/environment',  __FILE__)
  BOOT_PATH = File.expand_path('../config/boot',  __FILE__)
  APP_PATH  = File.expand_path('../config/application',  __FILE__)
  ROOT_PATH = File.expand_path('..',  __FILE__)

  require BOOT_PATH

  forkpoint rails: ->{
    require 'rails/all'

    forkpoint default_bundle: -> {
      Bundler.require(:default)

      forkpoint(
        test: ->{
          ENV['RAILS_ENV'] = "test"
          Bundler.require(:test)
          require APP_PATH
          Rails.application.require_environment!

          forkpoint testrb: acceptor(".zeus.test_testrb.sock") {
            (r = Test::Unit::AutoRunner.new(true)).process_args(ARGV) or
              abort r.options.banner + " tests..."
            exit r.run
          }
        },

        dev: ->{
          Bundler.require(:development)
          ENV['RAILS_ENV'] = "development"
          require APP_PATH
          Rails.application.require_environment!

          forkpoint(
            generate: acceptor(".zeus.dev_generate.sock") {
              require 'rails/commands/generate'
            },
            console: acceptor(".zeus.dev_console.sock") {
              require 'rails/commands/console'
              Rails::Console.start(Rails.application)
            },
            runnner: acceptor(".zeus.dev_runner.sock") {
              require 'rails/commands/runner'
            },
            server: acceptor(".zeus.dev_server.sock") {
              require 'rails/commands/server'
              server = Rails::Server.new
              Dir.chdir(Rails.application.root)
              server.start
            },
            rake: ->{
              require 'rake'
              load 'Rakefile'

              forkpoint rake2: acceptor(".zeus.dev_rake.sock") {
                Rake.application.run
              }
            }
          )
        }
      )
    }
  }
}
