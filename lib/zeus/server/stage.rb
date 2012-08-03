module Zeus
  class Server
    # NONE of the code in the module is run in the master process,
    # so every communication to the master must be done with IPC.
    class Stage
      HasNoChildren = Class.new(Exception)

      attr_accessor :name, :stages, :actions
      attr_reader :pid
      def initialize(server)
        @server = server
      end

      def notify_feature(feature)
        @server.__CHILD__pid_has_feature(Process.pid, feature)
      end

      def descendent_acceptors
        @stages.map(&:descendent_acceptors).flatten
      end

      def register_acceptors_as_errors(e)
        descendent_acceptors.each do |acc|
          acc.run_as_error(e)
        end
      end


      def handle_load_error(e)
        errored_file = e.backtrace[0].scan(/(.+?):\d+:in/)[0][0]

        # handle relative paths
        unless errored_file =~ /^\//
          errored_file = File.expand_path(errored_file, Dir.pwd)
        end

        register_acceptors_as_errors(e)
        # register all the decendent acceptors as stubs with errors

        notify_feature(errored_file)
        $LOADED_FEATURES.each { |f| notify_feature(f) }

        # we do not need to do anything. We wait, until a dependency changes.
        # At that point, we get killed and restarted.
        sleep
      end

      def run(close_parent_sockets = false)
        @pid = fork {
          # This is only passed to the top-level stage, from Server#run, not sub-stages.
          @server.__CHILD__close_parent_sockets if close_parent_sockets

          $0 = "zeus spawner: #{@name}"
          @server.__CHILD__pid_has_ppid(Process.pid, Process.ppid)

          Zeus.ui.as_zeus("starting spawner `#{@name}`")
          trap("INT") {
            Zeus.ui.as_zeus("killing spawner `#{@name}`")
            exit 0
          }

          begin
            @actions.each(&:call)
          rescue => e
            handle_load_error(e)
          end

          pids = {}
          @stages.each do |stage|
            pids[stage.run] = stage
          end

          Thread.new {
            $LOADED_FEATURES.each { |f| notify_feature(f) }
          }

          loop do
            begin
              pid = Process.wait
            rescue Errno::ECHILD
              raise HasNoChildren.new("Stage `#{@name}` - All terminal nodes must be acceptors")
            end
            stage = pids[pid]
            pids[stage.run] = stage
          end
        }
        currpid = Process.pid
        at_exit { Process.kill(9, @pid) if Process.pid == currpid rescue nil }
        @pid
      end

    end

  end
end
