[![Coverage Status](https://coveralls.io/repos/github/cloudradar-monitoring/tacoscript/badge.svg)](https://coveralls.io/github/cloudradar-monitoring/tacoscript)
[![Actions Status](https://github.com/cloudradar-monitoring/tacoscript/workflows/Go/badge.svg)](https://github.com/cloudradar-monitoring/tacoscript/actions)
## Overview
This library provides functionality for remote server provisioning. Practically it's a command line tool which is installed as a binary on Windows/.nix servers. Further on it accepts a list of commands as a simple yaml file and executes them on the host system. As a result you get a fully provisioned remote machine.

The idea was inspired by projects like Salt-Stack or Ansible, however they have a limited possibility to work on Windows. Salt requires the installation of Python and Ansible has no Windows support. 

The project targets beginners and semi-professionals. They would be overwhelmed with Ansible or Salt. This project provides a  Go driven, binary based implementation which doesn't require additional tools on the host system and has a full Windows support. 

## Installation

### As a compiled binary

- Jump to [our release page](https://github.com/cloudradar-monitoring/tacoscript/releases/tag/latest) and download a binary for your host OS. Don't forget to download a corresponding md5 file as well.


    # On MacOS
    wget https://github.com/cloudradar-monitoring/tacoscript/releases/download/latest/tacoscript-latest-darwin-amd64.tar.gz
    
    # On linux
    wget https://github.com/cloudradar-monitoring/tacoscript/releases/download/latest/tacoscript-latest-linux-amd64.tar.gz
    
    # On Windows
    Just download https://github.com/cloudradar-monitoring/tacoscript/releases/download/latest/tacoscript-latest-windows-amd64.zip
    Also download https://github.com/cloudradar-monitoring/tacoscript/releases/download/latest/tacoscript-latest-windows-amd64.zip.md5
     
- Verify the checksum

    
    #On MacOS
    curl -Ls https://github.com/cloudradar-monitoring/tacoscript/releases/download/latest/tacoscript-latest-darwin-amd64.tar.gz.md5 | sed 's:$: tacoscript-latest-darwin-amd64.tar.gz:' | md5sum -c
    
    #On linux
     curl -Ls https://github.com/cloudradar-monitoring/tacoscript/releases/download/latest/tacoscript-latest-linux-386.tar.gz.md5 | sed 's:$: tacoscript-latest-linux-amd64.tar.gz:' | md5sum -c
     
    #On Windows assuming you're in the directory with the donwloaded file
     CertUtil -hashfile tacoscript-latest-linux-amd64.tar.gz MD5
    
    # The output will be :
     MD5 hash of tacoscript-0.0.4pre-windows-amd64.zip:
     7103fcda170a54fa39cf92fe816833d1
     CertUtil: -hashfile command completed successfully.
    
    #Compare the command output to the contents of file tacoscript-latest-windows-amd64.zip.md5 they should match
    
_Note: if the checksums didn't match please don't continue the installation!_

- Unpack and install the tacoscript binary on your host machine

    
    #On linux/MacOS
    tar -xzvOf tacoscript-0.0.4pre-darwin-amd64.tar.gz >> /usr/local/bin/tacoscript
    chmod +x /usr/local/bin/tacoscript
    
- On Windows

Extract file contents:
![C:\Downloads](docs/Extract.png?raw=true "Extract")

Create a `Tacoscript` folder in `C:\Program Files`

![C:\Program Files](docs/ProgramFiles.png?raw=true "ProgramFiles")

Copy the tacoscript.exe binary to `C:\Program Files\Tacoscript`
![C:\Program Files\Tacoscript\tacoscript.exe](docs/ProgramFilesWithTacoscript.png?raw=true "ProgramFilesWithTacoscript")

Double click on the tacoscript.exe and allow it's execution:

![C:\Program Files\Tacoscript\tacoscript.exe](docs/AllowRun.png?raw=true "AllowRun")

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

The configuration file has yaml format. The file consists of a list of scripts which define a single desired state of the host system. A desired state can be an installed program or a file at certain location or a running daemon. 

Here is the minimal tacoscript.yaml:


    # unique id of the script, can be any string
    create-file:
      #task type which should be executed
      cmd.run:
        - name: touch /tmp/somefile.txt
            
We can read the script as:

    
    We have a script with the name `create-file`. It has 1 task of type `cmd.run` which executes `touch /tmp/somefile.txt` in a command line. The desired result of this script execution would be an empty file at `/tmp/somefile.txt`.
    

### Scripts
The tacoscript.yaml file contains a collection of scripts, where each script defines a desired state of the target system. You can add as many scripts as you want. The tacoscript binary will execute scripts one after another rather than parallel. If any script fails, the execution will be stopped and the error will be given in the (stderr) output with a non-zero exit code.

### Tasks
Each script contains a collection of tasks. Each task has a unique type id which identifies the kind of operation the task is doing. Each task gets parameters list specified under it as input e.g. from the example above the task `cmd.run` gets array of names with one element `touch /tmp/somefile.txt` as it's input. The parameters specify the execution details of a task.  

### Task types

_cmd.run_
Executes an arbitrary command in a shell of a host system. Here is the `cmd.run` task with all possible parameters:


    create-file:
      cmd.run:
        - name: echo 'data to backup' >> /tmp/data2001-01-01.txt
    backup-data:
      cmd.run:
        - names:
            - tar czf /dumps/data2001-01-01.txt.tar.gz /tmp/data2001-01-01.txt
            - md5sum /dumps/data2001-01-01.txt.tar.gz >> /dumps/data2001-01-01.txt.tar.gz.md5
        - cwd: /tmp
        - require:
            - create-file
        - creates:
            - /tmp/data2001-01-01.txt.tar.gz
        - env:
            - PASSWORD: bunny
        - shell: bash
        - unless: service backup-runner status
        - onlyif: date +%c|grep -q "^Thu"

You can read interpret this file as following:

The desired state of the script 'create-file' is a file at /tmp/data2001-01-01.txt. To achieve this the tacoscript binary should execute a task of type 'cmd.run', which means execution of command `echo 'data to backup' >> /tmp/data2001-01-01.txt` in the shell of the host system.
The second desired state of the script 'backup-data' is 2 files: archive of `/tmp/data2001-01-01.txt` under `/dumps/data2001-01-01.txt.tar.gz` and it's md5 sum file at `/dumps/data2001-01-01.txt.tar.gz.md5`.
It should be achieved by executing a task of type 'cmd.run' which requires 2 shell operations to be executed:

- `tar czf /dumps/data2001-01-01.txt.tar.gz /tmp/data2001-01-01.txt`
- `md5sum /dumps/data2001-01-01.txt.tar.gz >> /dumps/data2001-01-01.txt.tar.gz.md5`

The commands will be executed in the following context:
- current working directory (cwd) will be `/tmp`
- context with env variables will contain PASSWORD with value `bunny`
- as shell `bash` will be selected

Both commands will be only executed after `create-file` script. So the tacoscript interpreter will make sure that the `create-file` script is executed before `backup-data` and only if it was successful.

The commands will be executed only under the following conditions:
- The file `/tmp/data2001-01-01.txt.tar.gz` should be missing (to avoid data overriding)
- The service `service backup-runner` should not be running
- Today is Thursday

### Task parameters

### name

[string] type

    create-file:
      cmd.run:
        - name: echo 'data to backup' >> /tmp/data2001-01-01.txt
        

Name shows a single executable command. In the example above the tacoscript interpreter will run `echo 'data to backup' >> /tmp/data2001-01-01.txt` command. 

### names
[array] type

    backup-data:
      cmd.run:
        - names:
            - tar czf /dumps/data2001-01-01.txt.tar.gz /tmp/data2001-01-01.txt
            - md5sum /dumps/data2001-01-01.txt.tar.gz >> /dumps/data2001-01-01.txt.tar.gz.md5

Name contains the list of commands. All commands in the task will be executed in the order of appearance. If one fails, the whole execution will stop. All commands inside one task will be executed in the same context, which means in the current working directory, with same env variables, in the same shell and under same conditions.

If you want to change context of a command (e.g. use another shell), you should create another task e.g.


    backup-data:
      cmd.run:
        - name: mycmd in shell 'bash'
        - shell: bash
      cmd.run:
        - name: mycmd in shell 'sh'
        - shell: sh


The `names` parameter with a single value has the same meaning as `name` field. 

### cwd
[string] type

    backup-data:
      cmd.run:
        - cwd: /tmp

The `cwd` parameter gives the current working directory of a command. This value is quite useful if you want to use relative paths of files that you provide to your commands. 

For example imagine following file structure:

    C:\Some\Very\Long\Path
      someData1.txt
      someData2.txt


You want pack someData1.txt and someData2.txt with the zip.exe binary. You can do it with the script as:

    backup-data:
      cmd.run:
        - names:
            - tar.exe -a -c -z -f C:\Some\Very\Long\Path\someData1.txt.tar.gz C:\Some\Very\Long\Path\someData1.txt
            - tar.exe -a -c -z -f C:\Some\Very\Long\Path\someData2.txt.tar.gz C:\Some\Very\Long\Path\someData2.txt


Obviously we have quite long strings here. 
You can also change your working directory and use relative paths to get the same result:

    backup-data:
      cmd.run:
        - cwd: C:\Some\Very\Long\Path\
        - names:
            - tar.exe -a -c -z -f someData1.txt.tar.gz someData1.txt
            - tar.exe -a -c -z -f someData2.txt.tar.gz someData2.txt

### Require
[string] and [array] type

`require` parameter can have both string and array type e.g.

    backup-data:
      cmd.run:
      - name: echo 123
      - require:
        - script1
        - script2
    script1:
        ...
    script2:
        ...

    #or
    backup-data:
      cmd.run:
      - name: echo 123
      - require: script1
    script1: ...

### Creates
[string] and [array] type

`creates` parameter can have both string and array type e.g.

    backup-data:
      cmd.run:
      - name: echo 123
      - creates:
        - file1.txt
        - file2.txt

    #or
    backup-data:
      cmd.run:
      - name: echo 123
      - creates: file1.txt

The `creates` parameter contains the file path(s) which should be missing if you want the script to be executed. This is a quire useful option, when you want to protect a file from being overwritten. 
Imagine following scenario:

    backup-data:
      cmd.run:
      - name: echo 123
      - creates:
        - file1.txt
        - file2.txt

    #or
    backup-data:
      cmd.run:
      - name: echo 123
      - creates: file1.txt

