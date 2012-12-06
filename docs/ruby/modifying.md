# Modifying the boot process

Running `zeus init` creates a default [`zeus.json`](/burke/zeus/tree/master/examples/custom_plan/zeus.json) and [`custom_plan.rb`](/burke/zeus/tree/master/examples/custom_plan/custom_plan.rb) that define the boot order for your application. 

Each node under "plan" in the JSON file has an action associated with it. You can see the "command" value at the top of the hash requires `"./custom_plan.rb"`. Inspecting that file, you can see that it creates an empty subclass of `Zeus::Rails`. If you look at [that file](/burke/zeus/tree/master/rubygem/lib/zeus/rails.rb), you will see a class with a method name corresponding to each node -- something along the lines of:

```ruby
class Zeus::Rails < Zeus::Plan
  def boot
    # ...
  end
  def default_bundle
    # ...
  end
  # ...
end
```

Note that an instance of the subclass class is assigned to `Zeus.plan` at the end of `custom_plan.rb`. Zeus calls methods on `Zeus.plan` to boot the application. If you follow any path to a leaf node in the tree -- for example, boot, default_bundle, development_environment, prerake -- those methods are essentially called in sequence to construct an environment for a command (rake, in this case) to run in. Zeus forks the ruby process between each step, and can restart from any of these forks.

You can modify the plan by adding/removing/moving nodes in the json file and adding the corresponding methods in `custom_plan.rb`.

```ruby
# custom_plan.rb
require 'zeus/rails'
class CustomPlan < Zeus::Rails
  def boot
    something_else
  end
end
Zeus.plan = CustomPlan.new
```

Note that there's nothing special about the naming or location of `CustomPlan`. Feel free to rename and/or move as you please, just remember to update the `command` line in `zeus.json` -- and `zeus.json` must stay at your project root, or Zeus will just use the default configuration.
