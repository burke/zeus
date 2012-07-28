require 'socket'
require 'json'

require 'zeus/server/acceptor'

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
      SERVER_SOCK = ".zeus.sock"

      def datasource ; @server ; end
      def on_datasource_event ; handle_server_connection ; end

      def initialize(registration_monitor)
        @reg_monitor = registration_monitor
        begin
          @server = UNIXServer.new(SERVER_SOCK)
          @server.listen(10)
        rescue Errno::EADDRINUSE
          Zeus.ui.error "Zeus appears to be already running in this project. If not, remove .zeus.sock and try again."
          exit 1
        # ensure
        #   @server.close rescue nil
        #   File.unlink(SERVER_SOCK)
        end
      end

      def handle_server_connection
        s_client = @server.accept
        fork { handshake_client_to_acceptor(s_client) }
      end

      #  client clienthandler acceptor
      # 1  ---------->                | {command: String, arguments: [String]}
      # 2  ---------->                | Terminal IO
      # 3            ----------->     | Terminal IO
      # 4            ----------->     | Arguments (json array)
      # 5            <-----------     | pid
      # 6  <---------                 | pid
      def handshake_client_to_acceptor(s_client)
        # 1
        data = JSON.parse(s_client.readline.chomp)
        command, arguments = data.values_at('command', 'arguments')

        # 2
        client_terminal = s_client.recv_io

        # 3
        acceptor = @reg_monitor.find_acceptor_for_command(command)
        usock = UNIXSocket.for_fd(acceptor.socket.fileno)
        # TODO handle nothing found
        usock.send_io(client_terminal)

        puts "accepting connection for #{command}"

        # 4
        acceptor.socket.puts arguments.to_json

        # 5
        pid = acceptor.socket.readline.chomp.to_i

        # 6
        s_client.puts pid
      end

    end
  end
end
