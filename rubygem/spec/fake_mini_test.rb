module MiniTest
  module Unit
    class TestCase
    end
  end
end

def stub_mini_test_methods
  allow(MiniTest::Unit::TestCase).to receive(:test_suites).and_return([fake_suite])
  allow(MiniTest::Unit).to receive(:runner).and_return(fake_runner)
end

def fake_runner
 @runner ||= double("Runner", :run => 0)
end

def fake_suite
  @suite ||= double("TestSuite",
                  :test_methods => [fake_test_method],
                  :instance_method => fake_instance_method)
end

def fake_suite_with_special_characters
  @suite ||= double("TestSuite",
                  :test_methods => [fake_special_characters_test_method],
                  :instance_method => fake_instance_method(fake_special_characters_test_method))
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

