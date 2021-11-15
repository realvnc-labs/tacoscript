[![Coverage Status](https://coveralls.io/repos/github/cloudradar-monitoring/tacoscript/badge.svg)](https://coveralls.io/github/cloudradar-monitoring/tacoscript)
[![Actions Status](https://github.com/cloudradar-monitoring/tacoscript/workflows/Go/badge.svg)](https://github.com/cloudradar-monitoring/tacoscript/actions)

![](logo.svg)
## Overview
Tacoscript library provides functionality for provisioning of remote servers and local machines running on any OS. Tacoscript can be installed as a binary. Therefore it doesn't require any additional tools or programs on the host system. 

Tacoscript can manage host systems from simple yaml files written in a [Salt Project](https://saltproject.io/) inspired configuration language. 

Why do we need another provisioning tool? Unfortunately, the next competitors like Puppet, Ansible, or Salt have limited support for Windows. And they require the installation of additional dependencies, which is not always convenient.

Tacoscript is written in GO, so it is provided as an executable static binary for a big variety of host OS and platforms, and it requires no additional tools to be installed in the host system. You can find the list of supported OS and architectures [here](https://golang.org/doc/install/source#environment). 

## Installation

### As a compiled binary

Jump to [our release page](https://github.com/cloudradar-monitoring/tacoscript/releases/tag/latest) and download a binary for your host OS. Don't forget to download a corresponding md5 file as well and compare the checksums.

#### On MacOS
    wget https://github.com/cloudradar-monitoring/tacoscript/releases/download/latest/tacoscript-latest-darwin-amd64.tar.gz
    tar xzf tacoscript-latest-darwin-amd64.tar.gz -C /usr/local/bin/ tacoscript
    
#### On linux
    wget https://github.com/cloudradar-monitoring/tacoscript/releases/download/latest/tacoscript-latest-linux-amd64.tar.gz
    tar xzf tacoscript-latest-linux-amd64.tar.gz -C /usr/local/bin/ tacoscript

#### On Windows
    iwr https://github.com/cloudradar-monitoring/tacoscript/releases/download/latest/tacoscript-latest-windows-amd64.zip `
    -OutFile tacoscript-latest-windows-amd64.zip
    $dest = "C:\Program Files\tacoscript"
    mkdir $dest
    mkdir "$($dest)\bin"
    Expand-Archive -Path $file -DestinationPath $dest -force
    mv "$($dest)\tacoscript.exe" "$($dest)\bin"

    $ENV:PATH="$ENV:PATH;$($dest)\bin"

    [Environment]::SetEnvironmentVariable(
        "Path",
        [Environment]::GetEnvironmentVariable(
            "Path", [EnvironmentVariableTarget]::Machine
        ) + ";$($dest)\bin",
        [EnvironmentVariableTarget]::Machine
    )
    & tacoscript --version
     
## Install as a go binary:

    go get github.com/cloudradar-monitoring/tacoscript

## Program execution

Prepare a  script in the yaml format (see the _Configuration_ section below), e.g. `tascoscript.yaml`
The tacoscript binary expects a script file to be provided in the input:

    tacoscript tascoscript.yaml

You can also output the execution details with `-v` flag:

    tacoscript -v tascoscript.yaml

_You can use any file extension. Using `.taco` for example is fine too._
## Scripting

The script file uses the yaml format. The file consists of a list of tasks that define the states of the host system. The desired state can be an installed program or files with pre-defined content. 

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
> Inside a script, we have a task with the id `create-file`. It consists of the function `cmd.run` which executes `touch /tmp/somefile.txt` or its PowerShell equivalent. The desired result of this script execution would be an empty file at `/tmp/somefile.txt`.
    

### Scripts
The tacoscript.yaml file contains a collection of tasks. Each task defines a desired state of the host system. You can add as many tasks as you want. The tacoscript binary will execute tasks from the file sequentially.

![](docs/script-structure.png)

### Tasks
Each script contains a collection of tasks. Each task has a unique id, and a function that identifies the kind of operation the task can do. Tasks get a list of parameters under it as input data. In the example above, the function `cmd.run` receives the parameter `-name` with value `/tmp/somefile.txt` and interprets it as a command which should be executed.  

### Task types

- [cmd.run](docs/functions/cmd/README.md)
- [file.managed](docs/functions/file/README.md)
- [pkg.installed](docs/functions/pkg/README.md#pkginstalled)
- [pkg.uptodate](docs/functions/pkg/README.md#pkguptodate)
- [pkg.removed](docs/functions/pkg/README.md#pkgremoved)

### Templates
See [Templates rendering](docs/general/templates/README.md)

### Known limitations
- to use shell pipes, redirects or glob expands, please specify a `shell` parameter
- `user` parameter will require sudo rights for tacoscript, in Windows this parameter is ignored
- if you use cmd.run tasks in Windows, you'd better specify the shell parameter as `cmd.exe`, otherwise you will get errors like:
    `exec: "xxx": executable file not found in %PATH%`
- the order of the scripts is not guaranteed. If you don't use the [require](docs/general/dependencies/require.md) values, the scripts will be executed in any order.

## Development instructions

### Unit-tests
You can run unit-tests as:

    make test
    
### Static Code Analyses
Execute static code analytic tools

1. Install golangci-lint using instructions from [this site](https://golangci-lint.run/usage/install/)

1. Run the tool using `make sca`

### Compile tacoscript binary for your host OS
1. Compile tacoscript binary for Unix with `make build`.
1. Compile tacoscript binary for Windows with `make build-win`
