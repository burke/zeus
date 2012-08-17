require 'thrud'

module Zeus
  class CLI < Thrud

    def initialize(*)
      super
      Zeus.ui = Zeus::UI.new
      Zeus.ui.debug! #if options['verbose']
      if defined?(Bundler)
        Zeus.ui.warn "Don't run Zeus with `bundle exec`. It's unnecessarily slow."
      end
    end

    desc "init", "Generates a zeus config file in the current working directory"
    long_desc <<-D
      Init tries to determine what kind of project is in the current working directory,
      and generates a relevant config file. Currently the only supported template is
      rails.
    D
    # method_option "rails", type: :string, banner: "Use the rails template instead of auto-detecting based on project contents"
    def init
      require 'fileutils'

      if File.exist?(".zeus.rb")
        Zeus.ui.error ".zeus.rb already exists at #{Dir.pwd}/.zeus.rb"
        exit 1
      end

      Zeus.ui.info "Writing new .zeus.rb to #{Dir.pwd}/.zeus.rb"
      FileUtils.cp(File.expand_path("../templates/rails.rb", __FILE__), '.zeus.rb')
    end

    desc "start", "Start a zeus server for the project in the current working directory"
    long_desc <<-D
      starts a server that boots your application using the config file in
      .zeus.rb. The server will take several seconds to start, after which you may
      use the zeus runner commands (see `zeus help` for a list of available commands).
    D
    def start
      begin
        require self.class.definition_file
      rescue LoadError
        Zeus.ui.error("Your project is missing a config file (.zeus.rb), and it doesn't appear\n"\
          "to be a rails project. You can run `zeus init` to generate a config file.")
        exit 1
      end
      Zeus::Server.new.run
    end

    def help(*)
      super
    end

    desc "version", "Print zeus's version information and exit"
    def version
      Zeus.ui.info "Zeus version #{Zeus::VERSION}"
    end
    map %w(-v --version) => :version

    def self.definition_file
      if !File.exists?('.zeus.rb') && File.exists?('script/rails')
        File.expand_path("../templates/rails.rb", __FILE__)
      else
        './.zeus.rb'
      end
    end

    begin
      require definition_file
    rescue LoadError
    end

    Zeus::Server.acceptors.each do |acc|
      desc acc.name, (acc.description || "#{acc.name} task defined in zeus definition file")
      define_method(acc.name) { |*args|
        Zeus::Client.run(acc.name, args)
      }
      map acc.aliases => acc.name
    end

  end
end
