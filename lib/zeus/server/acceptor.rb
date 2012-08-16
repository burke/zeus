require 'json'
require 'socket'

# See Zeus::Server::ClientHandler for relevant documentation
module Zeus
  class Server
    class Acceptor
      attr_accessor :aliases, :description, :action, :name
      def initialize(server)
        @server = server
      end

      def descendent_acceptors
        self
      end

      def run
        register_with_client_handler(Process.pid)
        Zeus.ui.info("starting #{process_type} `#{@name}`")

        thread_with_backtrace_on_error { runloop! }
      end

      private

      def command_runner
        CommandRunner.new(name, action, @s_acceptor)
      end

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

      def print_error(io, error)
        io.puts "#{error.backtrace[0]}: #{error.message} (#{error.class})"
        error.backtrace[1..-1].each do |line|
          io.puts "\tfrom #{line}"
        end
      end

      def thread_with_backtrace_on_error(&b)
        Thread.new {
          begin
            b.call
          rescue => e
            print_error($stdout, e)
          end
        }
      end

      def runloop!
        loop do
          terminal, arguments = accept_connection # blocking
          command_runner.run(terminal, arguments)
        end
      end

      module ErrorState
        NOT_A_PID = 0
        attr_accessor :error

        def process_type
          "error-state acceptor"
        end

        def runloop!
          loop do
            terminal = @s_acceptor.recv_io
            _ = @s_acceptor.readline
            @s_acceptor << NOT_A_PID << "\n"
            print_error(terminal, @error)
            terminal.close
          end
        end

      end
    end
  end
end
