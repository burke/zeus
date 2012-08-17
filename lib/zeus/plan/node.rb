module Zeus
  module Plan

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

  end
end
