module Zeus
  class Server
    class AcceptorRegistrationMonitor

      def datasource          ; @reg_monitor ; end
      def on_datasource_event ; handle_registration ; end

      def initialize
        # note: if these aren't ivars, they go out of scope, get GC'd,
        # and cause the UNIXSockets to quit working... often in perplexing ways.
        @s, @r = Socket.pair(:UNIX, :DGRAM)
        @reg_monitor  = UNIXSocket.for_fd(@s.fileno)
        @reg_acceptor = UNIXSocket.for_fd(@r.fileno)
        @acceptors = []
      end

      AcceptorStub = Struct.new(:pid, :socket, :commands, :description)

      def handle_registration
        io = @reg_monitor.recv_io

        data = JSON.parse(io.readline.chomp)
        pid         = data['pid'].to_i
        commands    = data['commands']
        description = data['description']

        @acceptors.reject!{|ac|ac.commands == commands}
        @acceptors << AcceptorStub.new(pid, io, commands, description)
      end

      def find_acceptor_for_command(command)
        @acceptors.detect { |acceptor|
          acceptor.commands.include?(command)
        }
      end

      def acceptor_registration_socket
        @reg_acceptor
      end

    end

  end
end
