module Zeus
  class Server
    class AcceptorRegistrationMonitor

      def datasource          ; @sock ; end
      def on_datasource_event ; handle_message ; end
      # @__CHILD__sock is not closed here, as it's used by the master to respond
      # on behalf of unbooted acceptors
      def close_child_socket  ; end
      def close_parent_socket ; @sock.close ; end

      def initialize
        @sock, @__CHILD__sock = UNIXSocket.pair
        @acceptors = []
        @pings = {}
      end

      AcceptorStub = Struct.new(:pid, :socket, :commands, :description)

      def handle_message
        io = @sock.recv_io

        data = JSON.parse(io.readline.chomp)
        type = data['type']

        case type
        when 'wait' ; handle_wait(io, data)
        when 'registration' ; handle_registration(io, data)
        else raise "invalid message"
        end
      end

      def handle_wait(io, data)
        command = data['command'].to_s
        @pings[command] ||= []
        @pings[command] << io
      end

      def handle_registration(io, data)
        pid         = data['pid'].to_i
        commands    = data['commands']
        description = data['description']

        @acceptors.reject!{|ac|ac.commands == commands}
        @acceptors << AcceptorStub.new(pid, io, commands, description)
        notify_pings_for_commands(commands)
      end

      def notify_pings_for_commands(commands)
        (commands || []).each do |command|
          (@pings[command.to_s] || []).each do |ping|
            ping.puts "ready\n"
            ping.close
          end
          @pings[command.to_s] = nil
        end
      end

      module ChildProcessApi

        def __CHILD__find_acceptor_for_command(command)
          @acceptors.detect { |acceptor|
            acceptor.commands.include?(command)
          }
        end

        def __CHILD__register_acceptor(io)
          @__CHILD__sock.send_io(io)
        end

      end ; include ChildProcessApi

    end
  end
end
