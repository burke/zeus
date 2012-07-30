require 'json'
require 'socket'

require 'rb-kqueue'
require 'zeus/process'

module Zeus
  class Server

    autoload :Stage,                       'zeus/server/stage'
    autoload :Acceptor,                    'zeus/server/acceptor'
    autoload :FileMonitor,                 'zeus/server/file_monitor'
    autoload :ClientHandler,               'zeus/server/client_handler'
    autoload :ProcessTreeMonitor,          'zeus/server/process_tree_monitor'
    autoload :AcceptorRegistrationMonitor, 'zeus/server/acceptor_registration_monitor'

    def self.define!(&b)
      @@definition = Zeus::DSL::Evaluator.new.instance_eval(&b)
    end

    def self.acceptors
      @@definition.acceptors
    end

    attr_reader :client_handler, :acceptor_registration_monitor
    def initialize
      @file_monitor                  = FileMonitor.new(&method(:dependency_did_change))
      @acceptor_registration_monitor = AcceptorRegistrationMonitor.new
      @process_tree_monitor          = ProcessTreeMonitor.new
      @client_handler                = ClientHandler.new(acceptor_registration_monitor)

      # TODO: deprecate Zeus::Server.define! maybe. We can do that better...
      @plan = @@definition.to_domain_object(self)
    end

    def dependency_did_change(file)
      @process_tree_monitor.kill_nodes_with_feature(file)
    end

    PID_TYPE = "P"
    def w_pid line
      @w_msg.send(PID_TYPE + line, 0)
    rescue Errno::ENOBUFS
      sleep 0.2
      retry
    end

    FEATURE_TYPE = "F"
    def w_feature line
      @w_msg.send(FEATURE_TYPE + line, 0)
    rescue Errno::ENOBUFS
      sleep 0.2
      retry
    end

    def run
      $0 = "zeus master"
      trap("INT") { exit 0 }
      at_exit { Process.killall_descendants(9) }

      @r_msg, @w_msg = Socket.pair(:UNIX, :DGRAM)

      # boot the actual app
      @plan.run
      @w_msg.close

      loop do
        @file_monitor.process_events

        # TODO: Make @r_msg a Monitor instead. All that logic should be its own thing.
        monitors = [@acceptor_registration_monitor, @client_handler]
        datasources = [@r_msg, *monitors.map(&:datasource)]

        # TODO: It would be really nice if we could put the queue poller in the select somehow.
        #   --investigate kqueue. Is this possible?
        ready, _, _ = IO.select(datasources, [], [], 1)
        next unless ready
        monitors.each do |m|
          m.on_datasource_event if ready.include?(m.datasource)
        end
        handle_messages if ready.include?(@r_msg)
      end

    ensure
      File.unlink(Zeus::SOCKET_NAME)
    end

    def handle_messages
      loop do
        handle_message
      end
    rescue Errno::EAGAIN
    end

    def handle_message
      data = @r_msg.recv_nonblock(1024)
      case data[0]
      when FEATURE_TYPE
        handle_feature_message(data[1..-1])
      when PID_TYPE
        handle_pid_message(data[1..-1])
      else
        raise "Unrecognized message"
      end
    end

    def handle_pid_message(data)
      data =~ /(\d+):(\d+)/
        pid, ppid = $1.to_i, $2.to_i
      @process_tree_monitor.process_has_parent(pid, ppid)
    end

    def handle_feature_message(data)
      data =~ /(\d+):(.*)/
        pid, file = $1.to_i, $2
      @process_tree_monitor.process_has_feature(pid, file)
      @file_monitor.watch(file)
    end

    def self.pid_has_file(pid, file)
      @@files[file] ||= []
      @@files[file] << pid
    end

  end
end
