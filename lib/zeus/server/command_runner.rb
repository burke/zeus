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
        reconnect_things!
        @s_acceptor << $$ << "\n"
        reopen_streams(terminal, terminal, terminal)
        ARGV.replace(arguments)

        return_process_exit_status

        run_action
      end

      def return_process_exit_status
        append_at_exit do
          File.open("omg.log","a"){|f|f.puts(@exit_status_socket.inspect)}
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

      def reconnect_things!
        reconnect_activerecord
        restart_girl_friday
      end

      def restart_girl_friday
        return unless defined?(GirlFriday::WorkQueue)
        # The Actor is run in a thread, and threads don't persist post-fork.
        # We just need to restart each one in the newly-forked process.
        ObjectSpace.each_object(GirlFriday::WorkQueue) do |obj|
          obj.send(:start)
        end
      end

      def reconnect_activerecord
        ActiveRecord::Base.clear_all_connections! rescue nil
        ActiveRecord::Base.establish_connection   rescue nil
      end

    end
  end
end
