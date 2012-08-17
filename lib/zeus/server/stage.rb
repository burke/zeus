module Zeus
  class Server
    # NONE of the code in the module is run in the master process,
    # so every communication to the master must be done with IPC.
    class Stage
      autoload :ErrorState,      'zeus/server/stage/error_state'
      autoload :FeatureNotifier, 'zeus/server/stage/feature_notifier'

      attr_accessor :name, :stages, :actions
      def initialize(server)
        @server = server
      end

      def descendent_acceptors
        @stages.map(&:descendent_acceptors).flatten
      end

      def run(close_parent_sockets = false)
        @pid = fork {
          setup_fork(close_parent_sockets)
          run_actions
          feature_notifier.notify_new_features
          start_child_stages
          handle_child_exit_loop!
        }
      end

      private

      def setup_fork(close_parent_sockets)
        $0 = "zeus #{process_type}: #{@name}"
        @server.__CHILD__close_parent_sockets if close_parent_sockets
        notify_started
        trap("INT") { exit }
        trap("TERM") { notify_terminated ; exit }
        ActiveRecord::Base.clear_all_connections! rescue nil
      end

      def feature_notifier
        FeatureNotifier.new(@server, @name)
      end

      def start_child_stages
        @pids = {}
        @stages.each do |stage|
          @pids[stage.run] = stage
        end
      end

      def run_actions
        begin
          @actions.each(&:call)
        rescue => e
          extend(ErrorState)
          handle_load_error(e)
        end
      end

      def handle_child_exit_loop!
        loop do
          begin
            pid = Process.wait
          rescue Errno::ECHILD
            sleep # if this is a terminal node, just let acceptors run...
          end
          stage = @pids[pid]
          @pids[stage.run] = stage
        end
      end

      def notify_started
        @server.__CHILD__stage_starting_with_pid(@name, Process.pid)
        Zeus.ui.info("starting #{process_type} `#{@name}`")
      end

      def notify_terminated
        Zeus.ui.info("killing #{process_type} `#{@name}`")
      end


      def process_type
        "spawner"
      end

    end
  end
end
