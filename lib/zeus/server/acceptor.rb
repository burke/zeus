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

      def run
        fork {
          $0 = "zeus acceptor: #{@name}"
          pid = Process.pid

          register_with_client_handler(pid)

          @server.w_pid "#{pid}:#{Process.ppid}"

          Zeus.ui.as_zeus "starting acceptor `#{@name}`"
          trap("INT") {
            Zeus.ui.as_zeus "killing acceptor `#{@name}`"
            exit 0
          }

          $LOADED_FEATURES.each do |f|
            @server.w_feature "#{pid}:#{f}"
          end

          loop do
            prefork_action!
            terminal = @s_acceptor.recv_io
            arguments = JSON.parse(@s_acceptor.readline.chomp)
            child = fork do
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
