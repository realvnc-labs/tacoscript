# pkg.installed

The task `pkg.installed` ensures that the specified package(s) is installed in the managed host system. 

`pkg.installed` has following format:

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
        - manager: apt

We can interpret this script as:

The desired state of the script 'install-neovim' neovim package in version `0.4.3-3` to be installed in the target host system. As shell it will use `bash`.Before running installation it will update list of available packages (`refresh` = true). This script will be executed with the `apt` package manager as:
    
    - apt update -y
    - apt install -y neovim=0.4.3-3

The script will be only executed after `create-file` script. So the tacoscript interpreter will make sure that the `create-file` script is executed before `install-neovim` and only if it was successful.

The commands will be executed only under the following conditions:
- The package `neovim` is not installed yet, so `neovim --version` returns an error
- Today is Thursday

You can specify multiple packages to be installed by using `names` field:

    vim-family:
      pkg.installed:
        - refresh: true
        - names:
            - vim
            - neovim
            - vi

In this case `vim`, `neovim` and `vi` packages will be installed as e.g.

    apt update
    apt install -y vim neovim vi
    
When you use multiple packages, the version value will be applied to all packages. If version value is empty, the tacoscript will install the latest versions of all packages.

## Task parameters

### name

[string] type, required

Name of the package to be installed. 
        
### names
[array] type

Name contains the list of packages to be installed. 

### shell
[string] type

Shell is a program that takes commands from input and gives them to the operating system to perform. Known Linux shells are [bash](https://www.gnu.org/software/bash/), [sh](https://www.gnu.org/software/bash/), [zsh](https://ohmyz.sh/) etc. Windows supports [cmd.exe](https://ss64.com/nt/cmd.html) shell.

### require
see [require](../../general/dependencies/require.md)

### onlyif
see [onlyif](../../general/conditionals/onlyif.md)

### unless
see [onlyif](../../general/conditionals/unless.md)

### version

[string] type, optional

Version of the package to be installed. If ommitted, the tacoscript will install the latest default version.

### manager

[string] type, optional

Package manager to be used for dependency management. Currently, tacoscript supports following package managers:
#### Linux
- apt
- apt-get
- yum
- dnf
#### Windows
- choco
- winget
#### MacOS
- brew

Please note, you should be careful when specifying a package manager explicitly, e.g. `apt` cannot be executed under Fedora, so it will fail there. But if you don't specify it explicitly, tacoscript will execute `yum`, since it detects it to be available there.

### refresh

[bool] type, optional

If true, the tacoscript will update list of available packages, e.g. execute `apt update` under Ubuntu/Debian OS.

## OS Support
<table>
<tr>
<th>OS</th>
<th>OS Platform</th>
<th>Package manager</th>
<th>Installation script to be executed, e.g. vim</th>
</tr>
<tr>
<td>MacOS</td>
<td>Darwin</td>
<td>brew</td>
<td>brew install vim</td>
</tr>
<tr>
<td>Linux</td>
<td>Ubuntu/Debian</td>
<td>apt (fallback to apt-get)</td>
<td>apt install -y vim</td>
</tr>
<tr>
<td>Linux</td>
<td>CentOS/Redhat</td>
<td>dfm (fallback to yum)</td>
<td>dfm install -y vim</td>
</tr>
<tr>
<td>Windows</td>
<td></td>
<td>choco / winget (fallback to choco)</td>
<td>choco install -y vim / winget install -e -h --accept-package-agreements --accept-source-agreements vim</td>
</tr>
</table>

Note if a corresponding package manager is not installed in the host system, a fallback one will be used. If both are not available, the script will fail. If user explicitly specified a package manager (e.g. winget), tacoscript will try to use it by all means, no matter if it's installed or not.

# pkg.uptodate

The task `pkg.uptodate` ensures that the specified package(s) is upgraded to the latest version. 

`pkg.uptodate` has following format:

    upgrade-neovim:
      pkg.uptodate:
        - name: neovim

This script will upgrade `neovim` package to the latest stable version.
If this script is executed under the Ubuntu/Debian Linux OS, the tacoscript will execute this command as:
    
    - apt upgrade -y neovim

You can specify multiple packages to be installed by using `names` field:

    vim-family-latest:
      pkg.uptodate:
        - refresh: true
        - names:
            - vim
            - neovim
            - vi

In this case `vim`, `neovim` and `vi` packages will be upgraded as e.g.

    apt update
    apt upgrade -y vim neovim vi.

## Task parameters

### name

[string] type, required

Name of the package to be upgraded. 
        
### names
[array] type

Name contains the list of packages to be upgraded. 

### shell
[string] type

See #pkg.installed for reverence.

### manager

[string] type, optional
See #pkg.installed for reverence.

### require
see [require](../../general/dependencies/require.md)

### onlyif
see [onlyif](../../general/conditionals/onlyif.md)

### unless
see [onlyif](../../general/conditionals/unless.md)

### refresh
[bool] type, optional
See #pkg.installed for reverence.

## OS Support
<table>
<tr>
<th>OS</th>
<th>OS Platform</th>
<th>Package manager</th>
<th>Installation script to be executed, e.g. vim</th>
</tr>
<tr>
<td>MacOS</td>
<td>Darwin</td>
<td>brew</td>
<td>brew upgrade vim</td>
</tr>
<tr>
<td>Linux</td>
<td>Ubuntu/Debian</td>
<td>apt (fallback to apt-get)</td>
<td>apt upgrade -y vim</td>
</tr>
<tr>
<td>Linux</td>
<td>CentOS/Redhat</td>
<td>dfm (fallback to yum)</td>
<td>dfm upgrade -y vim</td>
</tr>
<tr>
<td>Windows</td>
<td></td>
<td>choco / winget</td>
<td>choco upgrade -y vim / winget upgrade -e -h --accept-package-agreements --accept-source-agreements vim</td>
</tr>
</table>

Note if a corresponding package manager is not installed in the host system, a fallback one will be used. If both are not available, the script will fail. If user explicitly specified a package manager (e.g. winget), tacoscript will try to use it by all means, no matter if it's installed or not.

# pkg.removed

The task `pkg.removed` ensures that the specified package(s) is removed from the host system. 

`pkg.removed` has following format:

    delete-neovim:
      pkg.removed:
        - name: neovim

This script will delete `neovim` package.
If this script is executed under the Ubuntu/Debian Linux OS, the tacoscript will execute this command as:
    
    - apt remove -y neovim

You can specify multiple packages to be removed by using `names` field:

    vim-family-removed:
      pkg.removed:
        - names:
            - vim
            - neovim
            - vi

In this case `vim`, `neovim` and `vi` packages will be removed as e.g.

    apt update
    apt remove -y vim neovim vi.

## Task parameters

### name

[string] type, required

Name of the package to be removed. 
        
### names
[array] type

Name contains the list of packages to be removed. 

### shell
[string] type

See #pkg.installed for reverence.

### manager
[string] type, optional

See #pkg.installed for reverence.

### require
see [require](../../general/dependencies/require.md)

### onlyif
see [onlyif](../../general/conditionals/onlyif.md)

### unless
see [onlyif](../../general/conditionals/unless.md)

### refresh
[bool] type, optional

See #pkg.installed for reverence.

## OS Support
<table>
<tr>
<th>OS</th>
<th>OS Platform</th>
<th>Package manager</th>
<th>Installation script to be executed, e.g. vim</th>
</tr>
<tr>
<td>MacOS</td>
<td>Darwin</td>
<td>brew</td>
<td>brew uninstall vim</td>
</tr>
<tr>
<td>Linux</td>
<td>Ubuntu/Debian</td>
<td>apt (fallback to apt-get)</td>
<td>apt remove -y vim</td>
</tr>
<tr>
<td>Linux</td>
<td>CentOS/Redhat</td>
<td>dfm (fallback to yum)</td>
<td>dfm remove -y vim</td>
</tr>
<tr>
<td>Windows</td>
<td></td>
<td>choco / winget</td>
<td>choco uninstall -y vim / winget uninstall -e -h --accept-source-agreements vim</td>
</tr>
</table>

Note if a corresponding package manager is not installed in the host system, a fallback one will be used. If both are not available, the script will fail. If user explicitly specified a package manager (e.g. winget), tacoscript will try to use it by all means, no matter if it's installed or not.
