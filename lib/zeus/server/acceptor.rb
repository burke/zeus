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
        postfork_action!
        @s_acceptor << $$ << "\n"
        $stdin.reopen(terminal)
        $stdout.reopen(terminal)
        $stderr.reopen(terminal)
        ARGV.replace(arguments)

        @action.call
      ensure
        dnw, dnr = File.open("/dev/null", "w+"), File.open("/dev/null", "r+")
        $stdin.reopen(dnw)
        $stdout.reopen(dnr)
        $stderr.reopen(dnr)
        terminal.close
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

    end
  end
end
