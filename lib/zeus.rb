module Zeus

  autoload :UI,      'zeus/ui'
  autoload :CLI,     'zeus/cli'
  autoload :Dsl,     'zeus/dsl'
  autoload :Server,  'zeus/server'
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
