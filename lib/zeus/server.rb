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
      @@process_tree.kill_nodes_with_feature(file)
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

    def self.notify(event)
      if event.flags.include?(:delete)
        # file was deleted, so we need to close and reopen it.
        event.watcher.disable!
        begin
          @@queue.watch_file(event.watcher.path, :write, :extend, :rename, :delete, &method(:notify))
        rescue Errno::ENOENT
          lost_files << event.watcher.path
        end
      end
      puts "\x1b[37m#{event.watcher.path}\x1b[0m"
      killall_with_file(event.watcher.path)
    end

    def self.run
      $0 = "zeus master"
      configure_number_of_file_descriptors
      trap("INT") { exit 0 }
      at_exit { Process.killall_descendants(9) }

      $r_features, $w_features = IO.pipe
      $w_features.sync = true

      $r_pids, $w_pids = IO.pipe
      $w_pids.sync = true

      @@process_tree = ProcessTree.new
      @@root_stage_pid = @@root.run

      @@queue = KQueue::Queue.new

      lost_files = []

      @@file_watchers = {}
      loop do
        @@queue.poll

        # TODO: It would be really nice if we could put the queue poller in the select somehow.
        #   --investigate kqueue. Is this possible?
        rs, _, _ = IO.select([$r_features, $r_pids], [], [], 1)
        rs.each do |r|
          case r
          when $r_pids     ; handle_pid_message(r.readline)
          when $r_features ; handle_feature_message(r.readline)
          end
        end if rs
      end

    end

    class ProcessTree
      class Node
        attr_accessor :pid, :children, :features
        def initialize(pid)
          @pid, @children, @features = pid, [], {}
        end

        def add_child(node)
          self.children << node
        end

        def add_feature(feature)
          self.features[feature] = true
        end

        def has_feature?(feature)
          self.features[feature] == true
        end

        def inspect
          "(#{pid}:#{features.size}:[#{children.map(&:inspect).join(",")}])"
        end

      end

      def inspect
        @root.inspect
      end

      def initialize
        @root = Node.new(Process.pid)
        @nodes_by_pid = {Process.pid => @root}
      end

      def node_for_pid(pid)
        @nodes_by_pid[pid.to_i] ||= Node.new(pid.to_i)
      end

      def process_has_parent(pid, ppid)
        curr = node_for_pid(pid)
        base = node_for_pid(ppid)
        base.add_child(curr)
      end

      def process_has_feature(pid, feature)
        node = node_for_pid(pid)
        node.add_feature(feature)
      end

      def kill_node(node)
        @nodes_by_pid.delete(node.pid)
        # recall that this process explicitly traps INT -> exit 0
        Process.kill("INT", node.pid)
      end

      def kill_nodes_with_feature(file, base = @root)
        if base.has_feature?(file)
          if base == @root.children[0] || base == @root
            puts "\x1b[31mOne of zeus's dependencies changed. Not killing zeus. You may have to restart the server.\x1b[0m"
            return false
          end
          kill_node(base)
          return true
        else
          base.children.dup.each do |node|
            if kill_nodes_with_feature(file, node)
              base.children.delete(node)
            end
          end
          return false
        end
      end

    end

    def self.handle_pid_message(data)
      data =~ /(\d+):(\d+)/
      pid, ppid = $1.to_i, $2.to_i
      @@process_tree.process_has_parent(pid, ppid)
    end

    def self.handle_feature_message(data)
      data =~ /(\d+):(.*)/
      pid, file = $1.to_i, $2
      @@process_tree.process_has_feature(pid, file)
      return if @@file_watchers[file]
      begin
        @@file_watchers[file] = true
        @@queue.watch_file(file.chomp, :write, :extend, :rename, :delete, &method(:notify))
      # rescue Errno::EMFILE
      #   exit 1
      rescue Errno::ENOENT
        puts "No file found at #{file.chomp}"
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
          $w_pids.puts "#{pid}:#{Process.ppid}\n"
          $LOADED_FEATURES.each do |f|
            $w_features.puts "#{pid}:#{f}\n"
          end
          puts "\x1b[35m[zeus] starting spawner `#{@name}`\x1b[0m"
          trap("INT") {
            puts "\x1b[35m[zeus] killing spawner `#{@name}`\x1b[0m"
            exit 0
          }

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
          $w_pids.puts "#{pid}:#{Process.ppid}\n"
          $LOADED_FEATURES.each do |f|
            $w_features.puts "#{pid}:#{f}\n"
          end
          puts "\x1b[35m[zeus] starting acceptor `#{@name}`\x1b[0m"
          trap("INT") {
            puts "\x1b[35m[zeus] killing acceptor `#{@name}`\x1b[0m"
            exit 0
          }

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
