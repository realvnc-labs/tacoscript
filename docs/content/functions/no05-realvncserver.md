---
title: 'RealVNC Server'
weight: 5
slug: realvncserver
---

{{< toc >}}

## Preface

Tacoscript comes with functions to assist in managing your RealVNC servers. Initially you can update a
limited set of popular configuration parameters. In the future you'll be able to install / remove servers,
update an extended set of configuration parameters, etc.

Note that tacoscript uses snake case so `blank_screen` rather than the camel case (such as `BlankScreen`)
used by the RealVNC configuration files. Tacoscript automatically handles the translation.

## `realvnc_server.config_update`

The task `realvnc_server.config_update` allows you to update a limited set of your your RealVNC
configuration. Updating the RealVNC configuration on both linux/mac and windows machines is supported.

`realvnc_server.config_update` has following format:

```yaml
update-my-realvnc-server:
  realvnc_server.config_update:
    - encryption: AlwaysOn
    - blank_screen: true
```

We can interpret this script as:

1. Either add or update the RealVNC `Encryption` configuration parameter to be `AlwaysOn` and add or
   update the `BlankScreen` parameter to `true`. If the is a current setting then it will be replaced
   with the new value. If the setting is not currently set then it will added (either to the end of the
   configuration file for linux/mac or as a new registry setting for Windows).

{{< heading-supported-parameters >}}

### `config_file`

{{< parameter required=1 type=string >}}

When using with linux/mac then the path to the config file must be specified. When using Windows the
parameter is ignored.

If there is no existing config file then a new file will be created with the config parameters
added. The file will have regular permissions of `0644`.

### `server_mode`

{{< parameter required=0 type=string default="Service" >}}

The `server_mode` parameter tells tacoscript whether to target a `Service` or
`User` mode for the RealVNC server.

### `exec_path`

{{< parameter required=0 type=string default="(depends on platform and server mode)" >}}

The `exec_path` parameter can be used to override the default path for the RealVNC
server executable used to reload the config parameters. RealVNC defaults are used
depending on the platform (e.g. `linux` or `windows`) and the `server_mode` (specified
above). See the RealVNC docs for more information.

### `exec_cmd`

{{< parameter required=0 type=string default="(depends on platform and server mode)" >}}

The `exec_cmd` parameter can be used to override the default name for the RealVNC
server executable used to reload the config parameters. RealVNC defaults are used
depending on the platform (e.g. `linux` or `windows`) and the `server_mode` (specified
above).

### `skip_reload`

{{< parameter required=0 type=string default="false" >}}

if the `skip_reload` parameter is set then Tacoscript will NOT attempt to automatically
reload the config parameters after an update.

### `backup`

{{< parameter required=0 type=string default="bak" >}}

If using linux/mac then the `backup` parameter can be used to specify the backup
extension for the config file backup.

### `skip_backup`

{{< parameter required=0 type=string default="false" >}}

If set to `true` then a backup file will not be created. Linux/Mac only.

## Supported RealVNC Parameters

For details on each parameter, please see the links below.

### `encryption`

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Encryption)

### `authentication`

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Authentication)

### `permissions`

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Permissions)

### `query_connect`

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Query_Connect)

### `query_only_if_logged_on`

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Query_Only_If_Logged_On)

### `query_connect_timeout`

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Query_Connect_Timeout)

### `blank_screen`

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Blank_Screen)

### `conn_notify_timeout`

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Conn_Notify_Timeout)

### `conn_notify_always`

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Conn_Notify_Always)

### `idle_timeout`

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Idle_Timeout)

### `log`

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Log)

### `capture_method`

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Capture_Method)
