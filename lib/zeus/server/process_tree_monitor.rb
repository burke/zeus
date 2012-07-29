module Zeus
  class Server
    class ProcessTreeMonitor

      def initialize
        @tree = ProcessTree.new
      end

      def kill_nodes_with_feature(file)
        @tree.kill_nodes_with_feature(file)
      end

      def process_has_feature(pid, file)
        @tree.process_has_feature(pid, file)
      end

      def process_has_parent(pid, ppid)
        @tree.process_has_parent(pid, ppid)
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
              Zeus.ui.error "One of zeus's dependencies changed. Not killing zeus. You may have to restart the server."
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

    end
  end
end
