require 'socket'
require 'json'

# This class is a little confusing. See the docs/ directory for guidance.
module Zeus
  class Server
    class ClientHandler
      def datasource          ; @listener ; end
      def on_datasource_event ; handle_server_connection ; end
      def close_child_socket  ; end
      def close_parent_socket ; @listener.close ; end

      REATTEMPT_HANDSHAKE = 204

      def initialize(acceptor_commands, server)
        @server = server
        @acceptor_commands = acceptor_commands
        @listener = UNIXServer.new(Zeus::SOCKET_NAME)
        @listener.listen(10)
      rescue Errno::EADDRINUSE
        Zeus.ui.error "Zeus appears to be already running in this project. If not, remove #{Zeus::SOCKET_NAME} and try again."
        exit 1
      end

      private

      # See docs/client_server_handshake.md for details
      def handle_server_connection
        s_client = @listener.accept

        data = JSON.parse(s_client.readline.chomp)
        command, arguments = data.values_at('command', 'arguments')

        client_terminal = s_client.recv_io

        Thread.new {
          # This is a little ugly. Gist: Try to handshake the client to the acceptor.
          # If the acceptor is not booted yet, this will hang until it is, then terminate with 
          # REATTEMPT_HANDSHAKE. We catch that exit code and try once more.
          begin
            loop do
              pid = fork { handshake_client_to_acceptor(s_client, command, arguments, client_terminal) ; exit }
              Process.wait(pid)
              break if $?.exitstatus != REATTEMPT_HANDSHAKE
            end
          ensure
            client_terminal.close
            s_client.close
          end
        }
      end

      def handshake_client_to_acceptor(s_client, command, arguments, client_terminal)
        unless @acceptor_commands.include?(command.to_s)
          msg = "no such command `#{command}`."
          return exit_with_message(s_client, client_terminal, msg)
        end

        unless acceptor = send_io_to_acceptor(client_terminal, command)
          wait_for_acceptor(s_client, client_terminal, command)
          exit REATTEMPT_HANDSHAKE
        end

        Zeus.ui.info "accepting connection for #{command}"

        acceptor.socket.puts arguments.to_json
        pid = acceptor.socket.readline.chomp.to_i
        s_client.puts pid
        s_client.close
      end

      def exit_with_message(s_client, client_terminal, msg)
        s_client << "0\n"
        client_terminal << "[zeus] #{msg}\n"
        client_terminal.close
        s_client.close
        exit 1
      end

      def wait_for_acceptor(s_client, client_terminal, command)
        s_client << "0\n"
        client_terminal << "[zeus] waiting for `#{command}` to finish booting...\n"

        s, r = UNIXSocket.pair
        s << {type: 'wait', command: command}.to_json << "\n"
        @server.__CHILD__register_acceptor(r)

        s.readline # wait until acceptor is booted
      end

      def send_io_to_acceptor(io, command)
        return false unless acceptor = @server.__CHILD__find_acceptor_for_command(command)
        return false unless usock = UNIXSocket.for_fd(acceptor.socket.fileno)
        usock.send_io(io)
        io.close
        return acceptor
      rescue Errno::EPIPE
        return false
      end

    end
  end
end
