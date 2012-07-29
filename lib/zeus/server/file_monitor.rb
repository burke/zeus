require 'rb-kqueue'

module Zeus
  class Server
    class FileMonitor

      TARGET_FD_LIMIT = 8192

      def initialize(&change_callback)
        configure_file_descriptor_resource_limit
        @queue = KQueue::Queue.new
        @watched_files = {}
        @deleted_files = []
        @change_callback = change_callback
      end

      def process_events
        @queue.poll
      end

      def watch(file)
        return if @watched_files[file]
        @watched_files[file] = true
        @queue.watch_file(file, :write, :extend, :rename, :delete, &method(:file_did_change))
      rescue Errno::ENOENT
        Zeus.ui.debug("No file found at #{file}")
      end

      private

      def file_did_change(event)
        Zeus.ui.info("Dependency change at #{event.watcher.path}")
        resubscribe_deleted_file(event) if event.flags.include?(:delete)
        @change_callback.call(event.watcher.path)
      end

      def configure_file_descriptor_resource_limit
        limit = Process.getrlimit(Process::RLIMIT_NOFILE)
        if limit[0] < TARGET_FD_LIMIT && limit[1] >= TARGET_FD_LIMIT
          Process.setrlimit(Process::RLIMIT_NOFILE, TARGET_FD_LIMIT)
        else
          Zeus.ui.warn "Warning: increase the max number of file descriptors. If you have a large project, this max cause a crash in about 10 seconds."
        end
      end

      def resubscribe_deleted_file(event)
        event.watcher.disable!

        @queue.watch_file(event.watcher.path, :write, :extend, :rename, :delete, &method(:file_did_change))
      rescue Errno::ENOENT
        @deleted_files << event.watcher.path
      end

    end
  end
end
