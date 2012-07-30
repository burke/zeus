module Zeus
  SOCKET_NAME = '.zeus.sock'

  autoload :UI,      'zeus/ui'
  autoload :CLI,     'zeus/cli'
  autoload :DSL,     'zeus/dsl'
  autoload :Server,  'zeus/server'
  autoload :Client,  'zeus/client'
  autoload :Version, 'zeus/version'

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

end
