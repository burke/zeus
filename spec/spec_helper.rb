require 'zeus'

module FolderHelpers
  def write(file, content)
    ensure_folder File.dirname(file)
    File.open(file, 'w'){|f| f.write content }
  end

  def read(file)
    File.read file
  end

  def delete(file)
    `rm #{file}`
  end

  def ensure_folder(folder)
    `mkdir -p #{folder}` unless File.exist?(folder)
  end

  def root
    File.expand_path '../..', __FILE__
  end
end

module ProcessHelpers
  def kill_all_children
    pids = child_pids
    `kill -s KILL #{pids.join " "} 2>/dev/null` unless pids.empty?
  end

  def child_pids
    base = Process.pid
    descendants = Hash.new{ |ht,k| ht[k]=[k] }
    Hash[*`ps -eo pid,ppid`.scan(/\d+/).map{ |x| x.to_i }].each do |pid, ppid|
      descendants[ppid] << descendants[pid]
    end
    descendants[base].flatten - [base]
  end
end

RSpec.configure do |config|
  config.include FolderHelpers
  config.include ProcessHelpers

  config.around do |example|
    folder = File.expand_path("../tmp", __FILE__)
    `rm -rf #{folder}`
    ensure_folder folder
    Dir.chdir folder do
      example.call
    end
    `rm -rf #{folder}`
  end
end
