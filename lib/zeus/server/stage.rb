module Zeus
  class Server
    class Stage
      HasNoChildren = Class.new(Exception)

      attr_accessor :name, :stages, :actions
      attr_reader :pid
      def initialize(server)
        @server = server
      end

      # There are a few things we want to accomplish:
      # 1. Running all the actions (each time this stage is killed and restarted)
      # 2. Starting all the substages (and restarting them when necessary)
      # 3. Starting all the acceptors (and restarting them when necessary)
      def run
        @pid = fork {
          $0 = "zeus spawner: #{@name}"
          pid = Process.pid
          @server.w_pid "#{pid}:#{Process.ppid}"

          Zeus.ui.as_zeus("starting spawner `#{@name}`")
          trap("INT") {
            Zeus.ui.as_zeus("killing spawner `#{@name}`")
            exit 0
          }

          @actions.each(&:call)

          pids = {}
          @stages.each do |stage|
            pids[stage.run] = stage
          end

          $LOADED_FEATURES.each do |f|
            @server.w_feature "#{pid}:#{f}"
          end

          loop do
            begin
              pid = Process.wait
            rescue Errno::ECHILD
              raise HasNoChildren.new("Stage `#{@name}` - All terminal nodes must be acceptors")
            end
            if (status = $?.exitstatus) > 0
              exit status
            else # restart the stage that died.
              stage = pids[pid]
              pids[stage.run] = stage
            end
          end

        }
      end

    end

  end
end
