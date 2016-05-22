require 'zeus/rails'
require 'fake_mini_test'

module Zeus::M
  describe Runner do
    let(:test_method) { fake_test_method }

    matcher :exit_with_code do |exp_code|
      actual = nil
      match do |block|
        begin
          block.call
        rescue SystemExit => e
          actual = e.status
        end
        actual and actual == exp_code
      end
      failure_message do |_block|
        "expected block to call exit(#{exp_code}) but exit" +
          (actual.nil? ? " not called" : "(#{actual}) was called")
      end
      failure_message_when_negated do |_block|
        "expected block not to call exit(#{exp_code})"
      end
      description do
        "expect block to call exit(#{exp_code})"
      end
    end

    before(:each) do
      allow(Dir).to receive(:glob).and_return(["path/to/file.rb"])
      allow(Kernel).to receive(:load)

      stub_mini_test_methods
    end

    context "given a test with a question mark" do
      let(:test_method) { fake_special_characters_test_method }

      it "escapes the question mark when using line number" do
        argv = ["path/to/file.rb:2"]

        expect(lambda { Runner.new(argv).run }).to exit_with_code(0)
        expect(ARGV).to eq(["-n", "/(test_my_test_method\\?)/"])
      end

      it "does not escape regex on explicit names" do
        argv = ["path/to/file.rb", "--name", fake_special_characters_test_method]

        expect(lambda { Runner.new(argv).run }).to exit_with_code(0)

        expect(ARGV).to eq(["-n", "test_my_test_method?"])
      end
    end

    context "given no option" do
      it "runs the test" do
        argv = ["path/to/file.rb"]

        expect(lambda { Runner.new(argv).run }).to exit_with_code(0)

        expect(ARGV).to eq([])
      end
    end

    context "given a line number" do
      it "aborts if no test is found on that line number" do
        argv = ["path/to/file.rb:100"]

        expect(STDERR).to receive(:write).with(/No tests found on line 100/)

        expect(lambda { Runner.new(argv).run }).to_not exit_with_code(0)
      end

      it "runs the test if the correct line number is given" do
        argv = ["path/to/file.rb:2"]

        expect(lambda { Runner.new(argv).run }).to exit_with_code(0)
        expect(ARGV).to eq(["-n", "/(#{fake_test_method})/"])
      end
    end

    context "given a specific test name" do
      it "runs the specified tests when using a pattern in --name option" do
        argv = ["path/to/file.rb", "--name", "/#{fake_test_method}/"]

        expect(lambda { Runner.new(argv).run }).to exit_with_code(0)
        expect(ARGV).to eq(["-n", "/#{fake_test_method}/"])
      end

      it "runs the specified tests when using a pattern in -n option" do
        argv = ["path/to/file.rb", "-n", "/method/"]

        expect(lambda { Runner.new(argv).run }).to exit_with_code(0)
        expect(ARGV).to eq(["-n", "/method/"])
      end

      it "aborts if no test matches the given pattern" do
        argv = ["path/to/file.rb", "-n", "/garbage/"]

        expect(STDERR).to receive(:write).with(%r{No test name matches \'/garbage/\'})
        expect(fake_runner).to_not receive :run

        expect(lambda { Runner.new(argv).run }).to_not exit_with_code(0)
      end

      it "runs the specified tests when using a name (no pattern)" do
        argv = ["path/to/file.rb", "-n", "#{fake_test_method}"]

        expect(lambda { Runner.new(argv).run }).to exit_with_code(0)
        expect(ARGV).to eq(["-n", fake_test_method])
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
