require 'open3'

module Zeus
  class Server
    module FileMonitor
      class FSEvent
        WRAPPER_PATH = File.expand_path("../../../../../ext/fsevents-wrapper/fsevents-wrapper", __FILE__)

        def datasource          ; @io_out ; end
        def on_datasource_event ; handle_changed_files ; end
        def close_child_socket  ; end
        def close_parent_socket ; [@io_in, @io_out].each(&:close) ; end

        def initialize(&change_callback)
          @change_callback = change_callback
          @io_in, @io_out, _ = Open3.popen2e(WRAPPER_PATH)
          @watched_files = {}
          @buffer = ""
        end

        def handle_changed_files
          50.times { read_and_notify_files }
        rescue Errno::EAGAIN
        end

        def read_and_notify_files
          lines = @io_out.read_nonblock(30000)
          files = lines.split("\n")
          files[0] = "#{@buffer}#{files[0]}" unless @buffer == ""
          unless lines[-1] == "\n"
            @buffer = files.pop
          end

          files.each do |file|
            file_did_change(file)
          end
        end

        def watch(file)
          return false if @watched_files[file]
          @watched_files[file] = true
          @io_in.puts file
          true
        end

        private

        def file_did_change(file)
          Zeus.ui.info("Dependency change at #{file}")
          @change_callback.call(file)
        end

      end
    end
  end
end



