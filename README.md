[![Coverage Status](https://coveralls.io/repos/github/cloudradar-monitoring/tacoscript/badge.svg)](https://coveralls.io/github/cloudradar-monitoring/tacoscript)
[![Actions Status](https://github.com/cloudradar-monitoring/tacoscript/workflows/Go/badge.svg)](https://github.com/cloudradar-monitoring/tacoscript/actions)
## Overview
Tacoscript library provides functionality for provisioning of remote servers and local machines running on any OS. Tacoscript can be installed as a binary therefore it doesn't require any additional tools or programs on the host system. 

Tacoscript provisions host systems from simple yaml files written in a [SaltStack](https://www.saltstack.com/) driven configuration language. 

Why we need another provisioning tool? Unfortunately the next competitors like Puppet, Ansible or Salt have a limited support for Windows or require installation of additional tools which is not always convenient.

Tacoscript is written in GO, so its provided as an executable binary for a big variety of host OS and platforms and it requires no additional tools to be installed in the host system. You can find the list of supported OS and architectures [here](https://golang.org/doc/install/source#environment). 

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
    
    #The output will be :
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

_cmd.run_
Executes an arbitrary command in a shell of a host system. The `cmd.run` task can have following parameters:


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

You can interpret this script as following:

The desired state of the script 'create-file' is a file at /tmp/data2001-01-01.txt. To achieve it, the tacoscript binary should execute a task of type 'cmd.run'. This task type executes command `echo 'data to backup' >> /tmp/data2001-01-01.txt` in the shell of the host system.
The second desired state of the script 'backup-data' is creation of 2 files: `/dumps/data2001-01-01.txt.tar.gz` and it's md5 sum file at `/dumps/data2001-01-01.txt.tar.gz.md5`.

It should be achieved by executing a task of type 'cmd.run' which requires 2 shell commands:

- `tar czf /dumps/data2001-01-01.txt.tar.gz /tmp/data2001-01-01.txt`
- `md5sum /dumps/data2001-01-01.txt.tar.gz >> /dumps/data2001-01-01.txt.tar.gz.md5`

The commands will be executed in the following context:
- current working directory (cwd) will be `/tmp`
- env variables list will contain PASSWORD with value `bunny`
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

The `creates` parameter identifies the files which should be missing if you want to run the current task. In other words if any of the files in the `creates` section exist, the task will never run. 

A typical example use case for this parameter would be an exclusive access to a file, e.g. we don't run the backup, if the file is locked by another process:

    backup-data:
      cmd.run:
      - names:
        - touch serviceALock.txt
        - tar cvf somedata.txt.tar somedata.txt
        - rm serviceALock.txt
      - creates: serviceALock.txt
      
In this situation we expect that the script is running periodically, so at some point the lock will be removed, and the backup-data script get a chance to make a backup.


### shell
[string] type

Shell is a program that takes commands from input and gives them to the operating system to perform. Known Linux shells are [bash](https://www.gnu.org/software/bash/), [sh](https://www.gnu.org/software/bash/), [zsh](https://ohmyz.sh/) etc. Windows supports [cmd.exe](https://ss64.com/nt/cmd.html) shell. 

If you don't specify this parameter, tacoscript will use the default golang [exec function](https://golang.org/pkg/os/exec/)
which intentionally does not invoke the system shell and does not expand any glob patterns or handle other expansions, pipelines, or redirections typically done by shells.
 
 To expand glob patterns, you can specify the `shell` parameter, in this case you should take care to escape any dangerous input.
 
 **Note that if you use `cmd.run` task type without the `shell` parameter, usual patterns like pipelines and redirections won't work.** 
 
 If you specify a `shell` parameter, tacoscript will run your task commands as a '-c' parameter under Unix and '/C' parameter under Windows:
 
 The script below:
 
     backup-data:
       cmd.run:
       - names:
         - touch serviceALock.txt
         - tar cvf somedata.txt.tar somedata.txt
         - rm serviceALock.txt
       - shell: bash
       
will run as:

    bash -c touch serviceALock.txt
    bash -c tar cvf somedata.txt.tar somedata.txt
    bash -c rm serviceALock.txt
    
The script below:
 
     backup-data:
       cmd.run:
       - name: date.exe /T > C:\tmp\my-date.txt
       - shell: cmd.exe
       
will run as:

    cmd.exe \C date.exe /T > C:\tmp\my-date.txt

### user
[string] type

    create-user-file:
      cmd.run:
        - user: www-data
        - touch: data.txt

The `user` parameter allows to run commands as a specific user. In Linux systems this will require sudo rights for the tacoscript binary. In Windows this command will be ignored.
  Switching users allows to create resources (file, services, folders etc) with the ownership of the specified user. 

After running the above script, tacoscript will create a `data.txt` file with the ownership of `www-data` user.

    sudo tacoscript tacoscript.yaml
    ls -la
    #output will be
    -rw-r--r--    4 root  root          128 Jul 15  2019 tacoscript.yaml
    -rw-r--r--    1 root  root          3550 Oct 30  2019 tacoscript
    -rw-r--r--@   1 www-data  www-data  0 Apr 23 09:22 data.txt
    
    
### env
[keyValue] type 

The `env` parameter is a list of key value parameters where key represents the name of an environment variable and value it's content. Env variables are parameters which are set from the outside of a running program and can be used as configuration data.

    save-date:
      cmd.run:
        - name: psql
        - env:
            - PGUSER: bunny
            - PGPASSWORD: bug

In this example the psql will read login and password from the corresponding env variables and connect to the database without any input parameters or configuration data.

### onlyif
[string] and [array] type

`onlyif` parameter can be both string and array e.g.

    publish-kafka-message:
      cmd.run:
        - name: kafka-console-producer.sh --broker-list localhost:9092 --topic my_topic --new-producer < my_file.txt
        - onlyif: test -e my_file

    #or
    
    publish-kafka-message:
      cmd.run:
        - name: kafka-console-producer.sh --broker-list localhost:9092 --topic my_topic --new-producer < messages.txt
        - onlyif: 
            - service kafka status
            - service zookeeper status
            - test -e messages.txt

The `onlyif` parameter gives a list of commands which will be executed before the actual task and if any of them fails (returns non zero exit code), the task will be skipped. However, the failures won't stop the program execution. The `onlyif` checks are given to prove that the task should run, in case of failure the task will be completely ignored.

In the example above the command `kafka-console-producer.sh --broker-list localhost:9092 --topic my_topic --new-producer < my_file.txt` won't be triggered, if tacoscript detects, that kafka or zookeeper services are not running or `messages.txt` file doesn't exist.

### unless
[string] and [array] type

`unless` parameter can be both string and array e.g.

    start-myservice:
      cmd.run:
        - name: service myservice start
        - unless: test -e myservice_fatallogs.txt

    #or
    
    report-failure:
      cmd.run:
        - name: mail -s "Some configs are missing, check what is going on" admin@admin.com
        - unless:
            - test -e criticalConfigOne.txt
            - test -e criticalConfigTwo.txt
            - test -e criticalConfigThree.txt

The `unless` is a reverted `onlyif` parameter which means the task will run only of the `unless` condition failes (returns a non zero exit code). However if multiple `unless` parameters are used, tacoscript will execute the task only if **one** `unless` condition fails (the `onlyif` parameters should be all successful to unblock the task). 

The above examples show some use cases for this parameter. In the first one we execute the command `service myservice start` only if we detect that there are no critical logs. A typical use case is to stop an endless startup loop for a service when we detect some critical logs from it.

The second example demonstrates a check for some required configs. If tacoscript detects any missing config from the list, it will send an email to the server's adminstrator. 

### Known limitations
- to use shell pipes, redirects or glob expands, please specify a `shell` parameter
- `user` parameter will require sudo rights for tacoscript, in Windows this parameter is ignored
- if you use cmd.run tasks in Windows, you'd better specify the shell parameter as `cmd.exe`, otherwise you will get errors like:
    `exec: "xxx": executable file not found in %PATH%`

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
