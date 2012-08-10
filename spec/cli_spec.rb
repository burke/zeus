require 'spec_helper'

module Zeus
  describe CLI do
    let(:ui) { stub(debug!: nil) }

    before do
      Zeus::UI.stub(new: ui)
    end

    it "fails with unknown command" do
      ui.should_receive(:error).with(/Could not find task/m)
      Thrud.should_receive(:exit).with(1).and_raise(SystemExit)
      begin
        run_with_args("foo")
      rescue SystemExit
      end
    end

    describe "#help" do
      it "prints a generic help menu" do
        ui.should_receive(:info).with(/Global Commands.*zeus help.*show this help menu/m)
        run_with_args("help")
      end

      it "prints a usage menu per command" do
        ui.should_receive(:info).with(/Usage:.*zeus version.*version information/m)
        run_with_args(["help", "version"])
      end
    end

    describe "#start" do
      it "fails to start the zeus server in a non-rails project without a config" do
        ui.should_receive(:error).with(/is missing a config file.*rails project.*zeus init/m)
        run_with_args("start", :exit => 1)
      end

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
      it "currently only generates a rails file, even if the project doesn't look like rails" do
        ui.should_receive(:info).with(/Writing new .zeus.rb/m)
        run_with_args("init")
        read(".zeus.rb").should include("config/application")
      end

      it "prints an error and exits if the project already has a zeus config" do
        write(".zeus.rb", "FOO")
        ui.should_receive(:error).with(/.zeus.rb already exists /m)
        run_with_args("init", :exit => 1)
        read(".zeus.rb").should == "FOO"
      end
    end

    describe "generated tasks" do
      it "displays generated tasks in the help menu" do
        ui.should_receive(:info).with(/spec/)
        run_with_args("help")
      end
    end

    private

    def run_with_args(args, options={})
      ARGV.replace([*args])
      if options[:exit]
        Zeus::CLI.any_instance.should_receive(:exit).with(options[:exit]).and_raise(SystemExit)
        begin
          Zeus::CLI.start
        rescue SystemExit
        end
      else
        Zeus::CLI.start
      end
    end
  end
end

