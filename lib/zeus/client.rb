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
    SIGNAL_REGEX = Regexp.union(SIGNALS.keys)

    def self.maybe_raw(&b)
      if $stdout.tty?
        $stdout.raw(&b)
      else
        b.call
      end
    end

    def self.run(command, args)
      maybe_raw do
        PTY.open do |master, slave|
          $stdout.tty? and master.winsize = $stdout.winsize
          winch, winch_ = IO.pipe
          trap("WINCH") { winch_ << "\0" }

          socket = UNIXSocket.new(".zeus.sock")
          socket << {command: command, arguments: args}.to_json << "\n"
          socket.send_io(slave)
          slave.close

          pid = socket.readline.chomp.to_i

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

__FILE__ == $0 and Zeus::Client.run
