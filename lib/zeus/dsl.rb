require 'set'

module Zeus
  module DSL

    class Evaluator
      def stage(name, &b)
        stage = DSL::Stage.new(name)
        stage.root = true
        stage.instance_eval(&b)
        stage
      end
    end

    class Node
      attr_reader :name, :stages, :features
      attr_accessor :pid, :root

      def initialize(name)
        @name     = name
        @stages   = []
        @features = Set.new # hash might be faster than ruby's inane impl of set.
      end

      def stage_has_feature(name, file)
        node_for_name(name).features << file
      end

      def stage_has_pid(name, pid)
        node_for_name(name).pid = pid
      end

      def kill_nodes_with_feature(file)
        if features.include?(file)
          if root
            Zeus.ui.error "One of zeus's dependencies changed. Not killing zeus. You may have to restart the server."
          else
            kill!
          end
        else
          stages.each do |child|
            child.kill_nodes_with_feature(file)
          end
        end
      end

      # We send STOP before actually killing the processes here.
      # This is to prevent parents from respawning before all the children
      # are killed. This prevents a race condition.
      def kill!
        Process.kill("STOP", pid) if pid
        # Protected methods don't work with each(&:m) notation.
        stages.each { |stage| stage.kill! }
        old_pid  = pid
        self.pid = nil
        Process.kill("KILL", old_pid) if old_pid
      end

      private

      def node_for_name(name)
        @nodes_by_name ||= __nodes_by_name
        @nodes_by_name[name]
      end

      def __nodes_by_name
        nodes = {name => self}
        stages.each do |child|
          nodes.merge!(child.__nodes_by_name)
        end
        nodes
      end ; protected :__nodes_by_name

    end

    class Acceptor < Node

      attr_reader :name, :aliases, :description, :action
      def initialize(name, aliases, description, &b)
        super(name)
        @description = description
        @aliases = aliases
        @action = b
      end

      # ^ configuration
      # V later use

      def commands
        [name, *aliases].map(&:to_s)
      end

      def acceptors
        self
      end

      def to_process_object(server)
        Zeus::Server::Acceptor.new(server).tap do |stage|
          stage.name = @name
          stage.aliases = @aliases
          stage.action = @action
          stage.description = @description
        end
      end

    end

    class Stage < Node

      attr_reader :actions
      def initialize(name)
        super(name)
        @actions = []
      end

      def action(&b)
        @actions << b
        self
      end

      def desc(desc)
        @desc = desc
      end

      def stage(name, &b)
        @stages << DSL::Stage.new(name).tap { |s| s.instance_eval(&b) }
        self
      end

      def command(name, *aliases, &b)
        @stages << DSL::Acceptor.new(name, aliases, @desc, &b)
        @desc = nil
        self
      end

      # ^ configuration
      # V later use

      def acceptors
        stages.map(&:acceptors).flatten
      end

      def to_process_object(server)
        Zeus::Server::Stage.new(server).tap do |stage|
          stage.name = @name
          stage.stages = @stages.map { |stage| stage.to_process_object(server) }
          stage.actions = @actions
        end
      end

    end

  end
end
