require "zeus/version"
require 'zeus/server'
require "zeus/ui"

module Zeus
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
