require 'spec_helper'

class Zeus::Server
  describe ProcessTreeMonitor do

    let(:file_monitor) { stub }
    let(:tree) { stub }
    let(:monitor) { ProcessTreeMonitor.new(file_monitor, tree) }

    it "closes sockets not useful to forked processes" do
      parent, child = stub, stub
      ProcessTreeMonitor.any_instance.stub(open_socketpair: [parent, child])
      parent.should_receive(:close)
      monitor.close_parent_socket
    end

    it "closes sockets not useful to the master process" do
      parent, child = stub, stub
      ProcessTreeMonitor.any_instance.stub(open_socketpair: [parent, child])
      child.should_receive(:close)
      monitor.close_child_socket
    end

    it "kills nodes with a feature that changed" do
      tree.should_receive(:kill_nodes_with_feature).with("rails")
      monitor.kill_nodes_with_feature("rails")
    end

    it "passes process inheritance information to the tree" do
      IO.select([monitor.datasource], [], [], 0).should be_nil
      monitor.__CHILD__stage_starting_with_pid(:name, 1)
      IO.select([monitor.datasource], [], [], 0.5).should_not be_nil
      tree.should_receive(:stage_has_pid).with(:name, 1)
      monitor.on_datasource_event
    end

    it "passes process feature information to the tree" do
      IO.select([monitor.datasource], [], [], 0).should be_nil
      monitor.__CHILD__stage_has_feature(:name, "rails")
      IO.select([monitor.datasource], [], [], 0.5).should_not be_nil
      tree.should_receive(:stage_has_feature).with(:name, "rails")
      file_monitor.should_receive(:watch).with("rails")
      monitor.on_datasource_event
    end

    private

  end
end

