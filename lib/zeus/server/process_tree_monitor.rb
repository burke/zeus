module Zeus
  class Server
    class ProcessTreeMonitor
      PID_TYPE = "P"
      FEATURE_TYPE = "F"

      def datasource          ; @sock ; end
      def on_datasource_event ; handle_messages ; end
      def close_child_socket  ; @__CHILD__sock.close ; end
      def close_parent_socket ; @sock.close ; end

      def initialize(file_monitor)
        @tree = ProcessTree.new
        @file_monitor = file_monitor

        @sock, @__CHILD__sock = Socket.pair(:UNIX, :DGRAM)
      end

      def kill_nodes_with_feature(file)
        @tree.kill_nodes_with_feature(file)
      end

      def handle_messages
        50.times { handle_message }
      rescue Errno::EAGAIN
      end

      def handle_message
        data = @sock.recv_nonblock(1024)
        case data[0]
        when FEATURE_TYPE
          handle_feature_message(data[1..-1])
        when PID_TYPE
          handle_pid_message(data[1..-1])
        else
          raise "Unrecognized message"
        end
      end

      def handle_pid_message(data)
        data =~ /(\d+):(\d+)/
          pid, ppid = $1.to_i, $2.to_i
        @tree.process_has_parent(pid, ppid)
      end

      def handle_feature_message(data)
        data =~ /(\d+):(.*)/
          pid, file = $1.to_i, $2
        @tree.process_has_feature(pid, file)
        @file_monitor.watch(file)
      end


      module ChildProcessApi
        def __CHILD__send_pid(message)
          @__CHILD__sock.send(PID_TYPE + message, 0)
        rescue Errno::ENOBUFS
          sleep 0.2
          retry
        end

        def __CHILD__send_feature(message)
          @__CHILD__sock.send(FEATURE_TYPE + message, 0)
        rescue Errno::ENOBUFS
          sleep 0.2
          retry
        end
      end ; include ChildProcessApi

    end
  end
end
