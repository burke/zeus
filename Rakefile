#!/usr/bin/env rake
require "bundler/gem_tasks"

require 'rspec/core/rake_task'
require 'rake/clean'

# rule to build inotify-wrapper
file "ext/inotify-wrapper/inotify-wrapper" =>
    Dir.glob("ext/inotify-wrapper/*{.c}") do
  Dir.chdir("ext/inotify-wrapper") do
    ruby "extconf.rb"
    sh "make"
  end
end

desc "Compile inotify-wrapper"
task :compile => "ext/inotify-wrapper/inotify-wrapper"

CLEAN.include("ext/inotify-wrapper/{inotify-wrapper,*.o}")
CLEAN.include("ext/inotify-wrapper/Makefile")
CLOBBER.include("ext/inotify-wrapper/inotify-wrapper")

task :spec do
  desc "Run specs under spec/"
  RSpec::Core::RakeTask.new do |t|
    t.pattern = 'spec/**/*_spec.rb'
  end
end

task :default do
  puts "rake fails when run via bundle exec"
  Rake::Task["spec"].invoke
end
