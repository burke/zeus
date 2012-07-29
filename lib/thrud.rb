class Thrud

  class Task < Struct.new(:method_name, :desc, :long_desc, :method_options)
    def arity(obj)
      obj.method(method_name).arity
    end
  end

  def self.desc(name, a)
    @desc = a
  end
  def self.long_desc(a)
    @long_desc = a
  end
  def self.method_option(*a)
    @method_options ||= []
    @method_options << a
  end

  def self.method_added(m)
    desc, long_desc, method_options = @desc, @long_desc, (@method_options||[])
    @desc, @long_desc, @method_options = nil, nil, nil

    @tasks ||= {}
    @tasks[m.to_s] = Task.new(m, desc, long_desc, method_options)
  end

  def self.map(a)
    a.each do |aliases, target|
      aliases = [aliases] unless aliases.kind_of?(Array)
      aliases.each do |name|
        @tasks[name.to_s] = @tasks[target.to_s]
      end
    end
  end

  def self.task_for_name(name)
    @tasks[name.to_s]
  end

  def task_for_name(name)
    self.class.task_for_name(name)
  end

  def help(taskname = nil)
    if taskname && task = task_for_name(taskname)
      arity = task.arity(self)
      puts <<-BANNER
Usage:
  zeus #{taskname} #{arity == -1 ? "[ARGS]" : ""}

Description:
  #{task.long_desc || task.desc}
BANNER
    else
      # this is super non-generic. problem for future-burke.
      project_tasks = self.class.instance_variable_get("@tasks").
        reject{|k,_|['init', 'start', 'help'].include?(k)}.values.uniq

      tasks = project_tasks.map { |task|
        "  zeus %-14s # %s" % [task.method_name, task.desc]
      }

      puts <<-BANNER
Global Commands:
  zeus help           # show this help menu
  zeus help [COMMAND] # show help for a specific command
  zeus init           # #{task_for_name(:init).desc}
  zeus start          # #{task_for_name(:start).desc}

Project-local Commands:
#{tasks.join("\n")}
BANNER
    end
  end

  def self.start
    taskname = ARGV.shift
    arguments = ARGV

    taskname == "" and taskname = "help"

    unless task = @tasks[taskname.to_s]
      Zeus.ui.error "Could not find task \"#{taskname}\""
      exit 1
    end

    instance = new
    if instance.method(task.method_name).arity == 0 && arguments.any?
      Zeus.ui.error "\"#{task.method_name}\" was called incorrectly. Call as \"zeus #{task.method_name}\"."
      exit 1
    end
    instance.send(task.method_name, *arguments)
  end

end
