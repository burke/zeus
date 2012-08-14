require 'spec_helper'

describe Zeus::Server::LoadTracking do
  class Recorder
    def recorded
      @recorded ||= []
    end

    def add_extra_feature(*args)
      recorded << [:add_extra_feature, *args]
    end
  end

  let(:recorder){ Recorder.new }

  around do |example|
    $foo = nil
    Zeus::Server::LoadTracking.server = recorder
    example.call
    Zeus::Server::LoadTracking.server = nil
  end

  let(:tmp_path) { File.expand_path(Dir.pwd) }

  it "tracks loading of absolute paths" do
    write "foo.rb", "$foo = 1"
    load "#{Dir.pwd}/foo.rb"
    $foo.should == 1
    recorder.recorded.should == [[:add_extra_feature, tmp_path + "/foo.rb"]]
  end

  it "tracks loading of relative paths" do
    write "foo.rb", "$foo = 1"
    load "./foo.rb"
    $foo.should == 1
    recorder.recorded.should == [[:add_extra_feature, tmp_path + "/foo.rb"]]
  end

  it "tracks loading from library paths" do
    write "lib/foo.rb", "$foo = 1"
    restoring $LOAD_PATH do
      $LOAD_PATH << File.expand_path("lib")
      load "foo.rb"
    end
    $foo.should == 1
    recorder.recorded.should == [[:add_extra_feature, tmp_path + "/lib/foo.rb"]]
  end

  it "does not add unfound files" do
    write "lib/foo.rb", "$foo = 1"
    begin
      load "foo.rb"
    rescue LoadError
    end
    $foo.should == nil
    recorder.recorded.should == []
  end

  private

  def restoring(thingy)
    old = thingy.dup
    yield
  ensure
    thingy.replace(old)
  end
end
