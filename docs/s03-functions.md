# Functions

## cmd

### cmd.run

The task `cmd.run` executes an arbitrary command in a shell of a host system. It has following syntax:

```yaml
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
```

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

#### Task parameters

##### name

[string] type

```yaml
create-file:
  cmd.run:
    - name: echo 'data to backup' >> /tmp/data2001-01-01.txt
```

Name describes a single executable command. In the example above the tacoscript interpreter will run `echo 'data to backup' >> /tmp/data2001-01-01.txt` command.

##### names
[array] type

```yaml
backup-data:
  cmd.run:
    - names:
        - tar czf /dumps/data2001-01-01.txt.tar.gz /tmp/data2001-01-01.txt
        - md5sum /dumps/data2001-01-01.txt.tar.gz >> /dumps/data2001-01-01.txt.tar.gz.md5
```

Name contains the list of commands. All commands in the task will be executed in the order of appearance. If one fails, the whole execution will stop. All commands inside one task will be executed in the same context, which means in the current working directory, with same env variables, in the same shell and under same conditions.

If you want to change context of a command (e.g. use another shell), you should create another task e.g.

```yaml
backup-data:
  cmd.run:
    - name: mycmd in shell 'bash'
    - shell: bash
  cmd.run:
    - name: mycmd in shell 'sh'
    - shell: sh
```

The `names` parameter with a single value has the same meaning as `name` field.

##### cwd
[string] type

```yaml
backup-data:
  cmd.run:
    - cwd: /tmp
```

The `cwd` parameter gives the current working directory of a command. This value is quite useful if you want to use relative paths of files that you provide to your commands.

For example imagine following file structure:

```
C:\Some\Very\Long\Path
  someData1.txt
  someData2.txt
```

You want pack someData1.txt and someData2.txt with the zip.exe binary. You can do it with the script as:

```yaml
backup-data:
  cmd.run:
    - names:
        - tar.exe -a -c -z -f C:\Some\Very\Long\Path\someData1.txt.tar.gz C:\Some\Very\Long\Path\someData1.txt
        - tar.exe -a -c -z -f C:\Some\Very\Long\Path\someData2.txt.tar.gz C:\Some\Very\Long\Path\someData2.txt
```

Obviously we have quite long strings here.
You can also change your working directory and use relative paths to get the same result:

```yaml
backup-data:
  cmd.run:
    - cwd: C:\Some\Very\Long\Path\
    - names:
        - tar.exe -a -c -z -f someData1.txt.tar.gz someData1.txt
        - tar.exe -a -c -z -f someData2.txt.tar.gz someData2.txt
```

##### shell
[string] type

Shell is a program that takes commands from input and gives them to the operating system to perform. Known Linux shells are [bash](https://www.gnu.org/software/bash/), [sh](https://www.gnu.org/software/bash/), [zsh](https://ohmyz.sh/) etc. Windows supports [cmd.exe](https://ss64.com/nt/cmd.html) shell.

If you don't specify this parameter, tacoscript will use the default golang [exec function](https://golang.org/pkg/os/exec/)
which intentionally does not invoke the system shell and does not expand any glob patterns or handle other expansions, pipelines, or redirections typically done by shells.

 To expand glob patterns, you can specify the `shell` parameter, in this case you should take care to escape any dangerous input.

 **Note that if you use `cmd.run` task type without the `shell` parameter, usual patterns like pipelines and redirections won't work.**

 If you specify a `shell` parameter, tacoscript will run your task commands as a '-c' parameter under Unix and '/C' parameter under Windows:

 The script below:

```yaml
backup-data:
  cmd.run:
  - names:
    - touch serviceALock.txt
    - tar cvf somedata.txt.tar somedata.txt
    - rm serviceALock.txt
  - shell: bash
```

will run as:

```bash
bash -c touch serviceALock.txt
bash -c tar cvf somedata.txt.tar somedata.txt
bash -c rm serviceALock.txt
```

The script below:

```yaml
  backup-data:
    cmd.run:
      - name: date.exe /T > C:\tmp\my-date.txt
      - shell: cmd.exe
```

will run as:

```bash
cmd.exe \C date.exe /T > C:\tmp\my-date.txt
```

##### user
[string] type

```yaml
create-user-file:
  cmd.run:
    - user: www-data
    - touch: data.txt
```

The `user` parameter allows to run commands as a specific user. In Linux systems this will require sudo rights for the tacoscript binary. In Windows this command will be ignored.
  Switching users allows to create resources (file, services, folders etc) with the ownership of the specified user.

After running the above script, tacoscript will create a `data.txt` file with the ownership of `www-data` user.

```bash
sudo tacoscript tacoscript.yaml
ls -la
#output will be
-rw-r--r--    4 root  root          128 Jul 15  2019 tacoscript.yaml
-rw-r--r--    1 root  root          3550 Oct 30  2019 tacoscript
-rw-r--r--@   1 www-data  www-data  0 Apr 23 09:22 data.txt
```

##### env
[keyValue] type

The `env` parameter is a list of key value parameters where key represents the name of an environment variable and value it's content. Env variables are parameters which are set from the outside of a running program and can be used as configuration data.

```yaml
save-date:
  cmd.run:
    - name: psql
    - env:
        - PGUSER: bunny
        - PGPASSWORD: bug
```

In this example the psql will read login and password from the corresponding env variables and connect to the database without any input parameters or configuration data.

##### require
see [require](/docs#require)

##### creates
see [creates](./s02-conditionals.md#creates)

##### onlyif
see [onlyif](./s02-conditionals.md#onlyif)

##### unless
see [unless](./s02-conditionals.md#unless)

## file

### file.managed

The task `file.managed` ensures the existence of a file in the local file system. It can download files from remote urls (currently http(s)/ftp protocols are supported) or copy a file from the local file system. It can verify the checksums processed files and show content diffs in the source and target files.

`file.managed` has following format:

```yaml
maintain-my-file:
  file.managed:
    - name: C:\temp\progs\npp.7.8.8.Installer.x64.exe
    - source: https://github.com/notepad-plus-plus/notepad-plus-plus/releases/download/v7.8.8/npp.7.8.8.Installer.x64.exe
    - source_hash: md5=79eef25f9b0b2c642c62b7f737d4f53f
    - makedirs: true # default false
    - replace: false # default true
    - creates: 'C:\Program Files\notepad++\notepad++.exe'
```

We can interpret this script as:
Download a file from `https://github.com/notepad-plus-plus/notepad-plus-plus/releases/download/v7.8.8/npp.7.8.8.Installer.x64.exe` to some temp location.
Check if md5 hash of it matches `79eef25f9b0b2c642c62b7f737d4f53f`. If not, skip the task. Check if md5 hash of the target file `C:\temp\npp.7.8.8.Installer.x64.exe` matches `79eef25f9b0b2c642c62b7f737d4f53f`, if yes, it means the file exists and has the desired content, so the task will be skipped.
The tacoscript should make directory tree `C:\temp\progs`, if needed. If file at `C:\temp\npp.7.8.8.Installer.x64.exe` exists, it won't be replaced even if it has a different content. The task will be skipped if the file `C:\Program Files\notepad++\notepad++.exe` already exists.

Here is another `file.managed` task:

```yaml
another-file:
  file.managed:
    - name: /tmp/my-file.txt
    - contents: |
        My file content
        goes here
        Funny file
    - skip_verify: true # default false
    - user: root
    - group: www-data
    - mode: 0755
    - encoding: UTF-8
    - onlyif:
      - which apache2
```

We can read it as following:
Copy contents `My file content\ngoes here\nFunny file` to the `/tmp/my-file.txt` file. Don't check the hashsum of it. Implicitly tacoscript will compare the contents of the target file with the provided content and show the differences. If the contents don't differ, the task will be skipped. If file doesn't exist, it will be created. Tacoscript will make sure, that the file `/tmp/my-file.txt` is owned by user `root`, `group` - `www-data`, and has file mode 0755. The target content will be encoded as `UTF-8`. The task will be skipped if the target system has no `apache2` installed.

##### Task parameters

##### name

[string] type, required

Name is the file path of the target file. A `file.managed` will make sure that the file `/tmp/targetfile.txt` is created or has the expected content.

```yaml
create-file:
  file.managed:
    - name: /tmp/targetfile.txt
```

##### source, default empty string

URL or local path of the source file which should be copied to the target file. Source can be HTTP, HTTPS or FTP URL or a local file path. See some examples below:

```yaml
create-file:
  file.managed:
    - name: /tmp/targetfile.txt
    - source: ftp://user:pass@11.22.33.44:3101/file.txt
    #or
    - source: https://raw.githubusercontent.com/mathiasbynens/utf8.js/master/package.json
    #or
    - source: http://someur.com/somefile.json
    #or
    - source: C:\temp\downloadedFile.exe
```

##### source_hash

[string] type, default empty string

Contains the hash sum of the source file in format `[hash_algo]=[hash_sum]`.

```yaml
another-url:
  file.managed:
    - name: /tmp/sub/utf8-js-1.json
    - source: https://raw.githubusercontent.com/mathiasbynens/utf8.js/master/package.json
    - source_hash: sha256=40c5219fc82b478b1704a02d66c93cec2da90afa62dc18d7af06c6130d9966ed
```

Currently tacoscript supports following hash algorithms:

- sha512


    sha512=0e918f91ee22669c6e63c79d14031318cb90b379a431ef53b58c99c4a0257631d5fcd5c4cb3852038c16fe5a2f4fb7ce8277859bf626725a60e45cd6d711d048


- sha384


    sha384=3dc2e2491e8a4719804dc4dace0b6e72baa78fd757b9415bfbc8db3433eaa6b5306cfdd49fb46c0414a434e1bbae5ae3


- sha256


    sha256=5ea41a21fb3859bfe93b81fb0cf0b3846e563c0771adfd0228145efd9b9cb548


- sha224


    sha224=36a2bcb85488ae92c6e2d53c673ba0a750c0e4ff7bfd18161eb08359


- sha1


    sha1=b9456f802d9618f9a7853df1cd011848dd7298a0


- md5


    md5=549e80f319af070f8ea8d0f149a149c2


If `skip_verify` is set to false, tacoscript will check the hash of the target file defined in the `name` field. If it matches with the `source_hash`, the task will be skipped. Further on it will download the file from the `source` field to a temp location and will compare it's hash with the `source_hash` value.

If it doesn't match, tacoscript will fail. The reason for it is that `source_hash` is also used to verify that the source file was successfully downloaded and was not modified during the transmission.

This applies for both urls and local files. If `skip_verify` is set to true, the `source_hash` will be completely ignored. Tacoscript will compare hashes of source and target files by `sha256` algorithm and skip the task if they match.

`source_hash` will be used only to verify the source field. If it's empty and `contents` field is used, the hash won't be checked.

##### makedirs

[bool] type, default false

If the file is located in a path without a parent directory, then the task will fail. If makedirs is set to true, then the parent directories will be created to facilitate the creation of the named file.

Here is an example:

```yaml
another-url:
  file.managed:
    - name: /tmp/sub/some/dir/utf8-js-1.json
    - makedirs: true
```

If `makedirs` was false and dir path at `/tmp/sub/some/dir/` doesn't exist, the task will fail. Otherwise tacoscript will first create directories tree `/tmp/sub/some/dir/` and then place file `utf8-js-1.json` in it.

##### replace

[bool] type, default true

If set to false and the file already exists, the file will not be modified even if changes would otherwise be made. Permissions and ownership will still be enforced, however.

```yaml
another-url:
  file.managed:
    - name: /tmp/sub/some/dir/utf8-js-1.json
    - makedirs: true
    - replace: false
    - user: root
    - group: root
    - mode: 0755
```

A similar behaviour will be enforced by the following script

```yaml
another-url:
  file.managed:
    - name: /tmp/sub/some/dir/utf8-js-1.json
    - makedirs: true
    - creates:  /tmp/sub/some/dir/utf8-js-1.json
    - user: root
    - group: root
    - mode: 0755
```

however the user, group and mode changes will not be applied when `/tmp/sub/some/dir/utf8-js-1.json` exists.

##### skip_verify

[bool] type, default false

If set to true, tacoscript won't verify the hash of the source file from the `source_hash` field. The `source_hash` will be checked against the target file if it maches, the task will be skipped. Tacoscript will download source to a temp folder, calculate it's sha256 hash and compare to the sha256 hash of the target file. If they don't match, file will be replaced or created.
If set to false, tacoscript will check if `source_hash` matches to the hash of the source location. If not, the script will fail with an exception. Further `source_hash` will be checked for the target file. If not matched, file will be replaced/created and skipped otherwise.

```yaml
another-url:
  file.managed:
    - name: /tmp/sub/utf8-js-1.json
    - source: https://raw.githubusercontent.com/mathiasbynens/utf8.js/master/package.json
    - source_hash: sha256=40c5219fc82b478b1704a02d66c93cec2da90afa62dc18d7af06c6130d9966ed
    - skip_verify: true
```

In this script, the file `/tmp/sub/utf8-js-1.json` will be created/replaced only if sha256 hash of source https://raw.githubusercontent.com/mathiasbynens/utf8.js/master/package.json doesn't match with /tmp/sub/utf8-js-1.json.

##### contents

[string] type, default empty string

Multiline UTF-8 encoded string which expected to be the content of the target file. This value exludes the usage of `source` field as tacoscript uses data either from source or contents field. Additionally `source_hash` and `skip_verify` are ignored, if `contents` field is provided.

```yaml
another-file:
  file.managed:
    - name: my-file-win1251.txt
    - contents: |
        goes here
        Funny file
```

In this example we take the contents of the `my-file-win1251.txt` file and compare it with the `contents` field line by line. If they matched, no content modification will be done. If not, the target file `my-file-win1251.txt` will contain `goes here Funny file`, respecting multiline format.

Additionally, you will see in logs something like (assuming that `my-file-win1251.txt` is empty)

```yaml
expected: ""\ngoes here\nFunny file""
actual: """"
Diff:
--- Expected
+++ Actual
-
goes here
Funny file
+
```

which shows what was changed in target file: rows with - are added and rows with + stayed unchanged.

##### mode

[integer] type, default 0

This field shows the desired filemode for the target file. This value will be ignored in Windows. If mode is not set and file exists, no file mode will be changed. If mode is not set and file is created, the defailt mode 0774 will be set to it.

```yaml
another-file:
  file.managed:
    - name: /tmp/myfile.txt
    - contents: one
    - mode: 0777
```

As a result of this execution, file `/tmp/myfile.txt` will have `rwxrwxrwx` rights (see https://en.wikipedia.org/wiki/File-system_permissions)

##### user

[string] type, default empty string

This field shows the desired owner of the target file. This value will be ignored in Windows. By default all files are created with the ownership of the current user. However if `user` field is specified, tacoscript will try to change the ownership of the target file (no matter if it was created or updated) to the desired user name. Of course this would require to run tacoscript with sufficient rights (e.g. as a `root` user).

Additionally, tacoscript will not apply ownership changes to the target file if `onlyif`, `unless`, `creates` conditions failed or hash of the target file matches with `source_hash` value. If `skip_verify` is true and hash of target and source file matched or there was no diff between the `contents` field and the contents of the target file, tacoscript will change the ownership of the target file.

##### group

[string] type, default empty string

This field shows the desired group of the target file. This value will be ignored in Windows. By default all files are created with the ownership of the current user and his default group. However if `group` field is specified, tacoscript will try to change the ownership of the target file (no matter if it was created or updated) to the desired group name. Of course this would require to run tacoscript with sufficient rights (e.g. as a `root` user).
Consider following example:

```yaml
another-url:
  file.managed:
    - name: /tmp/myfile.txt
    - contents: one,two,three
    - user: root
    - group: wheel
```

As a result of this execution (`ls -la /tmp/`), you will see the corresponding owner and group of the target file:

```
       total 10536
       drwxrwxrwt  31 root        wheel      992 Sep  6 21:10 .
       drwxr-xr-x   6 root        wheel      192 Aug 29 09:42 ..
       rwxr-xr-x   2 root        wheel       64 Sep  0 10:18 myfile.txt
```

##### encoding
[string] type, default empty string

This field shows the desired encoding for the content of the target file. This field can be only used in combination with the `contents` field. Tacoscript accepts yaml file only in UTF-8 format. However user can specify a different value in `encoding` field. If target file exists, tacoscript will read it and convert from the specified encoding to UTF-8. Then the `contents` script value will be compared with the decoded contents of the target file. If target file is empty, or contents didn't match, tacoscript will convert the `contents` value to the `encoding` and write the result to the target file.

```yaml
another-file:
  file.managed:
    - name: my-file-win1251.txt
    - contents: One
    - encoding: windows1251
```

After this script, the file `my-file-win1251.txt` will be saved in windows1251 encoding.

The list of supported encoding names:

codepage037, codepage1047, codepage1140, codepage437, codepage850, codepage852, codepage855, codepage858, codepage860, codepage862, codepage863, codepage865, codepage866, iso8859_1, iso8859_10, iso8859_13, iso8859_14, iso8859_15, iso8859_16, iso8859_2, iso8859_3, iso8859_4, iso8859_5, iso8859_6, iso8859_7, iso8859_8, iso8859_9, koi8r, koi8u, macintosh, macintoshcyrillic, windows1250, windows1251, windows1252, windows1253, windows1254, windows1255, windows1256, windows1257, windows1258, windows874, gb18030, gbk, big5, eucjp, iso2022jp, shiftJIS, euckr, utf16be, utf16le, utf8, utf-8

Tacoscript will fail, if an unsupported encoding is provided.

##### require
see [require](/docs#require)

##### creates
see [creates](./s02-conditionals.md#creates)

##### onlyif
see [onlyif](./s02-conditionals.md#onlyif)

##### unless
see [unless](./s02-conditionals.md#unless)

## pkg

### pkg.installed

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

The desired state of the script 'install-neovim' neovim package in version `0.4.3-3` to be installed in the target host system. As shell it will use `bash`.Before running installation it will update list of available packages (`refresh` = true). If this script is executed under the Ubuntu/Debian Linux OS, the tacoscript will execute this command as:

```bash
    - apt update -y
    - apt install -y neovim=0.4.3-3
```

The script will be only executed after `create-file` script. So the tacoscript interpreter will make sure that the `create-file` script is executed before `install-neovim` and only if it was successful.

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

```bash
    apt update
    apt install -y vim neovim vi
```

When you use multiple packages, the version value will be applied to all packages. If version value is empty, the tacoscript will install the latest versions of all packages.

#### Task parameters

##### name

[string] type, required

Name of the package to be installed.

##### names
[array] type

Name contains the list of packages to be installed.

##### shell
[string] type

Shell is a program that takes commands from input and gives them to the operating system to perform. Known Linux shells are [bash](https://www.gnu.org/software/bash/), [sh](https://www.gnu.org/software/bash/), [zsh](https://ohmyz.sh/) etc. Windows supports [cmd.exe](https://ss64.com/nt/cmd.html) shell.

##### require
see [require](/docs#require)

##### onlyif
see [onlyif](./s02-conditionals.md#onlyif)

##### unless
see [unless](./s02-conditionals.md#unless)

##### version

[string] type, optional

Version of the package to be installed. If ommitted, the tacoscript will install the latest default version.

##### refresh
[bool] type, optional
If true, the tacoscript will update list of available packages, e.g. execute `apt update` under Ubuntu/Debian OS.

#### OS Support
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
<td>choco</td>
<td>choco install -y vim</td>
</tr>
</table>

Note if a corresponding package manager is not installed in the host system, a fallback one will be used. If both are not available, the script will fail.

### pkg.uptodate

The task `pkg.uptodate` ensures that the specified package(s) is upgraded to the latest version.

`pkg.uptodate` has following format:

```yaml
upgrade-neovim:
  pkg.uptodate:
    - name: neovim
```

This script will upgrade `neovim` package to the latest stable version.
If this script is executed under the Ubuntu/Debian Linux OS, the tacoscript will execute this command as:

    - apt upgrade -y neovim

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

```bash
apt update
apt upgrade -y vim neovim vi.
```

#### Task parameters

##### name

[string] type, required

Name of the package to be upgraded.

##### names
[array] type

Name contains the list of packages to be upgraded.

##### shell
[string] type

See #pkg.installed for reverence.

##### require
see [require](/docs#require)

##### creates
see [creates](./s02-conditionals.md#creates)

##### onlyif
see [onlyif](./s02-conditionals.md#onlyif)

##### unless
see [unless](./s02-conditionals.md#unless)

##### refresh
[bool] type, optional
See #pkg.installed for reverence.

#### OS Support
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
<td>choco</td>
<td>choco upgrade -y vim</td>
</tr>
</table>

Note if a corresponding package manager is not installed in the host system, a fallback one will be used. If both are not available, the script will fail. If the package is not yet installed, the behaviour will vary depending on the package manager e.g. `brew` will install this package and `apt-get` will fail.


### pkg.removed

The task `pkg.removed` ensures that the specified package(s) is removed from the host system.

`pkg.removed` has following format:

```yaml
delete-neovim:
  pkg.removed:
    - name: neovim
```

This script will delete `neovim` package.
If this script is executed under the Ubuntu/Debian Linux OS, the tacoscript will execute this command as:

```bash
- apt remove -y neovim
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

```bash
apt update
apt remove -y vim neovim vi.
```

#### Task parameters

##### name

[string] type, required

Name of the package to be removed.

##### names
[array] type

Name contains the list of packages to be removed.

##### shell
[string] type

See #pkg.installed for reverence.

##### require
see [require](/docs#require)

##### onlyif
see [onlyif](./s02-conditionals.md#onlyif)

##### unless
see [unless](./s02-conditionals.md#unless)

##### refresh
[bool] type, optional
See #pkg.installed for reverence.

#### OS Support
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
<td>choco</td>
<td>choco uninstall -y vim</td>
</tr>
</table>

Note if a corresponding package manager is not installed in the host system, a fallback one will be used. If both are not available, the script will fail.
