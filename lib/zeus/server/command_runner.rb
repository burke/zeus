module Zeus
  class Server
    class CommandRunner

      def initialize(name, action, s_acceptor)
        @name = name
        @action = action
        @s_acceptor = s_acceptor
      end

      def run(terminal, exit_status_socket, arguments)
        child = fork { _run(terminal, exit_status_socket, arguments) }
        terminal.close
        exit_status_socket.close
        Process.detach(child)
        child
      end

      private

      def _run(terminal, exit_status_socket, arguments)
        $0 = "zeus runner: #{@name}"
        @exit_status_socket = exit_status_socket
        @terminal = terminal
        Process.setsid
        reconnect_activerecord!
        @s_acceptor << $$ << "\n"
        reopen_streams(terminal, terminal, terminal)
        ARGV.replace(arguments)

        return_process_exit_status

        run_action
      end

      def return_process_exit_status
        at_exit do
          if $!.nil? || $!.is_a?(SystemExit) && $!.success?
            @exit_status_socket.puts(0)
          else
            code = $!.is_a?(SystemExit) ? $!.status : 1
            @exit_status_socket.puts(code)
          end

          @exit_status_socket.close
          @terminal.close
        end
      end

      def run_action
        @action.call
      rescue StandardError => error
        ErrorPrinter.new(error).write_to($stderr)
        raise
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
