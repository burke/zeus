module Zeus
  class Server
    # NONE of the code in the module is run in the master process,
    # so every communication to the master must be done with IPC.
    class Stage < ForkedProcess
      attr_accessor :stages, :actions

      def descendent_acceptors
        @stages.map(&:descendent_acceptors).flatten
      end


      def after_setup
        begin
          @actions.each(&:call)
        rescue => e
          handle_load_error(e)
        end

        @pids = {}
        @stages.each do |stage|
          @pids[stage.run] = stage
        end
      end

      def runloop!
        loop do
          begin
            pid = Process.wait
          rescue Errno::ECHILD
            raise HasNoChildren.new("Stage `#{@name}` - All terminal nodes must be acceptors")
          end
          stage = @pids[pid]
          @pids[stage.run] = stage
        end
      end


      private

      def register_acceptors_as_errors(e)
        descendent_acceptors.each do |acc|
          acc = acc.extend(Acceptor::ErrorState)
          acc.error = e
          acc.run
        end
      end

      def process_type
        "spawner"
      end

      def full_path_of_file_from_error(e)
        errored_file = e.backtrace[0].scan(/(.+?):\d+:in/)[0][0]

        # handle relative paths
        unless errored_file =~ /^\//
          errored_file = File.expand_path(errored_file, Dir.pwd)
        end
      end

      def handle_load_error(e)
        errored_file = full_path_of_file_from_error(e)

        # register all the decendent acceptors as stubs with errors
        register_acceptors_as_errors(e)

        notify_feature(errored_file)
        $LOADED_FEATURES.each { |f| notify_feature(f) }

        # we do not need to do anything. We wait, until a dependency changes.
        # At that point, we get killed and restarted.
        sleep
      end

    end
  end
end
