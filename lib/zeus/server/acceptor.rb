require 'json'
require 'socket'

# See Zeus::Server::ClientHandler for relevant documentation
module Zeus
  class Server
    class Acceptor < ForkedProcess
      attr_accessor :aliases, :description, :action

      def descendent_acceptors
        self
      end

      def before_setup
        register_with_client_handler(Process.pid)
      end

      def runloop!
        loop do
          prefork_action!
          terminal, arguments = accept_connection # blocking
          child = fork { __RUNNER__run(terminal, arguments) }
          terminal.close

          Process.detach(child)
        end
      end

      def __RUNNER__run(terminal, arguments)
        $0 = "zeus runner: #{@name}"
        Process.setsid
        postfork_action!
        @s_acceptor << $$ << "\n"
        $stdin.reopen(terminal)
        $stdout.reopen(terminal)
        $stderr.reopen(terminal)
        ARGV.replace(arguments)

        @action.call
      ensure
        # TODO this is a whole lot of voodoo that I don't really understand.
        # I need to figure out how best to make the process disconenct cleanly.
        dnw, dnr = File.open("/dev/null", "w+"), File.open("/dev/null", "r+")
        $stderr.reopen(dnr)
        $stdout.reopen(dnr)
        terminal.close
        $stdin.reopen(dnw)
        Process.kill(9, $$)
        exit 0
      end

      private

      def register_with_client_handler(pid)
        @s_client_handler, @s_acceptor = UNIXSocket.pair
        @s_acceptor.puts registration_data(pid)
        @server.__CHILD__register_acceptor(@s_client_handler)
      end

      def registration_data(pid)
        {type: 'registration', pid: pid, commands: [name, *aliases], description: description}.to_json
      end

      def accept_connection
        terminal = @s_acceptor.recv_io # blocking
        arguments = JSON.parse(@s_acceptor.readline.chomp)

        [terminal, arguments]
      end

      def process_type
        "acceptor"
      end


      # these two methods should be part of the configuration DSL.
      # They're here for now, but I want them out.
      def prefork_action! # TODO : refactor
        ActiveRecord::Base.clear_all_connections! rescue nil
      end

      def postfork_action! # TODO :refactor
        ActiveRecord::Base.establish_connection   rescue nil
        # ActiveSupport::DescendantsTracker.clear   rescue nil
        # ActiveSupport::Dependencies.clear         rescue nil
      end

      module ErrorState
        attr_accessor :error

        def print_error(io, error = @error)
          io.puts "#{error.backtrace[0]}: #{error.message} (#{error.class})"
          error.backtrace[1..-1].each do |line|
            io.puts "\tfrom #{line}"
          end
        end

        def run
          register_with_client_handler(Process.pid)
          Zeus.ui.info "starting error-state acceptor `#{@name}`"

          Thread.new do
            loop do
              terminal = @s_acceptor.recv_io
              _ = @s_acceptor.readline
              @s_acceptor << 0 << "\n"
              print_error(terminal)
              terminal.close
            end
          end
        end

      end
    end
  end
end
