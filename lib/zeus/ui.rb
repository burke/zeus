module Zeus
  class UI

    def initialize
      @quiet = false
      @debug = ENV['DEBUG']
    end

    def info(msg)
      tell_me(msg, nil) if !@quiet
    end

    def confirm(msg)
      tell_me(msg, :green) if !@quiet
    end

    def warn(msg)
      tell_me(msg, :yellow)
    end

    def error(msg)
      tell_me(msg, :red)
    end

    def be_quiet!
      @quiet = true
    end

    def debug?
      # needs to be false instead of nil to be newline param to other methods
      !!@debug && !@quiet
    end

    def debug!
      @debug = true
    end

    def debug(msg)
      tell_me(msg, nil) if debug?
    end

    private
    def tell_me(msg, color = nil)
      msg = case color
            when :red    ; "\x1b[31m#{msg}\x1b[0m"
            when :green  ; "\x1b[32m#{msg}\x1b[0m"
            when :yellow ; "\x1b[33m#{msg}\x1b[0m"
            else         ; msg
            end
      puts msg
    end


  end

end
