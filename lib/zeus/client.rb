require "io/console"
require "json"
require "pty"
require "socket"

module Zeus
  class Client
    ComandNotFound = Class.new(Exception)

    SIGNALS = {
      "\x03" => "TERM",
      "\x1C" => "QUIT"
    }
    SIGNAL_REGEX = Regexp.union(SIGNALS.keys)

    SOCKETS = {
      'testrb'    => '.zeus.test_testrb.sock',
      'console'   => '.zeus.dev_console.sock',
      'server'    => '.zeus.dev_server.sock',
      'rake'      => '.zeus.dev_rake.sock',
      'runner'    => '.zeus.dev_runner.sock',
      'generate'  => '.zeus.dev_generate.sock'
    }

    def self.maybe_raw(&b)
      if $stdout.tty?
        $stdout.raw(&b)
      else
        b.call
      end
    end

    def self.cleanup!
      SOCKETS.values.each do |socket|
        FileUtils.rm_rf socket
      end
    end

    def self.run

      maybe_raw do
        PTY.open do |master, slave|
          $stdout.tty? and master.winsize = $stdout.winsize
          winch, winch_ = IO.pipe
          trap("WINCH") { winch_ << "\0" }

          command = ARGV.shift

          socket = case command
          when 'testrb', 't'
            UNIXSocket.new(".zeus.test_testrb.sock")
          when 'console', 'c'
            UNIXSocket.new(".zeus.dev_console.sock")
          when 'server', 's'
            UNIXSocket.new(".zeus.dev_server.sock")
          when 'rake'
            UNIXSocket.new(".zeus.dev_rake.sock")
          when 'runner', 'r'
            UNIXSocket.new(".zeus.dev_runner.sock")
          when 'generate', 'g'
            UNIXSocket.new(".zeus.dev_generate.sock")
          else
            raise ComandNotFound.new(command)
          end

          socket.send_io(slave)
          socket << ARGV.to_json << "\n"
          slave.close

          pid = socket.gets.strip.to_i

          begin
            buffer = ""

            while ready = select([winch, master, $stdin])[0]
              if ready.include?(winch)
                winch.read(1)
                $stdout.tty? and master.winsize = $stdout.winsize
                Process.kill("WINCH", pid)
              end

              if ready.include?($stdin)
                input = $stdin.readpartial(4096, buffer)
                input.scan(SIGNAL_REGEX).each { |signal|
                  Process.kill(SIGNALS[signal], pid)
                }
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
