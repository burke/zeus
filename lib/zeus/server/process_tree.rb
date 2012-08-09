module Zeus
  class Server
    class ProcessTree

      def initialize
        @root = Node.new(Process.pid)
        @nodes_by_pid = {Process.pid => @root}
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

      def kill_nodes_with_feature(file, base = @root)
        if base.has_feature?(file)
          kill_node(base)
        else
          base.children.dup.each do |node|
            if kill_nodes_with_feature(file, node)
              base.children.delete(node)
            end
          end
          return false
        end
      end

      private

      def node_for_pid(pid)
        @nodes_by_pid[pid.to_i] ||= Node.new(pid.to_i)
      end

      def kill_node(node)
        if node == @root.children[0] || node == @root
          Zeus.ui.error "One of zeus's dependencies changed. Not killing zeus. You may have to restart the server."
          return false
        end
        @nodes_by_pid.delete(node.pid)
        node.kill
      end

      class Node
        attr_accessor :pid, :children, :features
        def initialize(pid)
          @pid, @children, @features = pid, [], {}
        end

        def kill
          # recall that this process explicitly traps TERM -> exit 0
          Process.kill("TERM", pid)
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

      end

    end
  end
end
