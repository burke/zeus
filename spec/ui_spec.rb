require 'zeus'

module Zeus
  describe UI do

    let(:ui) {
      ui = UI.new
      # override this method to return the result rather than printing it.
      def ui.tell_me(msg, color)
        return make_message(msg, color)
      end
      ui
    }

    it "prints errors in red, regardless of verbosity level" do
      ui.error("error").should == "\x1b[31merror\x1b[0m\n"
      ui.be_quiet!
      ui.error("error").should == "\x1b[31merror\x1b[0m\n"
    end

    it "prints warnings in yellow, regardless of verbosity level" do
      ui.warn("warning").should == "\x1b[33mwarning\x1b[0m\n"
      ui.be_quiet!
      ui.warn("warning").should == "\x1b[33mwarning\x1b[0m\n"
    end

    it "prints info messages in magenta, but not if quiet-mode is set" do
      ui.info("info").should == "\x1b[35minfo\x1b[0m\n"
      ui.be_quiet!
      ui.info("info").should == nil
    end

    it "doesn't print debug messages by default" do
      ui.debug("debug").should == nil
    end

    it "prints debug messages if debug-mode is set" do
      ui.debug!
      ui.debug("debug").should == "debug\n"
    end

    it "sets debug if ENV['DEBUG']" do
      ENV['DEBUG'] = "yup"
      ui.debug?.should be_true
    end

    it "doesn't print debug messages if both quiet-mode and debug-mode are set" do
      ui.be_quiet!
      ui.debug!
      ui.debug("debug").should == nil
    end

  end
end
