module Zeus
  class Server
    class ProcessTreeMonitor
      STARTING_MARKER = "P"
      FEATURE_MARKER  = "F"

      def datasource          ; @sock ; end
      def on_datasource_event ; handle_messages ; end
      def close_child_socket  ; @__CHILD__sock.close ; end
      def close_parent_socket ; @sock.close ; end

      def initialize(file_monitor, tree)
        @root = tree
        @file_monitor = file_monitor

        @sock, @__CHILD__sock = open_socketpair
      end

      def kill_nodes_with_feature(file)
        @root.kill_nodes_with_feature(file)
      end

      module ChildProcessApi
        def __CHILD__stage_starting_with_pid(name, pid)
          buffer_send("#{STARTING_MARKER}#{name}:#{pid}")
        end

        def __CHILD__stage_has_feature(name, feature)
          buffer_send("#{FEATURE_MARKER}#{name}:#{feature}")
        end

        private

        def buffer_send(msg)
          @__CHILD__sock.send(msg, 0)
        rescue Errno::ENOBUFS
          sleep 0.2
          retry
        end

      end ; include ChildProcessApi

      private

      def handle_messages
        50.times { handle_message }
      rescue Errno::EAGAIN
      end

      def handle_message
        data = @sock.recv_nonblock(4096)
        case data[0]
        when STARTING_MARKER
          handle_starting_message(data[1..-1])
        when FEATURE_MARKER
          handle_feature_message(data[1..-1])
        end
      end

      def open_socketpair
        Socket.pair(:UNIX, :DGRAM)
      end

      def handle_starting_message(data)
        data =~ /(.+):(\d+)/
        name, pid = $1.to_sym, $2.to_i
        @root.stage_has_pid(name, pid)
      end

      def handle_feature_message(data)
        data =~ /(.+?):(.*)/
        name, file = $1.to_sym, $2
        @root.stage_has_feature(name, file)
        @file_monitor.watch(file)
      end


    end
  end
end
