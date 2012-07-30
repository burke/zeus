module Zeus
  class Server
    class AcceptorRegistrationMonitor

      def datasource          ; @reg_monitor ; end
      def on_datasource_event ; handle_registration ; end

      def initialize
        @reg_monitor, @reg_acceptor = UNIXSocket.pair
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
