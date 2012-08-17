#!/usr/bin/env rake
require "bundler/gem_tasks"

require 'rspec/core/rake_task'
task :spec do
  desc "Run specs under spec/"
  RSpec::Core::RakeTask.new do |t|
    t.pattern = 'spec/**/*_spec.rb'
  end
end

task default: :spec
