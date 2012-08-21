module Kernel

  alias_method :prepend_at_exit, :at_exit
  if ENV['EXIT_DEBUG']
    def at_exit(&block)
      puts "\x1b[36mRegistering block using at_exit in #{caller[0]}\x1b[0m"
      blk = caller[0]
      prepend_at_exit {
        print "\x1b[36mRunning block using at_exit in #{blk}...\x1b[0m" ; $stdout.flush
        block.call
        puts "\x1b[36mdone!\x1b[0m"
      }
    end
  end

  def append_at_exit(&block)
    ENV['EXIT_DEBUG'] and puts "\x1b[36mRegistering block using \x1b[32mappend\x1b[36m in #{caller[0]}\x1b[0m"
    @@at_exit_blocks ||= []
    @@at_exit_blocks << block
  end

  at_exit {
    ENV['EXIT_DEBUG'] and puts "\x1b[36mRunning #{at_exit_blocks.size} at_exit_blocks\x1b[0m"
    at_exit_blocks.each do |b|
      ENV['EXIT_DEBUG'] and puts "\x1b[36mRunning \x1b[32mappended\x1b[36m block #{b.inspect}\x1b[0m"
      b.call
    end
  }

  private

  def at_exit_blocks
    defined?(@@at_exit_blocks) ? @@at_exit_blocks : []
  end

end
