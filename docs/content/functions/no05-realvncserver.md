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
configuration. Updating the RealVNC configuration on both Linux/Mac and Windows machines is supported.

`realvnc_server.config_update` has following format:

```yaml
update-my-realvnc-server:
  realvnc_server.config_update:
    - encryption: AlwaysOn
    - blank_screen: true
    - idle_timeout: '!UNSET!'
```

We can interpret this script as:

1. Either add or update the RealVNC `Encryption` configuration parameter to be `AlwaysOn` and add or
   update the `BlankScreen` parameter to `true`. If the is a current setting then it will be replaced
   with the new value. If the setting is not currently set then it will added (either to the end of the
   configuration file for Linux/Mac or as a new registry setting for Windows).

2. Remove the `idle_timeout` parameter from the configuration. The RealVNC server will then apply the
   default value. Note that quotes must be used.

{{< heading-supported-parameters >}}

### `config_file`

{{< parameter type=string default="depends on platform and server mode">}}

When using with Linux/Mac then the path to the config file will default according to the `server_mode`
(see below). When using Windows the parameter is ignored.

If there is no existing config file then a new file will be created with the config parameters
added. The file will have regular permissions of `0644`.

### `server_mode`

{{< parameter required=0 type=string default="Service" >}}

The `server_mode` parameter will tell tacoscript whether to target a `Service`,
`User` or `Virtual` mode for the RealVNC server. `Virtual` mode will only be supported when running
Linux.

**Note: Initially tacoscript only supports the `Service` server mode. Future releases will support
`User` and `Virtual` modes.**

### `reload_exec_path`

{{< parameter required=0 type=string default="(depends on platform and server mode)" >}}

The `reload_exec_path` parameter can be used to override the default path and executable name
for the RealVNC server executable used to reload the config parameters. RealVNC defaults are
used depending on the platform (e.g. Linux, Mac or Windows) and the `server_mode` (specified
above). Note: This is for the executable only and does not accept command line arguments.

[See the RealVNC docs for more information.](https://help.realvnc.com/hc/en-us/articles/360002253878#reloading-parameters-0-6)

### `skip_reload`

{{< parameter required=0 type=bool default="false" >}}

if the `skip_reload` parameter is set then Tacoscript will NOT attempt to automatically
reload the config parameters after an update.

### `use_vnclicense_reload`

{{< parameter required=0 type=bool default="false" >}}

If set to `true` then will trigger a configuration parameters reload via the `vnclicense`
executable. Will be forced to `true` when using `Virtual` mode, otherwise defaults to `false`.
Can be used when the `server_mode` is `Service` or `User` but must be explicitly set to `true`.

### `backup`

{{< parameter required=0 type=string default="bak" >}}

If using Linux/Mac then the `backup` parameter can be used to specify the backup
extension for the config file backup.

### `skip_backup`

{{< parameter required=0 type=bool default="false" >}}

If set to `true` then a backup file will not be created. Linux/Mac only.

## Supported RealVNC Parameters

For details on each parameter, please see the links below.

### `encryption`

{{< parameter required=0 type=string >}}

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Encryption)

### `authentication`

{{< parameter required=0 type=string >}}

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Authentication)

### `permissions`

{{< parameter required=0 type=string >}}

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Permissions)

### `query_connect`

{{< parameter required=0 type=string >}}

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Query_Connect)

### `query_only_if_logged_on`

{{< parameter required=0 type=bool >}}

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Query_Only_If_Logged_On)

### `query_connect_timeout`

{{< parameter required=0 type=int >}}

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Query_Connect_Timeout)

### `blank_screen`

{{< parameter required=0 type=bool >}}

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Blank_Screen)

### `conn_notify_timeout`

{{< parameter required=0 type=int >}}

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Conn_Notify_Timeout)

### `conn_notify_always`

{{< parameter required=0 type=bool >}}

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Conn_Notify_Always)

### `idle_timeout`

{{< parameter required=0 type=int >}}

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Idle_Timeout)

### `log`

{{< parameter required=0 type=string >}}

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Log)

### `capture_method`

{{< parameter required=0 type=string >}}

[here](https://help.realvnc.com/hc/en-us/articles/360002251297-VNC-Server-Parameter-Reference#Capture_Method)
