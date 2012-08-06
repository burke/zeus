require 'json'
require 'socket'

module Zeus
  class Server

    autoload :Stage,                       'zeus/server/stage'
    autoload :Acceptor,                    'zeus/server/acceptor'
    autoload :FileMonitor,                 'zeus/server/file_monitor'
    autoload :ProcessTree,                 'zeus/server/process_tree'
    autoload :ForkedProcess,               'zeus/server/forked_process'
    autoload :ClientHandler,               'zeus/server/client_handler'
    autoload :AcceptorErrorState,          'zeus/server/acceptor_error_state'
    autoload :ProcessTreeMonitor,          'zeus/server/process_tree_monitor'
    autoload :AcceptorRegistrationMonitor, 'zeus/server/acceptor_registration_monitor'

    def self.define!(&b)
      @@definition = Zeus::DSL::Evaluator.new.instance_eval(&b)
    end

    def self.acceptors
      defined?(@@definition) ? @@definition.acceptors : []
    end

    def initialize
      @file_monitor                  = FileMonitor::FSEvent.new(&method(:dependency_did_change))
      @acceptor_registration_monitor = AcceptorRegistrationMonitor.new
      @process_tree_monitor          = ProcessTreeMonitor.new(@file_monitor)
      acceptor_commands = self.class.acceptors.map(&:commands).flatten
      @client_handler                = ClientHandler.new(acceptor_commands, self)

      @plan = @@definition.to_domain_object(self)
    end

    def dependency_did_change(file)
      @process_tree_monitor.kill_nodes_with_feature(file)
    end

    def monitors
      [@file_monitor, @process_tree_monitor, @acceptor_registration_monitor, @client_handler]
    end

    def run
      $0 = "zeus master"
      trap("INT") { exit 0 }

      @plan.run(true) # boot the actual app
      monitors.each(&:close_child_socket)

      loop do
        ready, = IO.select(monitors.map(&:datasource), [], [], 1)
        next unless ready
        monitors.each do |m|
          m.on_datasource_event if ready.include?(m.datasource)
        end
      end
    ensure
      File.unlink(Zeus::SOCKET_NAME)
    end

    module ChildProcessApi

      def __CHILD__close_parent_sockets
        monitors.each(&:close_parent_socket)
      end

      def __CHILD__pid_has_ppid(pid, ppid)
        @process_tree_monitor.__CHILD__send_pid("#{pid}:#{Process.ppid}")
      end

      def __CHILD__pid_has_feature(pid, feature)
        @process_tree_monitor.__CHILD__send_feature("#{pid}:#{feature}")
      end

      def __CHILD__register_acceptor(io)
        @acceptor_registration_monitor.__CHILD__register_acceptor(io)
      end

      def __CHILD__find_acceptor_for_command(command)
        @acceptor_registration_monitor.__CHILD__find_acceptor_for_command(command)
      end

    end ; include ChildProcessApi

  end
end
