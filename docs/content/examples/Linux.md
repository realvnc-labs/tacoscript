---
title: "Linux"
weight: 1
slug: linux
---
{{< toc >}}

## Execute a command

```yaml
my command:
  cmd.run:
    - name: date
```

## Execute multiple command using a block

```yaml
a block of commands:
  cmd.run:
    - name: |
        echo "first line"
        echo "second line"
```

## A single command spanned over multiple lines

```yaml
spanned command:
  cmd.run:
    - name: >-
        echo
        this all goes
        on a single line
```

## Execute multiple commands using a list

```yaml
a pile of commands:
  cmd.run:
    - names:
        - echo "first line"
        - echo "second line"
        - free
        - uptime
```

## Execute a command in a specific directory

```yaml
go to a dir first:
  cmd.run:
    - name: ls -l|wc -l
    - cwd: /etc
```

## Redirect stdout to a file

```yaml
redirect output:
  cmd.run:
    - names:
        - echo 'data to backup' >> /tmp/data2001-01-01.txt
        - date>/tmp/i-was-here.txt
```

## Skip if file exists

```yaml
skip on file:
  cmd.run:
    - name: echo SKIP
    - creates:
        - /etc/hosts
```

## run on previously failing command

```yaml
require failing command:
  cmd.run:
    - name: date
    - unless: echo 1|grep -q 0
```

## change the shell

```yaml
zchange shell:
  cmd.run:
    - name: print("hallo python")
    - shell: /usr/local/bin/python3
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

## Configure RealVNC VNC Server Access Control List

```yaml
# Determine <permissions_string> using RealVNC Permissions Creator
# https://help.realvnc.com/hc/en-us/articles/360002253618#using-vnc-permissions-creator-0-2

realvnc-server-display-fix:
  realvnc_server.config_update:
    - server_mode: Service
    - permissions: <permissions_string>
```