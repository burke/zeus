module MiniTest
  module Unit
    class TestCase
    end
  end
end

def stub_mini_test_methods
  MiniTest::Unit::TestCase.stub!(:test_suites).and_return [fake_suite]
  MiniTest::Unit.stub!(:runner).and_return fake_runner
end

def fake_runner
 @runner ||= stub("Runner", :run => 0)
end

def fake_suite
  @suite ||= stub("TestSuite",
                  :test_methods => [fake_test_method],
                  :instance_method => fake_instance_method)
end

def fake_test_method
  "test_method"
end

def fake_instance_method
  @instance_method ||=  stub("InstanceMethod",
                             :source_location => ["path/to/file.rb", 2],
                             :source => "def #{fake_test_method} \n assert true \n end")
end
