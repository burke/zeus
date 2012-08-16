module Zeus
  class Server
    class CommandRunner

      def initialize(name, action, s_acceptor)
        @name = name
        @action = action
        @s_acceptor = s_acceptor
      end

      def run(terminal, arguments)
        child = fork { _run(terminal, arguments) }
        terminal.close
        Process.detach(child)
        child
      end

      private

      def _run(terminal, arguments)
        $0 = "zeus runner: #{@name}"
        Process.setsid
        reconnect_activerecord!
        @s_acceptor << $$ << "\n"
        $stdin.reopen(terminal)
        $stdout.reopen(terminal)
        $stderr.reopen(terminal)
        ARGV.replace(arguments)

        @action.call
      ensure
        # TODO this is a whole lot of voodoo that I don't really understand.
        # I need to figure out how best to make the process disconenct cleanly.
        dnw, dnr = File.open("/dev/null", "w+"), File.open("/dev/null", "r+")
        $stderr.reopen(dnr)
        $stdout.reopen(dnr)
        terminal.close
        $stdin.reopen(dnw)
        Process.kill(9, $$)
        exit 0
      end

      def reconnect_activerecord!
        ActiveRecord::Base.clear_all_connections! rescue nil
        ActiveRecord::Base.establish_connection   rescue nil
      end

    end
  end
end
