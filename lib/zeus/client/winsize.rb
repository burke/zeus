module Zeus
  class Client
    module Winsize

      attr_reader :winch

      def set_winsize
        $stdout.tty? and @master.winsize = $stdout.winsize
      end

      def make_winch_channel
        @winch, winch_ = IO.pipe
        trap("WINCH") { winch_ << "\0" }
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

    end
  end
end
