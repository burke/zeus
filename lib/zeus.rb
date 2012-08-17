module Zeus
  SOCKET_NAME = '.zeus.sock'

  autoload :UI,           'zeus/ui'
  autoload :CLI,          'zeus/cli'
  autoload :DSL,          'zeus/dsl'
  autoload :Server,       'zeus/server'
  autoload :Client,       'zeus/client'
  autoload :VERSION,      'zeus/version'
  autoload :ErrorPrinter, 'zeus/error_printer'

  class ZeusError < StandardError
    def self.status_code(code)
      define_method(:status_code) { code }
    end
  end

  def self.ui
    @ui ||= UI.new
  end

  def self.ui=(ui)
    @ui = ui
  end

  def self.after_fork(&b)
    @after_fork ||= []
    @after_fork << b
  end

  def self.run_after_fork!
    @after_fork.map(&:call) if @after_fork
    @after_fork = []
  end

  def self.thread_with_backtrace_on_error(&b)
    Thread.new {
      begin
        b.call
      rescue => e
        ErrorPrinter.new(e).write_to($stdout)
      end
    }
  end

end
