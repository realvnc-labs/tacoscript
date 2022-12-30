---
title: "Run Tacoscripts"
weight: 1
slug: run-tacoscripts
resources:
- name: script-structure
  src: script-structure.png
  title: "Script structure"
---

## Create and run your first script

Take your preferred text editor and create a plain text file `yummy-taco.yml`.

```yaml
new-file:
  file.managed:
    - name: my-file.txt
    - contents: |
        I love tacos.
        I can eat them all days.
```

Now execute the file by invoking `tacoscript yummy-taco.yml`. You will get an output like this:

```text
results:
- ID: new-file
  Function: file.managed
  Name: my-file.txt
  Result: true
  Comment: File my-file.txt updated
  Started: "14:18:31.994407"
  Duration: 3.1106ms
  Changes:
    diff: |
      expected: ""I love tacos.\nI can eat them all days.""
      actual: """"
      Diff:
      --- Expected
      +++ Actual
      -I love tacos.
      -I can eat them all days.
      +
    length: 38 bytes written
summary:
  Script: .\playground.yml
  Succeeded: 1
  Failed: 0
  Aborted: 0
  Changes: 0
  TotalTasksRun: 1
  TotalRunTime: 6.1952ms
```

Now run the tacoscript again. Note that the file has not overwritten or changed, because the content of the file is
already in the desired state.

## Structure of a tacoscript file

A tacoscript file consist of one or many tasks. Each task must have a unique task id (per file).
Each task starts with a function and many optional parameters can follow.

{{< img name="script-structure" size="medium" lazy=false >}}

{{< hint type=tip title="Yaml indentation">}}
Remember that yaml only allows indentation by blank space. **Never use tabs!**

You can freely choose by how many blank spaces you want to indent.
{{< /hint>}}
