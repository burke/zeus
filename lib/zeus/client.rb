# encoding: utf-8
begin
  require "io/console"
rescue LoadError
  Zeus.ui.error "io/console not found. Please `gem install io-console` or, preferably, " +
    "install ruby 1.9.3 by following the instructions at: " +
    "https://gist.github.com/1688857"
  exit 1
end
require "json"
require "pty"
require "socket"

require 'zeus/client/winsize'

module Zeus
  class Client
    include Winsize

    attr_accessor :pid

    SIGNALS = {
      "\x03" => "INT",
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
          make_winch_channel
          @pid = connect_to_server(command, args, slave)

          select_loop!
        end
      end
    end

    private

    def select_loop!
      buffer = ""
      while ready = select([winch, @master, $stdin])[0]
        handle_winch          if ready.include?(winch)
        handle_stdin(buffer)  if ready.include?($stdin)
        handle_master(buffer) if ready.include?(@master)
      end
    rescue EOFError
    end

    def connect_to_server(command, arguments, slave, socket_path = Zeus::SOCKET_NAME)
      socket = UNIXSocket.new(socket_path)
      socket << {command: command, arguments: arguments}.to_json << "\n"
      socket.send_io(slave)
      slave.close

      pid = socket.readline.chomp.to_i
    rescue Errno::ENOENT, Errno::ECONNREFUSED, Errno::ECONNRESET
      # we need a \r at the end because the terminal is in raw mode.
      Zeus.ui.error "Zeus doesn't seem to be running, try 'zeus start`\r"
      exit 1
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

    def maybe_raw(&b)
      if $stdout.tty?
        $stdout.raw(&b)
      else
        b.call
      end
    end

  end
end
