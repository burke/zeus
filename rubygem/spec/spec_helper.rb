require 'zeus/rails'
require 'pry'

RSpec::Matchers.define :exit_with_code do |exp_code|
  actual = nil
  match do |block|
    begin
      block.call
    rescue SystemExit => e
      actual = e.status
    end
    actual and actual == exp_code
  end
  failure_message do |block|
    "expected block to call exit(#{exp_code}) but exit" +
      (actual.nil? ? " not called" : "(#{actual}) was called")
  end
  failure_message_when_negated do |block|
    "expected block not to call exit(#{exp_code})"
  end
  description do
    "expect block to call exit(#{exp_code})"
  end
end

def stub_system_methods
  allow(Dir).to receive(:glob).and_return(["path/to/file.rb"])
  allow(Kernel).to receive(:load)
end

def mock_file_existence(file, result)
  expect(File).to receive(:exists?).with(file).and_return(result)
end

RSpec.configure do |config|
  config.before(:each) do
    stub_system_methods
  end
end
