require 'spec_helper'

describe "Integration" do
  after do
    kill_all_children
  end

  context "in a non-rails project with a .zeus.rb" do
    it "starts the zeus server and responds to commands" do
      write ".zeus.rb", <<-RUBY
        Zeus::Server.define! do
          stage :foo do
            command :bar do
              puts "YES"
            end
          end
        end
      RUBY

      start, run = start_and_run("bar")
      start.should include "spawner `foo`"
      start.should include "acceptor `bar`"
      run.should == ["YES\r\n"]
    end

    it "can run via command alias" do
      write ".zeus.rb", <<-RUBY
        Zeus::Server.define! do
          stage :foo do
            command :bar, :b do
              puts "YES"
            end
          end
        end
      RUBY

      start, run = start_and_run("b")
      start.should include "spawner `foo`"
      start.should include "acceptor `bar`"
      run.should == ["YES\r\n"]
    end
  end

  private

  def zeus(command, options={})
    command = zeus_command(command)
    result = `#{command}`
    raise "FAILED #{command}\n#{result}" if $?.success? == !!options[:fail]
    result
  end

  def zeus_command(command)
    "ruby -I #{root}/lib #{root}/bin/zeus #{command} 2>&1"
  end

  def record_start(output)
    IO.popen(zeus_command("start")) do |pipe|
      while str = pipe.readpartial(100)
        output << str
      end rescue EOFError
    end
  end

  def start_and_run(commands)
    start_output = ""
    t1 = Thread.new { record_start(start_output) }
    sleep 0.1
    run_output = [*commands].map{ |cmd| zeus(cmd) }
    sleep 0.2
    t1.kill
    [start_output, run_output]
  end

  def kill_all_children
    `kill -9 #{child_pids.join(" ")}`
  end

  def child_pids
    pid = Process.pid
    pipe = IO.popen("ps -ef | grep #{pid}")
    pipe.readlines.map do |line|
      parts = line.split(/\s+/)
      parts[2] if parts[3] == pid.to_s and parts[2] != pipe.pid.to_s
    end.compact
  end
end
