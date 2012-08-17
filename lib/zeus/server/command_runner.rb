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
        reopen_streams(terminal, terminal, terminal)
        ARGV.replace(arguments)

        run_action
      ensure
        Process.kill(9, 0)
      end

      def run_action
        @action.call
      rescue Exception => error
        ErrorPrinter.new(error).write_to($stderr)
      end

      def reopen_streams(i, o, e)
        $stdin.reopen(i)
        $stdout.reopen(o)
        $stderr.reopen(e)
      end

      def reconnect_activerecord!
        ActiveRecord::Base.clear_all_connections! rescue nil
        ActiveRecord::Base.establish_connection   rescue nil
      end

    end
  end
end
