require 'zeus/rails'

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
  failure_message_for_should do |block|
    "expected block to call exit(#{exp_code}) but exit" +
      (actual.nil? ? " not called" : "(#{actual}) was called")
  end
  failure_message_for_should_not do |block|
    "expected block not to call exit(#{exp_code})"
  end
  description do
    "expect block to call exit(#{exp_code})"
  end
end

def stub_system_methods
  Dir.stub!(:glob).and_return(["path/to/file.rb"])
  Kernel.stub!(:load).and_return
end

RSpec.configure do |config|
  config.before(:each) do
    stub_system_methods
  end
end
