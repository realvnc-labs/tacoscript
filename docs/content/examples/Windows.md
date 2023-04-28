---
title: "Windows"
weight: 1
slug: windows
---
{{< toc >}}

## Run command with dependency

```yaml
pack-result:
  cmd.run:
    - names:
        - tar.exe -a -c -z -f C:\tmp\my-date.tar.gz C:\tmp\my-date.txt
        - move C:\tmp\my-date.tar.gz C:\tmp\mydate
    - require:
        - save-date
        - remove-date
        - create-folder
    - creates:
        - C:\tmp\my-date.tar.gz
    - shell: cmd.exe
```

## Run commands with conditions and dependency

```yaml
save-date:
  # Name of the class and the module
  cmd.run:
    - name: date.exe /T > C:\tmp\my-date.txt
    - cwd: C:\tmp
    #- user: breathbath ##Use it only as root
    - shell: cmd.exe
    - env:
        - PASSWORD: bunny
    - creates: C:\tmp\my-date.txt # Don't execute if file exists.
remove-date:
  cmd.run:
    - name: del C:\tmp\my-date.txt
    - shell: cmd.exe
    - require:
        - save-date
    - onlyif: date.exe /T | findstr -i "^Thu" # Execute only on Thursdays
```

## Download file from the internet

```yaml
create-folder:
  cmd.run:
    - names:
        - mkdir C:\tmp\mydate
    - unless: dir C:\tmp\mydate
    - shell: cmd.exe
another-file:
  file.managed:
    - name: my-file-win1251.txt
    - contents: |
        goes here
        Funny file
    - mode: 0755
    - encoding: windows1251
another-url:
  file.managed:
    - name: C:\tmp\sub\utf8-js-1.json
    - source: https://raw.githubusercontent.com/mathiasbynens/utf8.js/master/package.json
    - source_hash: sha256=40c5219fc82b478b1704a02d66c93cec2da90afa62dc18d7af06c6130d9966ed
    - skip_verify: false
    - makedirs: true
    - replace: false
    - user: Guest
    - group: Guest
    - mode: 0777
```

## Configure RealVNC VNC Server with 256-bit AES encryption

```yaml
realvnc-server-max-encryption:
  realvnc_server.config_update:
    - server_mode: Service
    - encryption: AlwaysMaximum
```

## Configure RealVNC VNC Server for attended access

```yaml
realvnc-server-attended-access:
  realvnc_server.config_update:
    - server_mode: Service
    - query_connect: true
    - query_only_if_logged_on: true
    - query_connect_timeout: 10
    - blank_screen: false
    - conn_notify_always: true
```

## Disable DirectX Capture in RealVNC VNC Server to troubleshoot display issues

```yaml
realvnc-server-display-fix:
  realvnc_server.config_update:
    - server_mode: Service
    - capture_method: 1
```

## Configure RealVNC VNC Server Access Control List

```yaml
# Determine <permissions_string> using RealVNC Permissions Creator
# https://help.realvnc.com/hc/en-us/articles/360002253618#using-vnc-permissions-creator-0-2

realvnc-server-display-fix:
  realvnc_server.config_update:
    - server_mode: Service
    - permissions: <permissions_string>
```

## Enable debug logging for RealVNC VNC Server

```yaml
realvnc-server-debug-logging:
  realvnc_server.config_update:
    - server_mode: Service
    - log: '*:EventLog:10,*:file:100'
```
