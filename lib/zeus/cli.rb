require 'zeus'
require 'zeus/vendored_thor'

module Zeus
  class CLI < Thor
    include Thor::Actions

    def initialize(*)
      super
      the_shell = (options['no-color'] ? Thor::Shell::Basic.new : shell)
      Zeus.ui = UI::Shell.new(the_shell)
      Zeus.ui.debug! if options['verbose']
    end

    check_unknown_options!

    default_task :help
    class_option "no-color", type: :boolean, banner: "Disable colorization in output"
    class_option "verbose",  type: :boolean, banner: "Enable verbose output mode", aliases: "-V"

    def help(cli = nil)
      case cli
      when nil then command = "zeus"
      else command = "zeus-#{cli}"
      end

      # manpages = %w(
      #   zeus-init
      #   zeus-start
      # )
      manpages = []

      if manpages.include?(command)
        root = File.expand_path("../man", __FILE__)

        if have_groff? && root !~ %r{^file:/.+!/META-INF/jruby.home/.+}
          groff   = "groff -Wall -mtty-char -mandoc -Tascii"
          pager   = ENV['MANPAGER'] || ENV['PAGER'] || 'less -R'

          Kernel.exec "#{groff} #{root}/#{command} | #{pager}"
        else
          puts File.read("#{root}/#{command}.txt")
        end
      else
        super
      end
    end

    desc "init", "Generates a zeus config file in the current working directory"
    long_desc <<-D
      Init tries to determine what kind of project is in the current working directory,
      and generates a relevant config file. Currently the only supported template is
      rails. You can force zeus to generate a rails config file with the --rails option.
    D
    method_option "rails", type: :string, banner: "Use the rails template instead of auto-detecting based on project contents"
    def init
      opts = options.dup
      if File.exist?(".zeus.rb")
        Zeus.ui.error ".zeus.rb already exists at #{Dir.pwd}/.zeus.rb"
        exit 1
      end

      if opts[:rails]
        template = :rails
      else
        # TODO: attempt to detect
        template = :rails
      end
      puts "Writing new .zeus.rb to #{Dir.pwd}/.zeus.rb"
      FileUtils.cp(File.expand_path("../templates/rails.rb", __FILE__), '.zeus.rb')
    end

    desc "start", "Start a zeus server for the project in the current working directory"
    long_desc <<-D
      starts a server that boots your application using the config file in
      .zeus.rb. The server will take several seconds to start, after which you may
      use the zeus runner commands (see `zeus help` for a list of available commands).
    D
    def start
      require 'zeus/server'
      begin
        require './.zeus.rb'
      rescue LoadError
        Zeus.ui.error("Your project is missing a config file (.zeus.rb), or you are not\n"\
          "in the project root. You can run `zeus init` to generate a config file.")
        exit 1
      end
      Zeus::Server.run
    end

    desc "version", "Print zeus's version information"
    def version
      Zeus.ui.info "Zeus version #{Zeus::VERSION}"
    end
    map %w(-v --version) => :version

    begin
      require './.zeus.rb'
      Zeus::Server.acceptor_names.each do |name|
        desc name, "#{name} task defined in .zeus.rb"
        define_method(name) {
          require 'zeus/client'
          Zeus::Client.run
        }
      end
    rescue LoadError
    end

    private

    def have_groff?
      !(`which groff` rescue '').empty?
    end


  end
end
