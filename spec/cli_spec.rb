require 'zeus'

module Zeus
  describe CLI do

    let(:ui) { stub(debug!: nil) }

    before do
      Zeus::UI.stub(new: ui)
    end

    describe "#help" do
      it "prints a generic help menu" do
        ui.should_receive(:info).with(/Global Commands.*zeus help.*show this help menu/m)
        run_with_args("help")
      end

      it "prints a usage menu per command" do
        ui.should_receive(:info).with(/Usage:.*zeus version.*version information/m)
        run_with_args("help", "version")
      end
    end

    describe "#start" do
      it "starts the zeus server"
      it "uses the rails template file if the project is missing a config file but looks like rails"
      it "prints an error and exits if there is no config file and the project doesn't look like rails"
    end

    describe "#version" do
      STRING_INCLUDING_VERSION = %r{#{Regexp.escape Zeus::VERSION}}

      it "prints the version and exits" do
        ui.should_receive(:info).with(STRING_INCLUDING_VERSION)
        run_with_args("version")
      end

      it "has aliases" do
        ui.should_receive(:info).with(STRING_INCLUDING_VERSION).twice
        run_with_args("--version")
        run_with_args("-v")
      end

    end

    describe "#init" do
      it "currently only generates a rails file, even if the project doesn't look like rails"
      it "prints an error and exits if the project already has a zeus config"
    end

    describe "generated tasks" do
      it "displays generated tasks in the help menu" do
        ui.should_receive(:info).with(/spec/)
        run_with_args("help")
      end
    end

    private

    def run_with_args(*args)
      ARGV.replace(args)
      Zeus::CLI.start
    end

  end
end

