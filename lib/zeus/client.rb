require "io/console"
require "json"
require "pty"
require "socket"

module Zeus
  class Client

    attr_accessor :pid

    SIGNALS = {
      "\x03" => "TERM",
      "\x1C" => "QUIT"
    }
    SIGNAL_REGEX = Regexp.union(SIGNALS.keys)

    def self.run(command, args)
      new.run(command, args)
    end

    def run(command, args)
      maybe_raw do
        PTY.open do |master, slave|
          @master = master
          set_winsize

          @winch = make_winch_channel
          @pid = connect_to_server(command, args, slave)

          buffer = ""
          begin
            while ready = select([@winch, @master, $stdin])[0]
              handle_winch          if ready.include?(@winch)
              handle_stdin(buffer)  if ready.include?($stdin)
              handle_master(buffer) if ready.include?(@master)
            end
          rescue EOFError
          end
        end
      end
    end

    private

    def connect_to_server(command, arguments, slave, socket_path = Zeus::SOCKET_NAME)
      socket = UNIXSocket.new(socket_path)
      socket << {command: command, arguments: arguments}.to_json << "\n"
      socket.send_io(slave)
      slave.close

      pid = socket.readline.chomp.to_i
    rescue Errno::ENOENT, Errno::ECONNREFUSED, Errno::ECONNRESET
      Zeus.ui.error "Zeus doesn't seem to be running, try 'zeus start`"
      abort
    end

    def make_winch_channel
      winch, winch_ = IO.pipe
      trap("WINCH") { winch_ << "\0" }
      winch
    end

    def handle_winch
      @winch.read(1)
      set_winsize
      begin
        Process.kill("WINCH", pid) if pid
      rescue Errno::ESRCH
        exit # the remote process died. Just quit.
      end
    end

    def handle_stdin(buffer)
      input = $stdin.readpartial(4096, buffer)
      input.scan(SIGNAL_REGEX).each { |signal|
        begin
          Process.kill(SIGNALS[signal], pid)
        rescue Errno::ESRCH
          exit # the remote process died. Just quit.
        end
      }
      @master << input
    end

    def handle_master(buffer)
      $stdout << @master.readpartial(4096, buffer)
    end

    def set_winsize
      $stdout.tty? and @master.winsize = $stdout.winsize
    end

    def maybe_raw(&b)
      if $stdout.tty?
        $stdout.raw(&b)
      else
        b.call
      end
    end

  end
end

__FILE__ == $0 and Zeus::Client.run
