require "io/console"
require "json"
require "pty"
require "socket"

module Zeus
  class Client

    SIGNALS = {
      "\x03" => "TERM",
      "\x1C" => "QUIT"
    }

    def self.maybe_raw(&b)
      if $stdout.tty?
        $stdout.raw(&b)
      else
        b.call
      end
    end

    def self.start!

      ENV["RAILS_ENV"] ||= "test"

      maybe_raw do
        PTY.open do |master, slave|
          $stdout.tty? and master.winsize = $stdout.winsize
          winch, winch_ = IO.pipe
          trap("WINCH") { winch_ << "\0" }

          case ARGV.shift
          when 'testrb'
            socket = UNIXSocket.new(".zeus.test_testrb.sock")
          when 'console'
            socket = UNIXSocket.new(".zeus.dev_console.sock")
          when 'server'
            socket = UNIXSocket.new(".zeus.dev_server.sock")
          when 'rake'
            socket = UNIXSocket.new(".zeus.dev_rake.sock")
          when 'runner'
            socket = UNIXSocket.new(".zeus.dev_runner.sock")
          when 'generate'
            socket = UNIXSocket.new(".zeus.dev_generate.sock")
          end
          socket.send_io(slave)
          socket << { arguments: ARGV, environment: ENV["RAILS_ENV"], tty: slave.path }.to_json << "\n"
          slave.close

          response = JSON.load(socket.gets.strip)
          raise "server said no" unless response["status"] == "OK"
          pid = response["pid"]

          begin
            buffer = ""
            signals = Regexp.union(SIGNALS.keys)

            while ready = select([winch, master, $stdin])[0]
              if ready.include?(winch)
                winch.read(1)
                $stdout.tty? and master.winsize = $stdout.winsize
                Process.kill("WINCH", pid)
              end

              if ready.include?($stdin)
                input = $stdin.readpartial(4096, buffer)
                input.scan(signals).each { |signal| Process.kill(SIGNALS[signal], pid) }
                master << input
              end

              if ready.include?(master)
                $stdout << master.readpartial(4096, buffer)
              end
            end
          rescue EOFError
          end
        end
      end
    end
  end
end
