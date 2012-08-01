module Process

  def self.killall_descendants(sig, base=Process.pid)
    descendants(base).each do |pid|
      begin
        Process.kill(sig, pid)
      rescue Errno::ESRCH
      end
    end
  end

  def self.pids_to_ppids
    Hash[*`ps -eo pid,ppid`.scan(/\d+/).map(&:to_i)]
  end

  def self.descendants(base=Process.pid)
    descendants = Hash.new{|ht,k| ht[k]=[k]}

    pids_to_ppids.each do |pid,ppid|
      descendants[ppid] << descendants[pid]
    end

    descendants[base].flatten - [base]
  end
end

