require 'spec_helper'
require 'fake_mini_test'

module Zeus::M
  describe Runner do

    context "given a test with a question mark" do
      before(:each) do
        allow(MiniTest::Unit::TestCase).to receive(:test_suites).and_return([fake_suite_with_special_characters])
        allow(MiniTest::Unit).to receive(:runner).and_return(fake_runner)
      end

      it "escapes the question mark when using line number" do
        argv = ["path/to/file.rb:2"]

        expect(fake_runner).to receive(:run).with(["-n", "/(test_my_test_method\\?)/"])

        expect(lambda { Runner.new(argv).run }).to exit_with_code(0)
      end

      it "does not escape regex on explicit names" do
        argv = ["path/to/file.rb", "--name", fake_special_characters_test_method]

        allow(fake_runner).to receive(:run).with(["-n", "test_my_test_method?"])

        expect(lambda { Runner.new(argv).run }).to exit_with_code(0)
      end
    end
  end

  describe Runner do
    before(:each) do
      stub_mini_test_methods
    end

    context "no option is given" do
      it "runs the test without giving any option" do
        argv = ["path/to/file.rb"]

        allow(fake_runner).to receive(:run).with([])

        expect(lambda { Runner.new(argv).run }).to exit_with_code(0)
      end
    end

    context "given a line number" do
      it "aborts if no test is found" do
        argv = ["path/to/file.rb:100"]

        expect(STDERR).to receive(:write).with(/No tests found on line 100/)
        expect(fake_runner).to_not receive :run

        expect(lambda { Runner.new(argv).run }).to_not exit_with_code(0)
      end

      it "runs the test if the correct line number is given" do
        argv = ["path/to/file.rb:2"]

        expect(fake_runner).to receive(:run).with(["-n", "/(#{fake_test_method})/"])

        expect(lambda { Runner.new(argv).run }).to exit_with_code(0)
      end
    end

    context "specifying test name" do
      it "runs the specified tests when using a pattern in --name option" do
        argv = ["path/to/file.rb", "--name", "/#{fake_test_method}/"]

        expect(fake_runner).to receive(:run).with(["-n", "/#{fake_test_method}/"])

        expect(lambda { Runner.new(argv).run }).to exit_with_code(0)
      end

      it "runs the specified tests when using a pattern in -n option" do
        argv = ["path/to/file.rb", "-n", "/method/"]

        expect(fake_runner).to receive(:run).with(["-n", "/method/"])

        expect(lambda { Runner.new(argv).run }).to exit_with_code(0)
      end

      it "aborts if no test matches the given pattern" do
        argv = ["path/to/file.rb", "-n", "/garbage/"]

        expect(STDERR).to receive(:write).with(%r{No test name matches \'/garbage/\'})
        expect(fake_runner).to_not receive :run

        expect(lambda { Runner.new(argv).run }).to_not exit_with_code(0)
      end

      it "runs the specified tests when using a name (no pattern)" do
        argv = ["path/to/file.rb", "-n", "#{fake_test_method}"]

        expect(fake_runner).to receive(:run).with(["-n", fake_test_method])

        expect(lambda { Runner.new(argv).run }).to exit_with_code(0)
      end

      it "aborts if no test matches the given test name" do
        argv = ["path/to/file.rb", "-n", "method"]

        expect(STDERR).to receive(:write).with(%r{No test name matches \'method\'})
        expect(fake_runner).to_not receive :run

        expect(lambda { Runner.new(argv).run }).to_not exit_with_code(0)
      end
    end
  end

end
