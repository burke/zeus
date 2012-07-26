require "json"
require "socket"

module Zeus

  class Server

    def self.start!
      config = File.read(".zeus.rb")
      new.send(:eval,config)
    end

    trap("INT") { exit 0 }

    class Watchdog
      attr_reader :hierarchy
      def initialize(r, w)
        @r, @w = r, w
        @clients = []
        @hierarchy = ClientHierarchy.new
      end

      class ClientHierarchy
        def initialize
          @tree = {}
        end

        attr_reader :tree

        def add(client)
          hash = find_children_hash_for_pid(client.ppid)
          hash[client] = {}
        end

        def notify_changed(feature, root = @tree, parent = nil)
          root.each do |k, v|
            if k.features.include?(feature)
              root.delete(k)
              puts "\x1b[36m[ZEUS] killing client `#{parent ? parent.tag : "(root)"}` due to dependency change\x1b[0m"
              Process.kill("INT", k.ppid)
            else
              notify_changed(feature, v, k)
            end
          end
        end

        def find_children_hash_for_pid(pid, root = @tree)
          root.each do |k, v|
            return v if k.pid == pid
            if hash = find_children_hash_for_pid(pid, v)
              return hash
            end
          end

          root == @tree ? @tree : nil
        end
      end

      class Client < Struct.new(:tag, :pid, :ppid, :features)

      end

      def add_client(tag, pid, ppid, features)
        client = Client.new(tag, pid, ppid, features)
        @hierarchy.add(client)
        @clients << client
      end

      def register_client(message)
        pid      = message.fetch('pid')
        ppid     = message.fetch('ppid')
        features = message.fetch('features')
        tag      = message.fetch('tag')

        add_client(tag, pid, ppid, features)

        puts "\x1b[36m[ZEUS] started client `#{tag}` (#{features.size} features)\x1b[0m"
      end

      def run!
        loop {
          message = JSON.load(@r.readline.chomp)
          register_client message
        }
      end
    end

    r, $w = IO.pipe

    watchdog_pid = fork {
      $0 = "zeus watchdog"
      watchdog = Watchdog.new(r, $w)
      trap("USR1") {
        puts "\x1b[31mPretending a file in mocha changed... fsevents soon\x1b[0m"
        watchdog.hierarchy.notify_changed("/Users/burke/.rbenv/versions/1.9.3-p125-perf/lib/ruby/gems/1.9.1/gems/oauth-0.4.4/lib/oauth/request_proxy.rb")
      }
      watchdog.run!
    }

    def register_dependency(tag, pid, ppid, features)
      msg = {tag: tag, pid: pid, ppid: ppid, features: features}.to_json
      $w.puts msg
    end

    def forkpoint(forks)
      pids = {}
      forks.each do |k, v|
        pid = fork {
          $0 = "zeus spawner: #{k}"
          v.call
          loop { sleep 100 }
      }
      register_dependency(k, pid, Process.pid, $LOADED_FEATURES)
      pids[pid] = k
    end
    loop {
      pid = Process.wait
      if (status = $?.exitstatus) > 0
        exit status
      else # restart
        name = pids[pid]
        pids.delete(pid)
        new_pid = fork(&forks[name])
        register_dependency(name, new_pid, Process.pid, $LOADED_FEATURES)
        pids[new_pid] = name
      end
    }
  end

  def acceptor(socket, &b)
    ->{
      $0 = $0.sub(/spawner/, "acceptor")
      File.unlink(socket) rescue nil
      server = UNIXServer.new(socket)
      loop {
        client = server.accept
        ActiveRecord::Base.clear_all_connections!
        child = fork do
          ActiveRecord::Base.establish_connection
          ActiveSupport::DescendantsTracker.clear
          ActiveSupport::Dependencies.clear

          terminal = client.recv_io
          args = JSON.load(client.gets.strip).values_at("arguments", "environment")
          arguments, environment = args


          # if command == "test" and environment == "test"
          client << { status: "OK", pid: $$ }.to_json << "\n"
          # else
          #   client << { status: "NO", message: "invalid arguments or environment" }.to_json << "\n"
          #   exit
          # end

          $stdin.reopen(terminal)
          $stdout.reopen(terminal)
          $stderr.reopen(terminal)
          ARGV.replace(arguments)

          b.call
        end
        Process.detach(child)
        client.close
      }
    }
  end
end
end
