---
title: 'Packages'
weight: 3
slug: packages
---

{{< toc >}}

## Preface

Tacoscript comes with functions to install software packages via package manager on you system.

## `pkg.installed`

The task `pkg.installed` ensures that the specified package(s) is installed in the managed host system.

`pkg.installed` has following format:

```yaml
install-neovim:
  pkg.installed:
    - name: neovim
    - version: 0.4.3-3
    - refresh: true
    - shell: bash
    - require:
        - create-file
    - unless: neovim --version
    - onlyif: date +%c|grep -q "^Thu"
```

We can interpret this script as:

The desired state of the script 'install-neovim' neovim package in version `0.4.3-3` to be installed in the target host
system. The `bash` shell will be used. Before running installation it will update list of available packages
(`refresh` = true). If this script is executed under the Ubuntu/Debian Linux OS, the tacoscript will execute this
command as:

```shell
apt update -y
apt install -y neovim=0.4.3-3
```

The script will be only executed after `create-file` script. So the tacoscript interpreter will make sure that the
`create-file` script is executed before `install-neovim` and only if it was successful.

The commands will be executed only under the following conditions:

- The package `neovim` is not installed yet, so `neovim --version` returns an error
- Today is Thursday

You can specify multiple packages to be installed by using `names` field:

```yaml
vim-family:
  pkg.installed:
    - refresh: true
    - names:
        - vim
        - neovim
        - vi
```

In this case `vim`, `neovim` and `vi` packages will be installed as e.g.

```shell
apt update
apt install -y vim neovim vi
```

When you use multiple packages, the version value will be applied to all packages. If version value is empty,
tacoscript will install the latest versions of all packages.

{{< heading-supported-parameters >}}

### `name`

{{< parameter required=1 type=string >}}

Name of the package to be installed.

_Either `name` or `names` is required._

### `names`

{{< parameter required=1 type=array >}}

Name contains the list of packages to be installed.

_Either `name` or `names` is required._

### `shell`

{{< parameter required=0 type=string >}}

Shell is a program that takes commands from input and gives them to the operating system to perform.
Known Linux shells are [bash](https://www.gnu.org/software/bash/), [sh](https://www.gnu.org/software/bash/),
[zsh](https://ohmyz.sh/) etc. Windows supports [cmd.exe](https://ss64.com/nt/cmd.html) shell.

### `version`

{{< parameter required=0 type=string >}}

Version of the package to be installed. If omitted, the tacoscript will install the latest default version.

### `refresh`

{{< parameter required=0 type=boolean >}}

If true, the tacoscript will update list of available packages, e.g. execute `apt update` under Ubuntu/Debian OS.

### OS Support

| OS      | OS Platform   | Package manager           | Installation script to be executed, e.g. vim |
| ------- | ------------- | ------------------------- | -------------------------------------------- |
| macOS   | Darwin        | brew                      | `brew install vim`                           |
| Linux   | Ubuntu/Debian | apt (fallback to apt-get) | `apt install -y vim`                         |
| Linux   | CentOS/Redhat | dnf (fallback to yum)     | `dnf install -y vim`                         |
| Windows | Windows       | choco                     | `choco install -y vim`                       |

Note if a corresponding package manager is not installed on the host system, a fallback will be used.
If both are not available, the script will fail.

## `pkg.uptodate`

The task `pkg.uptodate` ensures that the specified package(s) is upgraded to the latest version.

`pkg.uptodate` has following format:

```yaml
upgrade-neovim:
  pkg.uptodate:
    - name: neovim
```

This script will upgrade `neovim` package to the latest stable version.
If this script is executed under the Ubuntu/Debian Linux OS, the tacoscript will execute this command as:

```shell
apt upgrade -y neovim
```

You can specify multiple packages to be installed by using `names` field:

```yaml
vim-family-latest:
  pkg.uptodate:
    - refresh: true
    - names:
        - vim
        - neovim
        - vi
```

In this case `vim`, `neovim` and `vi` packages will be upgraded as e.g.

```shell
apt update
apt upgrade -y vim neovim vi
```

### `name`

{{< parameter required=1 type=string >}}

Name of the package to be installed.

_Either `name` or `names` is required._

### `names`

{{< parameter required=1 type=array >}}

Name contains the list of packages to be installed.

_Either `name` or `names` is required._

### `shell`

{{< parameter required=0 type=string >}}

See #pkg.installed for reverence.

### `refresh`

{{< parameter required=1 type=boolean >}}

See #pkg.installed for reverence.

### OS Support

| OS      | OS Platform   | Package manager           | Installation script to be executed, e.g. vim |
| ------- | ------------- | ------------------------- | -------------------------------------------- |
| macOS   | Darwin        | brew                      | `brew upgrade vim`                           |
| Linux   | Ubuntu/Debian | apt (fallback to apt-get) | `apt upgrade -y vim`                         |
| Linux   | CentOS/Redhat | dnf (fallback to yum)     | `dnf upgrade -y vim`                         |
| Windows | Windows       | choco                     | `choco upgrade -y vim`                       |

Note if a corresponding package manager is not installed on the host system, a fallback will be used.
If both are not available, the script will fail. If the package is not yet installed, the behaviour will vary
depending on the package manager e.g. `brew` will install this package and `apt-get` will fail.

## `pkg.removed`

The task `pkg.removed` ensures that the specified package(s) is removed from the host system.

`pkg.removed` has following format:

```yaml
delete-neovim:
  pkg.removed:
    - name: neovim
```

This script will delete `neovim` package.
If this script is executed under the Ubuntu/Debian Linux OS, the tacoscript will execute this command as:

```shell
apt remove -y neovim
```

You can specify multiple packages to be removed by using `names` field:

```yaml
vim-family-removed:
  pkg.removed:
    - names:
        - vim
        - neovim
        - vi
```

In this case `vim`, `neovim` and `vi` packages will be removed as e.g.

```shell
apt update
apt remove -y vim neovim vi
```

### `name`

{{< parameter required=1 type=string >}}

Name of the package to be installed.

_Either `name` or `names` is required._

### `names`

{{< parameter required=1 type=array >}}

Name contains the list of packages to be installed.

_Either `name` or `names` is required._

### `shell`

{{< parameter required=0 type=string >}}

### `refresh`

{{< parameter required=1 type=boolean >}}

See #pkg.installed for reverence.

### OS Support

| OS      | OS Platform   | Package manager           | Installation script to be executed, e.g. vim |
| ------- | ------------- | ------------------------- | -------------------------------------------- |
| macOS   | Darwin        | brew                      | `brew uninstall vim`                         |
| Linux   | Ubuntu/Debian | apt (fallback to apt-get) | `apt remove -y vim`                          |
| Linux   | CentOS/Redhat | dnf (fallback to yum)     | `dnf remove -y vim`                          |
| Windows | Windows       | choco                     | `choco uninstall -y vim`                     |

Note if a corresponding package manager is not installed on the host system, a fallback be used.
If both are not available, the script will fail.
