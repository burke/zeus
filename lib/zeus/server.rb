require 'json'
require 'socket'

module Zeus
  module Server
    def self.define!(&b)
      @@root = Stage.new("(root)")
      @@root.instance_eval(&b)
    end

    def self.run
      trap("INT") { exit 0 }
      @@root_stage_pid = @@root.run
      Process.waitall
    end

    class Stage
      def initialize(name)
        @name = name
        @stages, @actions = [], []
      end

      def action(&b)
        @actions << b
      end

      def stage(name, &b)
        @stages << Stage.new(name).tap { |s| s.instance_eval(&b) }
      end

      def acceptor(name, socket, &b)
        @stages << Acceptor.new(name, socket, &b)
      end

      # There are a few things we want to accomplish:
      # 1. Running all the actions (each time this stage is killed and restarted)
      # 2. Starting all the substages (and restarting them when necessary)
      # 3. Starting all the acceptors (and restarting them when necessary)
      def run
        fork {
          $0 = "zeus spawner: #{@name}"
          puts "\x1b[35m[zeus] starting spawner `#{@name}`\x1b[0m"

          @actions.each(&:call)

          pids = {}
          @stages.each do |stage|
            pids[stage] = stage.run
          end

          loop do
            pid = Process.wait
            if (status = $?.exitstatus) > 0
              exit status
            else # restart the stage that died.
              stage = pids[pid]
              pids[stage] = stage.run
            end
          end

        }
      end

    end

    class Acceptor
      def initialize(name, socket, &b)
        @name = name
        @socket = socket
        @action = b
      end

      def run
        fork {
          $0 = "zeus acceptor: #{@name}"
          puts "\x1b[35m[zeus] starting acceptor `#{@name}`\x1b[0m"

          File.unlink(@socket) rescue nil
          server = UNIXServer.new(@socket)
          loop do
            ActiveRecord::Base.clear_all_connections! # TODO : refactor
            client = server.accept
            child = fork do
              ActiveRecord::Base.establish_connection # TODO :refactor
              ActiveSupport::DescendantsTracker.clear
              ActiveSupport::Dependencies.clear

              terminal = client.recv_io
              arguments = JSON.load(client.gets.strip)['arguments']

              client << {status: "OK", pid: $$}.to_json << "\n"
              $stdin.reopen(terminal)
              $stdout.reopen(terminal)
              $stderr.reopen(terminal)
              ARGV.replace(arguments)

              @action.call
            end
            Process.detach(child)
            client.close
          end
        }
      end

    end

  end
end
