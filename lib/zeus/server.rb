require 'json'
require 'socket'

require 'rb-kqueue'
require 'zeus/process'

module Zeus
  module Server
    def self.define!(&b)
      @@root = Stage.new("(root)")
      @@root.instance_eval(&b)
      @@files = {}
    end

    def self.pid_has_file(pid, file)
      @@files[file] ||= []
      @@files[file] << pid
    end

    def self.killall_with_file(file)
      pids = @@files[file]
      puts "\x1b[31mNot implemented: Here's the part where we'd kill any process that's required that file...\x1b[0m"
      # @@root.kill_pids(pids)
      # pids.each do |pid|
      #   begin
      #     # TODO: lots of dangling pids in @@files[<x>]
      #     Process.kill("INT", pid)
      #   rescue Errno::ESRCH
      #   end
      # end
    end

    TARGET_FD_LIMIT = 8192

    def self.configure_number_of_file_descriptors
      limit = Process.getrlimit(Process::RLIMIT_NOFILE)
      if limit[0] < TARGET_FD_LIMIT && limit[1] >= TARGET_FD_LIMIT
        Process.setrlimit(Process::RLIMIT_NOFILE, TARGET_FD_LIMIT)
      else
        puts "\x1b[33m[zeus] Warning: increase the max number of file descriptors. If you have a large project, this max cause a crash in about 10 seconds.\x1b[0m"
      end
    end

    def self.run
      $0 = "zeus master"
      configure_number_of_file_descriptors
      trap("INT") { exit 0 }
      at_exit { Process.killall_descendants(9) }

      $r, $w = IO.pipe
      $w.sync = true

      @@root_stage_pid = @@root.run

      queue = KQueue::Queue.new

      lost_files = []

      notify = ->(event){
        if event.flags.include?(:delete)
          # file was deleted, so we need to close and reopen it.
          event.watcher.disable!
          begin
            queue.watch_file(event.watcher.path, :write, :extend, :rename, :delete, &notify)
          rescue Errno::ENOENT
            lost_files << event.watcher.path
          end
        end
        puts "\x1b[37m#{event.watcher.path}\x1b[0m"
        killall_with_file(event.watcher.path)
      }

      files = {}
      loop do
        queue.poll

        # TODO: It would be really nice if we could put the queue poller in the select somehow.
        #   --investigate kqueue. Is this possible?
        rs, _, _ = IO.select([$r], [], [], 1)
        if rs && r = rs[0]
          data = r.readline
          data =~ /(\d+):(.*)/
          pid, file = $1, $2
          pid_has_file($1.to_i, $2)
          if files[file] == nil
            files[file] = true
            begin
              queue.watch_file(file.chomp, :write, :extend, :rename, :delete, &notify)
            rescue Errno::EMFILE
              print_ulimit_message
              exit 1
            rescue Errno::ENOENT
              puts "No file found at #{file.chomp}"
            end
          end
        end
      end

    end

    class Stage
      attr_reader :pid
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
        @pid = fork {
          $0 = "zeus spawner: #{@name}"
          pid = Process.pid
          $LOADED_FEATURES.each do |f|
            $w.puts "#{pid}:#{f}\n"
          end
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
      attr_reader :pid
      def initialize(name, socket, &b)
        @name = name
        @socket = socket
        @action = b
      end

      def run
        @pid = fork {
          $0 = "zeus acceptor: #{@name}"
          pid = Process.pid
          $LOADED_FEATURES.each do |f|
            $w.puts "#{pid}:#{f}\n"
          end
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
              arguments = JSON.load(client.gets.strip)

              client << $$ << "\n"
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
