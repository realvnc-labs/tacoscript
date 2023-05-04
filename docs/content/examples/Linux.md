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

## Enable debug logging for RealVNC VNC Server

```yaml
realvnc-server-debug-logging:
  realvnc_server.config_update:
    - server_mode: Service
    - log: '*:syslog:10,*:file:100'
```

## Download, install and license RealVNC VNC Server

```yaml
# Installs/updates VNC Server and either licenses it offline with license key or joins to the cloud with a cloud connectivity token

# CONFIGURE PARAMETERS BELOW TO YOUR REQUIREMENTS

# Set version of VNC Server to install, e.g. 7.1.0
# Defaults to Latest
{{$Version := "Latest"}}

# Set if we're using cloud or offline
# Accepted values: offline, cloud
{{$CloudOrOffline := "cloud"}}

# Set offline license to apply
# Not required if joining to the cloud
{{$OfflineLicense := ""}}

# Set cloud connectivity token to join VNC Server to the cloud
# Not required if using direct connections only
{{$CloudToken := ""}}

# Set group to join VNC Server to - group must exist in the VNC Connect portal
# Optional
{{$CloudGroup := ""}}


# DO NOT EDIT BELOW THIS LINE
{{$LinuxScriptPath := "/tmp/vnc.sh"}}

template:
  file.managed:
    - name: {{ $LinuxScriptPath }}
    - contents: |
        #!/bin/sh
        
        Version="{{ $Version }}"
        CloudOrOffline="{{ $CloudOrOffline }}"
        OfflineLicense="{{ $OfflineLicense }}"
        CloudToken="{{ $CloudToken }}"
        CloudGroup="{{ $CloudGroup }}"
        
        # Use TEMP for downloaded files
        TempPath="/var/tmp"
        
        # Detect architecture
        Architecture="x64"
        DetectedArchitecture="$(lscpu | grep Architecture | awk '{print $2}')"
        if [ "$DetectedArchitecture" = "i386" ] || [ "$DetectedArchitecture" = "i686" ]; then
          Architecture="x86"
        elif [ "$DetectedArchitecture" = "aarch64" ]; then
          Architecture="ARM64"
        elif echo "$DetectedArchitecture" | grep -q "arm"; then
          Architecture="ARM"
        fi
        
        # Set file extension based on package manager
        FileExt=""
        if type dpkg > /dev/null 2>&1; then
          FileExt=".deb"
        elif type rpm > /dev/null 2>&1; then
          FileExt=".rpm"
        fi
        
        # Download VNC Server package from RealVNC website
        curl -fsL --retry 3 "https://downloads.realvnc.com/download/file/vnc.files/VNC-Server-${Version}-Linux-${Architecture}${FileExt}" -o "${TempPath}/VNC${FileExt}"
        
        # Install VNC Server package
        if [ "$FileExt" = ".deb" ]; then
          apt install -y "${TempPath}/VNC${FileExt}"
        elif [ "$FileExt" = ".rpm" ]; then
          yum install -y "${TempPath}/VNC${FileExt}"
        fi
        
        # Cleanup package
        rm -f "${TempPath}/VNC${FileExt}"
        
        # Determine if we are licensing by key or cloud joining
        if [ "$CloudOrOffline" = "offline" ]; then
          # Call vnclicense.exe to apply the key
          /usr/bin/vnclicense -add "$OfflineLicense"
        elif [ "$CloudOrOffline" = "cloud" ]; then
          if [ "$(vncserver-x11 -service -cloudstatus | grep CloudJoined | cut -f2 -d':' | sed 's/,//')" = "false" ]; then 
            # If CloudGroup is set, use it to join VNC Server to that group - group must exist in the VNC Connect portal
            if [ -n "$CloudGroup" ]; then
              joinGroup="-joinGroup $CloudGroup"
              # Call vncserver.exe to do the cloud join
              /usr/bin/vncserver-x11 -service -joinCloud "$CloudToken" "$joinGroup"
            else
              # Call vncserver.exe to do the cloud join
              /usr/bin/vncserver-x11 -service -joinCloud "$CloudToken"
            fi
          fi
        fi
        
        # Add firewall rules if firewalld or ufw are detected
        if type firewall-cmd > /dev/null 2>&1; then
          firewall-cmd --zone=public --permanent --add-service=vncserver-x11-serviced
          firewall-cmd --reload
        elif type ufw > /dev/null 2>&1; then
          ufw allow 5900
        fi
        
        # Disable Wayland is detected
        gdmconf=""
        if [ -f "/etc/gdm3/custom.conf" ]; then gdmconf="/etc/gdm3/custom.conf"; fi
        if [ -f "/etc/gdm/custom.conf" ]; then gdmconf="/etc/gdm/custom.conf"; fi
        if [ -n "$gdmconf" ]; then
          if [ "$(grep -c "^#.*WaylandEnable=false" "$gdmconf")" -gt 0 ]; then
            cp -a "$gdmconf" "$gdmconf.bak"
            sed -i 's/^#.*WaylandEnable=.*/WaylandEnable=false/' "$gdmconf"
            systemctl restart gdm*
          fi
        fi
        
        systemctl enable vncserver-x11-serviced --now
        systemctl restart vncserver-x11-serviced
    - skip_verify: true
    - mode: 0755
    - encoding: UTF-8
  cmd.run:
    - names:
      - sh -x {{ $LinuxScriptPath }}
      - rm -f {{ $LinuxScriptPath }}
    - shell: sh
