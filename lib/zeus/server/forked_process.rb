module Zeus
  class Server
    class ForkedProcess

      attr_accessor :name
      attr_reader :pid
      def initialize(server)
        @server = server
      end

      def notify_feature(feature)
        @server.__CHILD__stage_has_feature(@name, feature)
      end

      def notify_started
        @server.__CHILD__stage_starting_with_pid(@name, Process.pid)
        Zeus.ui.info("starting #{process_type} `#{@name}`")
      end

      def notify_terminated
        Zeus.ui.info("killing #{process_type} `#{@name}`")
      end

      def setup_forked_process(close_parent_sockets)
        @server.__CHILD__close_parent_sockets if close_parent_sockets

        notify_started

        $0 = "zeus #{process_type}: #{@name}"

        trap("INT") { exit }
        trap("TERM") {
          notify_terminated
          exit
        }

        defined?(ActiveRecord::Base) and ActiveRecord::Base.clear_all_connections!

      end

      def newly_loaded_features
        old_features = defined?($previously_loaded_features) ? $previously_loaded_features : []
        ($LOADED_FEATURES + @server.extra_features) - old_features
      end

      def notify_new_features
        new_features = newly_loaded_features()
        $previously_loaded_features ||= []
        $previously_loaded_features |= new_features
        Thread.new {
          new_features.each { |f| notify_feature(f) }
        }
      end

    end

  end
end

