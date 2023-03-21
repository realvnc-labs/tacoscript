<!-- markdownlint-disable -->

## Overview

<!-- markdownlint-restore -->

![Logo](https://raw.githubusercontent.com/realvnc-labs/tacoscript/master/logo.svg)

Tacoscript is a command line tool for the automation of tasks, the management of the operating system and applications.  
It can be installed as a single dependency-free binary. It doesn't require any scripting languages. Any Unix/Linux and
Microsoft Windows are supported equally.

Tacoscript can manage host systems from simple yaml files written in a [Salt Project](https://saltproject.io/) inspired configuration language.

Why do we need another provisioning tool? All competitors like Puppet, Ansible, or Salt have limited
support for Windows. And they require the installation of additional dependencies, which is not always convenient.
Tacoscript is the first tool – written in Go – that is shipped as a static binary. A big variety of host OS and platforms,
in a convenient way.

Tacoscript is declarative. You define how the system, a file or a software should look like after tacoscript has run.
Before each run, Tacoscript compares the desired outcome with the current state. Only the missing steps are preformed.

## Program execution

Prepare a script in the yaml format , e.g. `tascoscript.yaml`, then execute it.

```shell
tacoscript tascoscript.yaml
```

You can also output the execution details with `-v` flag:

```shel
tacoscript -v tascoscript.yaml
```

_You can use any file extension. Using `.taco` for example is fine too._

## Scripting

The script file uses the yaml format. The file consists of a list of tasks that define the states of the host system.
The desired state can be an installed program or files with pre-defined content.

Here is the minimal possible `tacoscript.yaml` for Unix:

```yaml
# unique id of the task, can be any string
create-file:
  #task type (function) to be executed
  cmd.run:
    #Paramter, command to run
    - name: touch /tmp/somefile.txt
```

On Windows, the file can be:

```yaml
create-file:
  cmd.run:
    - name: New-Item -ItemType file C:\Users\Public\Documents\somefile.txt
    - shell: powershell
```

We can read the script as:

> Inside a script, we have a task with the id `create-file`. It consists of the function `cmd.run` which executes
> `touch /tmp/somefile.txt` or its PowerShell equivalent. The desired result of this script execution would be an empty
> file at `/tmp/somefile.txt`.

### Scripts

The tacoscript.yaml file contains a collection of tasks. Each task defines a desired state of the host system.
You can add as many tasks as you want. The tacoscript binary will execute tasks from the file sequentially.

### Tasks

Each script contains a collection of tasks. Each task has a unique id, and a function that identifies the kind of
operation the task can do. Tasks get a list of parameters under it as input data. In the example above, the function
`cmd.run` receives the parameter `-name` with value `/tmp/somefile.txt` and interprets it as a command which should be
executed.

### Supported functions aka task types

- `cmd.run` Run shell commands and scripts [Read more](https://tacoscript.io/functions/commands/)
- `file.managed` copy, manipulate, download and manage files [Read More](https://tacoscript.io/functions/file/)
- `pkg.installed` install packages via package manager [Read More](https://tacoscript.io/functions/packages/#pkginstalled)
- `pkg.uptodate` update packages via package manager [Read More](https://tacoscript.io/functions/packages/#pkguptodate)
- `pkg.removed` remove packages via package manager [Read More](https://tacoscript.io/functions/packages/#pkgremoved)

[Read full documentation]

### Templates

See [Templates rendering](https://tacoscript.io/get-started/template-engine/)

### Known limitations

- to use shell pipes, redirects or glob expands, please specify a `shell` parameter
- `user` parameter will require sudo rights for tacoscript, in Windows this parameter is ignored
- if you use cmd.run tasks in Windows, you'd better specify the shell parameter as `cmd.exe`, otherwise you will get errors like:
  `exec: "xxx": executable file not found in %PATH%`
- the order of the scripts is not guaranteed. If you don't use the [require](https://tacoscript.io/get-started/dependencies/)
  values, the scripts will be executed in any order.

## Development instructions

### Unit-tests

You can run unit-tests as:

```shell
make test
```

### Static Code Analyses

Execute static code analytic tools

1. Install golangci-lint using instructions from [this site](https://golangci-lint.run/usage/install/)

1. Run the tool using `make sca`

### Compile tacoscript binary for your host OS

1. Compile tacoscript binary for Unix with `make build`.
1. Compile tacoscript binary for Windows with `make build-win`
