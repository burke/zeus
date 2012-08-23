require 'spec_helper'

describe "Integration" do
  after do
    kill_all_children
  end

  context "in a non-rails project with a .zeus.rb" do
    it "starts the zeus server and responds to commands" do
      bar_setup("puts 'YES'")
      start, run = start_and_run("bar")
      start.should include "spawner `foo`"
      start.should include "acceptor `bar`"
      run.should == ["YES\r\n"]
    end

    it "receives ARGV after command" do
      bar_setup("puts ARGV.join(', ')")
      start, run = start_and_run("bar 1 '2 3 4' --123")
      run.should == ["1, 2 3 4, --123\r\n"]
    end

    it "cam exist with 0" do
      bar_setup("exit 0")
      start_and_run("bar")
    end

    it "can exist with non-0" do
      bar_setup("exit 1")
      start_and_run("bar", :fail => true)
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

  def bar_setup(inner)
    write ".zeus.rb", <<-RUBY
      Zeus::Server.define! do
        stage :foo do
          command :bar do
            #{inner}
          end
        end
      end
    RUBY
  end

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

  def start_and_run(commands, options={})
    start_output = ""
    t1 = Thread.new { record_start(start_output) }
    sleep 0.2
    run_output = [*commands].map{ |cmd| zeus(cmd, options) }
    sleep 0.2
    t1.kill
    [start_output, run_output]
  end
end
