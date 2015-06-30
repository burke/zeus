require 'spec_helper'

module Zeus
  describe Rails do
    subject(:rails) { Rails.new }

    describe "#test_helper" do
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

    describe "#test" do
      def expect_minitest_autorun
        # Zeus::Rails#test_helper will require minitest/unit by default.
        # We need to catch it first, before setting expectations on
        # helper_file requires below.
        expect_any_instance_of(Rails).to receive(:require).with("minitest/autorun")
      end

      context 'minitest' do
        before do
          module Zeus::M
          end
          expect(Zeus::M).to receive(:run)
        end

        it "requires autorun when testing with new minitest" do
          module ::Minitest
          end
          expect_minitest_autorun

          rails.test
        end

        it "requires autorun when testing with old minitest" do
          expect_minitest_autorun

          rails.test
        end
      end

      context 'rspec' do
        before do
          class ::RSpec::Core::Runner
          end
        end

        it "calls rspec core runner" do
          expect(RSpec::Core::Runner).to receive(:invoke)
          rails.test(['test_spec.rb'])
        end
      end
    end
  end
end
