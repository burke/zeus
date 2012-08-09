require 'zeus'

class Zeus::Server

  describe ProcessTree do

    ROOT_PID = Process.pid
    CHILD_1 = ROOT_PID + 1
    CHILD_2 = ROOT_PID + 2
    GRANDCHILD_1 = ROOT_PID + 3
    GRANDCHILD_2 = ROOT_PID + 4

    let(:process_tree) { ProcessTree.new }

    before do
      build_tree
      add_features
    end

    it "doesn't kill the root node" do
      Zeus.ui.should_receive(:error).with(/not killing zeus/i)
      Process.should_not_receive(:kill)
      process_tree.kill_nodes_with_feature("zeus")
    end

    it "kills a node that has a feature" do
      expect_kill(CHILD_2)
      process_tree.kill_nodes_with_feature("rails")
    end

    it "kills multiple nodes at the same level with a feature" do
      expect_kill(GRANDCHILD_1)
      expect_kill(GRANDCHILD_2)
      process_tree.kill_nodes_with_feature("model")
    end

    private

    def expect_kill(pid)
      Process.should_receive(:kill).with("USR1", pid)
    end

    def build_tree
      process_tree.process_has_parent(CHILD_1, ROOT_PID)
      process_tree.process_has_parent(CHILD_2, CHILD_1)
      process_tree.process_has_parent(GRANDCHILD_1, CHILD_2)
      process_tree.process_has_parent(GRANDCHILD_2, CHILD_2)
    end

    def add_features
      [CHILD_2, GRANDCHILD_1, GRANDCHILD_2].each do |pid|
        process_tree.process_has_feature(pid, "rails")
      end

      process_tree.process_has_feature(GRANDCHILD_1, "model")
      process_tree.process_has_feature(GRANDCHILD_2, "model")

      [ROOT_PID, CHILD_1, CHILD_2, GRANDCHILD_1, GRANDCHILD_2].each do |pid|
        process_tree.process_has_feature(pid, "zeus")
      end
    end

  end

end
