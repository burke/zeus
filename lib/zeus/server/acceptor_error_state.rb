require 'json'
require 'socket'

# See Zeus::Server::ClientHandler for relevant documentation
module Zeus
  class Server
    module AcceptorErrorState
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

