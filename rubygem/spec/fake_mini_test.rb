module Minitest
  class Runnable
  end
end

class MiniTest
  class Unit
    class TestCase
    end
  end
end

def stub_mini_test_methods
  allow(Minitest::Runnable).to receive(:runnables).and_return([fake_mt5_suite])
  allow(MiniTest::Unit::TestCase).to receive(:test_suite).and_return([fake_mt_old_suite])
end

def fake_runner
  @runner ||= double("Runner", :run => 0)
end

def fake_mt5_suite
  @suite ||= double("TestSuite",
                  :runnable_methods => [test_method],
                  :instance_method => fake_instance_method(test_method))
end

def fake_mt_old_suite
  @suite ||= double("TestSuite",
                  :test_methods => [test_method],
                  :instance_method => fake_instance_method(test_method))
end

def fake_test_method
  "test_method"
end

def fake_special_characters_test_method
  "test_my_test_method?"
end

def fake_instance_method(name=fake_test_method)
  @instance_method ||=  double("InstanceMethod",
                             :source_location => ["path/to/file.rb", 2],
                             :source => "def #{name} \n assert true \n end")
end

