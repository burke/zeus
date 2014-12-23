require 'spec_helper'

module Zeus
  describe Rails do
    subject(:rails) { Rails.new }

    context "#test_helper" do
      before(:each) do
        expect(rails).to receive(:require).with("minitest/unit")
      end

      it "when ENV['RAILS_TEST_HELPER'] is set helper is loaded from variable" do
        ENV['RAILS_TEST_HELPER'] = "a_test_helper"
        expect(rails).to receive(:require).with("a_test_helper")

        rails.test_helper
        ENV.clear
      end

      it "requires rails_helper when using rspec 3.0+" do
        mock_file_existence(ROOT_PATH + "/spec/rails_helper.rb", true)

        expect(rails).to receive(:require).with("rails_helper")

        rails.test_helper
      end

      it "when spec_helper exists spec_helper is required" do
        mock_file_existence(ROOT_PATH + "/spec/rails_helper.rb", false)
        mock_file_existence(ROOT_PATH + "/spec/spec_helper.rb", true)

        expect(rails).to receive(:require).with("spec_helper")

        rails.test_helper
      end

      it "when minitest_helper exists minitest_helper is required" do
        mock_file_existence(ROOT_PATH + "/spec/rails_helper.rb", false)
        mock_file_existence(ROOT_PATH + "/spec/spec_helper.rb", false)
        mock_file_existence(ROOT_PATH + "/test/minitest_helper.rb", true)

        expect(rails).to receive(:require).with("minitest_helper")

        rails.test_helper
      end

      it "when there is no rspec helpers or minitest_helper, test_helper is required" do
        mock_file_existence(ROOT_PATH + "/spec/rails_helper.rb", false)
        mock_file_existence(ROOT_PATH + "/spec/spec_helper.rb", false)
        mock_file_existence(ROOT_PATH + "/test/minitest_helper.rb", false)

        expect(rails).to receive(:require).with("test_helper")

        rails.test_helper
      end
    end

    context "#gem_is_bundled?" do
      it "returns gem version from Gemfile.lock" do
        allow(File).to receive(:read).and_return("
GEM
  remote: https://rubygems.org/
  specs:
    exception_notification-rake (0.0.6)
      exception_notification (~> 3.0.1)
      rake (>= 0.9.0)
    rake (10.0.4)
")
        expect(gem_is_bundled?('rake')).to eq '10.0.4'
      end
    end
  end
end
