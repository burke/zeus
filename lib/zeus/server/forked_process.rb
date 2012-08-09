module Zeus
  class Server
    # base class for Stage and Acceptor
    class ForkedProcess
      HasNoChildren = Class.new(Exception)

      attr_accessor :name
      attr_reader :pid
      def initialize(server)
        @server = server
      end

      def notify_feature(feature)
        @server.__CHILD__stage_has_feature(@name, feature)
      end

      def descendent_acceptors
        raise NotImplementedError
      end

      def process_type
        raise "NotImplementedError"
      end

      def notify_started
        @server.__CHILD__stage_starting_with_pid(@name, Process.pid)
        Zeus.ui.info("starting #{process_type} `#{@name}`")
      end

      def notify_terminated
        # @server.__CHILD__stage_terminating(@name)
        Zeus.ui.info("killing #{process_type} `#{@name}`")
      end

      def setup_forked_process(close_parent_sockets)
        @server.__CHILD__close_parent_sockets if close_parent_sockets

        notify_started

        $0 = "zeus #{process_type}: #{@name}"

        trap("INT") { exit }
        trap("TERM") {
          notify_terminating
          exit
        }

        defined?(ActiveRecord::Base) and ActiveRecord::Base.clear_all_connections!

        new_features = newly_loaded_features()
        $previously_loaded_features = new_features
        Thread.new {
          new_features.each { |f| notify_feature(f) }
        }
      end

      def newly_loaded_features
        old_features = defined?($previously_loaded_features) ? $previously_loaded_features : []
        ($LOADED_FEATURES + @server.extra_features) - old_features
      end

      def kill_pid_on_exit(pid)
        currpid = Process.pid
        at_exit { Process.kill(9, pid) if Process.pid == currpid rescue nil }
      end

      def runloop!
        raise NotImplementedError
      end

      def before_setup
      end

      def after_setup
      end

      def run(close_parent_sockets = false)
        @pid = fork {
          before_setup
          setup_forked_process(close_parent_sockets)

          Zeus.run_after_fork!

          after_setup
          runloop!
        }
        kill_pid_on_exit(@pid)
        @pid
      end

    end

  end
end

