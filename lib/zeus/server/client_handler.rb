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

      def initialize(acceptor_commands, server)
        @server = server
        @acceptor_commands = acceptor_commands
        @listener = UNIXServer.new(Zeus::SOCKET_NAME)
        @listener.listen(10)
      rescue Errno::EADDRINUSE
        Zeus.ui.error "Zeus appears to be already running in this project. If not, remove .zeus.sock and try again."
        exit 1
      end

      def handle_server_connection
        s_client = @listener.accept

        # 1
        data = JSON.parse(s_client.readline.chomp)
        command, arguments = data.values_at('command', 'arguments')

        # 2
        client_terminal = s_client.recv_io

        Thread.new {
          loop do
            pid = fork { handshake_client_to_acceptor(s_client, command, arguments, client_terminal) ; exit }
            Process.wait(pid)
            break unless $?.exitstatus == REATTEMPT_HANDSHAKE
          end
        }
      end

      REATTEMPT_HANDSHAKE = 204

      NoSuchCommand = Class.new(Exception)
      AcceptorNotBooted = Class.new(Exception)
      ApplicationLoadFailed = Class.new(Exception)

      def exit_with_message(s_client, client_terminal, msg)
        s_client << "0\n"
        client_terminal << "[zeus] #{msg}\n"
        client_terminal.close
        s_client.close
      end

      def wait_for_acceptor(s_client, client_terminal, command, msg)
        s_client << "0\n"
        client_terminal << "[zeus] #{msg}\n"

        regmsg = {type: 'wait', command: command}

        s, r = UNIXSocket.pair

        @server.__CHILD__register_acceptor(r)
        s << "#{regmsg.to_json}\n"

        s.readline # wait
        s.close

        exit REATTEMPT_HANDSHAKE
      end

      #  client clienthandler acceptor
      # 1  ---------->                | {command: String, arguments: [String]}
      # 2  ---------->                | Terminal IO
      # 3            ----------->     | Terminal IO
      # 4            ----------->     | Arguments (json array)
      # 5            <-----------     | pid
      # 6  <---------                 | pid
      def handshake_client_to_acceptor(s_client, command, arguments, client_terminal)
        # 3
        unless @acceptor_commands.include?(command.to_s)
          return exit_with_message(
            s_client, client_terminal,
            "no such command `#{command}`.")
        end
        acceptor = @server.__CHILD__find_acceptor_for_command(command)
        unless acceptor
          wait_for_acceptor(
            s_client, client_terminal, command,
            "waiting for `#{command}` to finish booting...")
        end
        usock = UNIXSocket.for_fd(acceptor.socket.fileno)
        if usock.closed?
          wait_for_acceptor(
            s_client, client_terminal, command,
            "waiting for `#{command}` to finish reloading dependencies...")
        end
        begin
          usock.send_io(client_terminal)
        rescue Errno::EPIPE
          wait_for_acceptor(
            s_client, client_terminal, command,
            "waiting for `#{command}` to finish reloading dependencies...")
        end


        Zeus.ui.info "accepting connection for #{command}"

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
