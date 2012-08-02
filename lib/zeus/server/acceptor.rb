require 'json'
require 'socket'

# See Zeus::Server::ClientHandler for relevant documentation
module Zeus
  class Server
    class Acceptor

      attr_accessor :name, :aliases, :description, :action
      def initialize(server)
        @server = server
        @client_handler = server.client_handler
        @registration_monitor = server.acceptor_registration_monitor
      end

      def register_with_client_handler(pid)
        @s_client_handler, @s_acceptor = UNIXSocket.pair

        @s_acceptor.puts registration_data(pid)

        @registration_monitor.acceptor_registration_socket.send_io(@s_client_handler)
      end

      def registration_data(pid)
        {pid: pid, commands: [name, *aliases], description: description}.to_json
      end

      def descendent_acceptors
        self
      end

      def print_error(io, error)
        io.puts "#{error.backtrace[0]}: #{error.message} (#{error.class})"
        error.backtrace[1..-1].each do |line|
          io.puts "\tfrom #{line}"
        end
      end

      def run_as_error(e)
        register_with_client_handler(Process.pid)
        Zeus.ui.as_zeus "starting error-state acceptor `#{@name}`"

        Thread.new do
          loop do
            terminal = @s_acceptor.recv_io
            _ = @s_acceptor.readline
            @s_acceptor << 0 << "\n"
            print_error(terminal, e)
            terminal.close
          end
        end
      end

      def run
        pid = fork {
          $0 = "zeus acceptor: #{@name}"
          pid = Process.pid

          register_with_client_handler(pid)

          @server.w_pid "#{pid}:#{Process.ppid}"

          Zeus.ui.as_zeus "starting acceptor `#{@name}`"
          trap("INT") {
            Zeus.ui.as_zeus "killing acceptor `#{@name}`"
            exit 0
          }

          # Apparently threads don't continue in forks.
          Thread.new {
            $LOADED_FEATURES.each do |f|
              @server.w_feature "#{pid}:#{f}"
            end
          }

          loop do
            prefork_action!
            terminal = @s_acceptor.recv_io
            arguments = JSON.parse(@s_acceptor.readline.chomp)
            child = fork do
              $0 = "zeus runner: #{@name}"
              postfork_action!
              @s_acceptor << $$ << "\n"
              $stdin.reopen(terminal)
              $stdout.reopen(terminal)
              $stderr.reopen(terminal)
              ARGV.replace(arguments)

              @action.call
            end
            Process.detach(child)
            terminal.close
          end
        }
        currpid = Process.pid
        at_exit { Process.kill(9, pid) if Process.pid == currpid rescue nil }
        pid
      end

      def prefork_action! # TODO : refactor
        ActiveRecord::Base.clear_all_connections!
      end

      def postfork_action! # TODO :refactor
        ActiveRecord::Base.establish_connection
        ActiveSupport::DescendantsTracker.clear
        ActiveSupport::Dependencies.clear
      end

    end
  end
end
