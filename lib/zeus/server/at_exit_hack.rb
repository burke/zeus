module Kernel

  alias_method :at_exit_without_queue, :at_exit
  def at_exit(&block)
    @@at_exit_blocks ||= []
    @@at_exit_blocks << block
  end
  alias_method :append_at_exit, :at_exit

  def prepend_at_exit(&block)
    @@at_exit_blocks ||= []
    @@at_exit_blocks.unshift(block)
  end

  at_exit_without_queue {
    @@at_exit_blocks.reverse.each(&:call) if defined?(@@at_exit_blocks)
  }

end
