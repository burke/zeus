require 'socket'
require 'json'

module Zeus
  class Server
    # The model here is kind of convoluted, so here's an explanation of what's
    # happening with all these sockets:
    #
    # #### Acceptor Registration
    # 1. ClientHandler creates a socketpair for Acceptor registration (S_REG)
    # 2. When an acceptor is spawned, it:
    #   1. Creates a new socketpair for communication with clienthandler (S_ACC)
    #   2. Sends one side of S_ACC over S_REG to clienthandler.
    #   3. Sends a JSON-encoded hash of `pid`, `commands`, and `description`. over S_REG.
    # 3. ClientHandler received first the IO and then the JSON hash, and stores them for later reference.
    #
    # #### Running a command
    # 1. ClientHandler has a UNIXServer (SVR) listening.
    # 2. ClientHandler has a socketpair with the acceptor referenced by the command (see Registration) (S_ACC)
    # 3. When clienthandler received a connection (S_CLI) on SVR:
    #   1. ClientHandler sends S_CLI over S_ACC, so the acceptor can communicate with the server's client.
    #   2. ClientHandler sends a JSON-encoded array of `arguments` over S_ACC
    #   3. Acceptor sends the newly-forked worker PID over S_ACC to clienthandler.
    #   4. ClientHandler forwards the pid to the client over S_CLI.
    #
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
        Zeus.ui.error "Zeus appears to be already running in this project. If not, remove .zeus.sock and try again."
        exit 1
      end

      #  client clienthandler acceptor
      # 1  ---------->                | {command: String, arguments: [String]}
      # 2  ---------->                | Terminal IO
      # 3            ----------->     | Terminal IO
      # 4            ----------->     | Arguments (json array)
      # 5            <-----------     | pid
      # 6  <---------                 | pid
      def handle_server_connection
        s_client = @listener.accept

        data = JSON.parse(s_client.readline.chomp) # step 1
        command, arguments = data.values_at('command', 'arguments')

        client_terminal = s_client.recv_io # step 2

        Thread.new {
          # This is a little ugly. Gist: Try to handshake the client to the acceptor.
          # If the acceptor is not booted yet, this will hang until it is, then terminate with 
          # REATTEMPT_HANDSHAKE. We catch that exit code and try once more.
          loop do
            pid = fork { handshake_client_to_acceptor(s_client, command, arguments, client_terminal) ; exit }
            Process.wait(pid)
            break if $?.exitstatus != REATTEMPT_HANDSHAKE
          end
        }
      end

      def handshake_client_to_acceptor(s_client, command, arguments, client_terminal)
        unless @acceptor_commands.include?(command.to_s)
          msg = "no such command `#{command}`."
          return exit_with_message(s_client, client_terminal, msg)
        end

        unless acceptor = send_io_to_acceptor(client_terminal, command) # step 3
          wait_for_acceptor(s_client, client_terminal, command)
          exit REATTEMPT_HANDSHAKE
        end

        Zeus.ui.info "accepting connection for #{command}"

        acceptor.socket.puts arguments.to_json # step 4
        pid = acceptor.socket.readline.chomp.to_i # step 5
        s_client.puts pid # step 6
      end

      private

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
        return acceptor
      rescue Errno::EPIPE
        return false
      end


    end
  end
end
