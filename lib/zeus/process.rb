module Process

  def self.killall_descendants(sig, base=Process.pid)
    descendants(base).each do |pid|
      begin
        Process.kill(sig, pid)
      rescue Errno::ESRCH
      end
    end
  end

  def self.descendants(base=Process.pid)
    descendants = Hash.new{|ht,k| ht[k]=[k]}
    Hash[*`ps -eo pid,ppid`.scan(/\d+/).map{|x|x.to_i}].each{|pid,ppid|
      descendants[ppid] << descendants[pid]
    }
    descendants[base].flatten - [base]
  end
end

