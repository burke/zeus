module Zeus
  class UI

    def initialize
      @quiet = false
      @debug = ENV['DEBUG']
    end

    def info(msg)
      tell_me(msg, :magenta) if !@quiet
    end

    def warn(msg)
      tell_me(msg, :yellow)
    end

    def error(msg)
      tell_me(msg, :red)
    end

    def debug(msg)
      tell_me(msg, nil) if debug?
    end

    def be_quiet!
      @quiet = true
    end

    def debug!
      @debug = true
    end

    def debug?
      !!@debug && !@quiet
    end

    private

    def tell_me(msg, color = nil)
      puts make_message(msg, color)
    end

    def make_message(msg, color)
      msg = case color
            when :red     ; "\x1b[31m#{msg}\x1b[0m"
            when :green   ; "\x1b[32m#{msg}\x1b[0m"
            when :yellow  ; "\x1b[33m#{msg}\x1b[0m"
            when :magenta ; "\x1b[35m#{msg}\x1b[0m"
            else          ; msg
            end
      msg[-1] == "\n" ? msg : "#{msg}\n"
    end


  end

end
