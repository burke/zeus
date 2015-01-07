require 'spec_helper'

module Zeus
  describe Rails do
    subject(:rails) { Rails.new }

    describe "#test_helper" do
      before(:each) do
        # Zeus::Rails#test_helper will require minitest/unit by default.
        # We need to catch it first, before setting expectations on
        # helper_file requires below.
        expect(rails).to receive(:require).with("minitest/unit")
      end

      context "when ENV['RAILS_TEST_HELPER'] is set" do
        it "loads the test helper file from the environment variable" do
          helper = "a_test_helper"
          allow(ENV).to receive(:[]).with("RAILS_TEST_HELPER").and_return(helper)

          expect(rails).to receive(:require).with(helper)
          rails.test_helper
        end
      end

      context "when using rspec" do
        context "3.0+" do
          it "requires rails_helper" do
            mock_file_existence(ROOT_PATH + "/spec/rails_helper.rb", true)

            expect(rails).to receive(:require).with("rails_helper")

            rails.test_helper
          end
        end

        it "requires spec_helper" do
          mock_file_existence(ROOT_PATH + "/spec/rails_helper.rb", false)
          mock_file_existence(ROOT_PATH + "/spec/spec_helper.rb", true)

          expect(rails).to receive(:require).with("spec_helper")

          rails.test_helper
        end
      end

      context "when using minitest" do
        it "requires minitest_helper" do
          mock_file_existence(ROOT_PATH + "/spec/rails_helper.rb", false)
          mock_file_existence(ROOT_PATH + "/spec/spec_helper.rb", false)
          mock_file_existence(ROOT_PATH + "/test/minitest_helper.rb", true)

          expect(rails).to receive(:require).with("minitest_helper")

          rails.test_helper
        end
      end

      context "when there are no rspec or minitest helpers" do
        it "requires test_helper" do
          mock_file_existence(ROOT_PATH + "/spec/rails_helper.rb", false)
          mock_file_existence(ROOT_PATH + "/spec/spec_helper.rb", false)
          mock_file_existence(ROOT_PATH + "/test/minitest_helper.rb", false)

          expect(rails).to receive(:require).with("test_helper")

          rails.test_helper
        end
      end
    end

    describe "#gem_is_bundled?" do
      context "for a bundled gem" do
        it "returns the bundled gem's version" do
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
end
