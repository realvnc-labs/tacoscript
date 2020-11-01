[![Coverage Status](https://coveralls.io/repos/github/cloudradar-monitoring/tacoscript/badge.svg)](https://coveralls.io/github/cloudradar-monitoring/tacoscript)
[![Actions Status](https://github.com/cloudradar-monitoring/tacoscript/workflows/Go/badge.svg)](https://github.com/cloudradar-monitoring/tacoscript/actions)
## Overview
Tacoscript library provides functionality for provisioning of remote servers and local machines running on any OS. Tacoscript can be installed as a binary therefore it doesn't require any additional tools or programs on the host system. 

Tacoscript provisions host systems from simple yaml files written in a [SaltStack](https://www.saltstack.com/) driven configuration language. 

Why we need another provisioning tool? Unfortunately the next competitors like Puppet, Ansible or Salt have a limited support for Windows or require installation of additional tools which is not always convenient.

Tacoscript is written in GO, so its provided as an executable binary for a big variety of host OS and platforms and it requires no additional tools to be installed in the host system. You can find the list of supported OS and architectures [here](https://golang.org/doc/install/source#environment). 

## Installation

### As a compiled binary

Jump to [our release page](https://github.com/cloudradar-monitoring/tacoscript/releases/tag/latest) and download a binary for your host OS. Don't forget to download a corresponding md5 file as well.


        # On MacOS
        wget https://github.com/cloudradar-monitoring/tacoscript/releases/download/latest/tacoscript-latest-darwin-amd64.tar.gz
        
        # On linux
        wget https://github.com/cloudradar-monitoring/tacoscript/releases/download/latest/tacoscript-latest-linux-amd64.tar.gz
        
        # On Windows
        Just download https://github.com/cloudradar-monitoring/tacoscript/releases/download/latest/tacoscript-latest-windows-amd64.zip
        Also download https://github.com/cloudradar-monitoring/tacoscript/releases/download/latest/tacoscript-latest-windows-amd64.zip.md5
     
     
Verify the checksum:

    
        #On MacOS
        curl -Ls https://github.com/cloudradar-monitoring/tacoscript/releases/download/latest/tacoscript-latest-darwin-amd64.tar.gz.md5 | sed 's:$: tacoscript-latest-darwin-amd64.tar.gz:' | md5sum -c
        
        #On linux
         curl -Ls https://github.com/cloudradar-monitoring/tacoscript/releases/download/latest/tacoscript-latest-linux-386.tar.gz.md5 | sed 's:$: tacoscript-latest-linux-amd64.tar.gz:' | md5sum -c
         
        #On Windows assuming you're in the directory with the donwloaded file
         CertUtil -hashfile tacoscript-latest-linux-amd64.tar.gz MD5
        
        #The output will be :
         MD5 hash of tacoscript-0.0.4pre-windows-amd64.zip:
         7103fcda170a54fa39cf92fe816833d1
         CertUtil: -hashfile command completed successfully.
        
        #Compare the command output to the contents of file tacoscript-latest-windows-amd64.zip.md5 they should match
  
  
    
_Note: if the checksums didn't match please don't continue the installation!_

Unpack and install the tacoscript binary on your host machine

    
        #On linux/MacOS
        tar -xzvOf tacoscript-0.0.4pre-darwin-amd64.tar.gz >> /usr/local/bin/tacoscript
        chmod +x /usr/local/bin/tacoscript
    

For Windows

Extract file contents:
![C:\Downloads](docs/Extract.png?raw=true "Extract")

Create a `Tacoscript` folder in `C:\Program Files`

![C:\Program Files](docs/ProgramFiles.png?raw=true "ProgramFiles")

Copy the tacoscript.exe binary to `C:\Program Files\Tacoscript`
![C:\Program Files\Tacoscript\tacoscript.exe](docs/ProgramFilesWithTacoscript.png?raw=true "ProgramFilesWithTacoscript")

Double click on the tacoscript.exe and allow it's execution:

![C:\Program Files\Tacoscript\tacoscript.exe](docs/AllowRun.png?raw=true "AllowRun")

## Install as a go binary:

    go get github.com/cloudradar-monitoring/tacoscript

## Program execution

1. Prepare a configuration script in the yaml format (see the _Configuration_ section below), e.g. `tascoscript.yaml`

On Linux/MacOS

The tacoscript binary expects a configuration file to be provided in the input, to execute script do:

    /usr/local/bin/tacoscript tascoscript.yaml

You can also output the execution details with -v flag:

    /usr/local/bin/tacoscript -v tascoscript.yaml
    
On Windows

1. Place the `tacoscript.yaml` file in `C:\Program Files\Tacoscript\`

2. Right-click on tacoscript.exe and create a shortcut with following properties:

    Target: `C:\Windows\System32\cmd.exe /k tacoscript.exe -v tascoscript.yaml`
    
    Start In: `C:\Program Files\Tacoscript\`
 
![C:\Program Files\Tacoscript\tacoscript.exe](docs/Shortcut.png?raw=true "Shortcut")

3. Click on the shortcut to execute the tacoscript.exe

## Configuration

The configuration file has yaml format. The file consists of a list of scripts which define a single desired state of the host system. A desired state can be an installed program or a file or a running service. 

Here is the minimal possible tacoscript.yaml:


    # unique id of the script, can be any string
    create-file:
      #task type which should be executed
      cmd.run:
        - name: touch /tmp/somefile.txt
            
We can read the script as:

    
    We have a script with the name `create-file`. It has 1 task of type `cmd.run` which executes `touch /tmp/somefile.txt` command. The desired result of this script execution would be an empty file at `/tmp/somefile.txt`.
    

### Scripts
The tacoscript.yaml file contains a collection of scripts. Each script defines a desired state of the host system. You can add as many scripts as you want. The tacoscript binary will execute scripts from file sequentially. If any script failures, the program will stop the execution.

### Tasks
Each script contains a collection of tasks. Each task has a unique type id which identifies the kind of operation the task can do. Each task gets parameters list specified under it as input data. In the example above the task `cmd.run` receives parameter -name with value `/tmp/somefile.txt` and interprets it as a command which should be executed.  

### Task types

- [cmd.run](docs/modules/cmd/README.md)
- [file.managed](docs/modules/file/README.md)
- [pkg.installed](docs/modules/pkg/README.md#pkg.installed)
- [pkg.uptodate](docs/modules/pkg/README.md#pkg.uptodate)
- [pkg.removed](docs/modules/pkg/README.md#pkg.removed)

### Templates
See [Templates rendering](docs/general/templates/README.md)

### Known limitations
- to use shell pipes, redirects or glob expands, please specify a `shell` parameter
- `user` parameter will require sudo rights for tacoscript, in Windows this parameter is ignored
- if you use cmd.run tasks in Windows, you'd better specify the shell parameter as `cmd.exe`, otherwise you will get errors like:
    `exec: "xxx": executable file not found in %PATH%`
- the order of the scripts is not guaranteed. If you don't use the [require](docs/general/dependencies/require.md) values, the scripts will be executed in any order.

## Development instructions

### You can run unit-tests as:

    make test
    
### Execute static code analytic tools

1. Install golangci-lint using instructions from [this site](https://golangci-lint.run/usage/install/)

2. Run the tool


    make sca

### Compile tacoscript binary for your host OS

    make build
    
### Compile tacoscript binary for Windows

    make build-win
