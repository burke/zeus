require 'spec_helper'

module Zeus
  describe Rails do
    subject(:rails) { Rails.new }

    context "#test_helper" do
      before(:each) do
        rails.should_receive(:require).with("minitest/unit")
      end

      it "when ENV['RAILS_TEST_HELPER'] is set helper is loaded from variable" do
        ENV['RAILS_TEST_HELPER'] = "a_test_helper"
        rails.should_receive(:require).with("a_test_helper")

        rails.test_helper
        ENV.clear
      end

      it "when spec_helper exists spec_helper is required" do
        File.should_receive(:exists?).with(ROOT_PATH + "/spec/spec_helper.rb").and_return(true)

        rails.should_receive(:require).with("spec_helper")

        rails.test_helper
      end

      it "when minitest_helper exists minitest_helper is required" do
        File.should_receive(:exists?).with(ROOT_PATH + "/spec/spec_helper.rb").and_return(false)
        File.should_receive(:exists?).with(ROOT_PATH + "/test/minitest_helper.rb").and_return(true)

        rails.should_receive(:require).with("minitest_helper")

        rails.test_helper
      end

      it "when there is no spec_helper or minitest_helper, test_helper is required" do
        File.should_receive(:exists?).with(ROOT_PATH + "/spec/spec_helper.rb").and_return(false)
        File.should_receive(:exists?).with(ROOT_PATH + "/test/minitest_helper.rb").and_return(false)

        rails.should_receive(:require).with("test_helper")

        rails.test_helper
      end
    end
  end
end
