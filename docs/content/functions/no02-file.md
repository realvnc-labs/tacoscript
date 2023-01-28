---
title: 'File'
weight: 2
slug: file
---

{{< toc >}}

## Preface

With tacoscript you can manage files by settings their contents directly from the yaml file or by copying them from
different sources.

## `file.managed`

The task `file.managed` ensures the existence of a file in the local file system. It can download files from remote urls
(currently http(s)/ftp protocols are supported) or copy a file from the local file system. It can verify the checksums
processed files and show content diffs in the source and target files.

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

1. Download a file from `https://github.com/notepad-plus-plus/notepad-plus-plus/releases/download/v7.8.8/npp.7.8.8.Installer.x64.exe`
   to some temp location.
2. Check if md5 hash of it matches `79eef25f9b0b2c642c62b7f737d4f53f`. If not, skip the task.
3. Check if md5 hash of the target file `C:\temp\npp.7.8.8.Installer.x64.exe` matches `79eef25f9b0b2c642c62b7f737d4f53f`,
   if yes, it means the file exists and has the desired content, so the task will be skipped.
4. The tacoscript should make directory tree `C:\temp\progs`, if needed. If file at `C:\temp\npp.7.8.8.Installer.x64.exe`
   exists, it won't be replaced even if it has a different content. The task will be skipped if the file
   `C:\Program Files\notepad++\notepad++.exe` already exists.

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

1. Copy contents `My file content\ngoes here\nFunny file` to the `/tmp/my-file.txt` file.
2. Don't check the hashsum of it. Implicitly tacoscript will compare the contents of the target file with the provided
   content and show the differences.
3. If the contents don't differ, the task will be skipped. If file doesn't exist, it will be created. Tacoscript will
   make sure, that the file `/tmp/my-file.txt` is owned by user `root`, `group` - `www-data`, and has file mode 0755.
   The target content will be encoded as `UTF-8`. The task will be skipped if the target system has no `apache2` installed.

{{< heading-supported-parameters >}}

### `name`

{{< parameter required=1 type=string >}}

Name is the file path of the target file. A `file.managed` will make sure that the file `/tmp/targetfile.txt` is created
or has the expected content.

```yaml
create-file:
  file.managed:
    - name: /tmp/targetfile.txt
```

### `source`

{{< parameter required=0 type=string >}}

URL or local path of the source file which should be copied to the target file. Source can be HTTP, HTTPS or FTP URL or
a local file path. See some examples below:

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

### `source_hash`

{{< parameter required=0 type=string >}}

Contains the hash sum of the source file in format `[hash_algo]=[hash_sum]`.

```yaml
another-url:
  file.managed:
    - name: /tmp/sub/utf8-js-1.json
    - source: https://raw.githubusercontent.com/mathiasbynens/utf8.js/master/package.json
    - source_hash: sha256=40c5219fc82b478b1704a02d66c93cec2da90afa62dc18d7af06c6130d9966ed
```

Currently, tacoscript supports following hash algorithms:

sha512
: `sha512=0e918f91ee22669c6e63c79d14031318cb90e...truncated...fe5a2f4fb7ce8277859bf626725a60e45cd6d711d048`

sha384
: `sha384=3dc2e2491e8a4719804dc4dace0b6e72baa78fd757b9415bfbc8db3433eaa6b5306cfdd49fb46c0414a434e1bbae5ae3`

sha256
: `sha256=5ea41a21fb3859bfe93b81fb0cf0b3846e563c0771adfd0228145efd9b9cb548`

sha224
: `sha224=36a2bcb85488ae92c6e2d53c673ba0a750c0e4ff7bfd18161eb08359`

sha1
: `sha1=b9456f802d9618f9a7853df1cd011848dd7298a0`

md5
: `md5=549e80f319af070f8ea8d0f149a149c2`

If `skip_verify` is set to false, tacoscript will check the hash of the target file defined in the `name` field.
If it matches with the `source_hash`, the task will be skipped. Further on it will download the file from the `source`
field to a temp location and will compare it's hash with the `source_hash` value.

If it doesn't match, tacoscript will fail. The reason for it is that `source_hash` is also used to verify that the
source file was successfully downloaded and was not modified during the transmission.

This applies for both urls and local files. If `skip_verify` is set to true, the `source_hash` will be completely
ignored. Tacoscript will compare hashes of source and target files by `sha256` algorithm and skip the task if they match.

`source_hash` will be used only to verify the source field. If it's empty and `contents` field is used, the hash won't
be checked.

### `makedirs`

{{< parameter required=0 type=boolean default="false" >}}

If the file is located in a path without a parent directory, then the task will fail. If `makedirs` is set to true, then
the parent directories will be created to facilitate the creation of the named file.

Here is an example:

```yaml
another-url:
  file.managed:
    - name: /tmp/sub/some/dir/utf8-js-1.json
    - makedirs: true
```

If `makedirs` was false and dir path at `/tmp/sub/some/dir/` doesn't exist, the task will fail. Otherwise, tacoscript
will first create directories tree `/tmp/sub/some/dir/` and then place file `utf8-js-1.json` in it.

### `replace`

{{< parameter required=0 type=boolean default="true" >}}

If set to false and the file already exists, the file will not be modified even if changes would otherwise be made.
Permissions and ownership will still be enforced, however.

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
    - creates: /tmp/sub/some/dir/utf8-js-1.json
    - user: root
    - group: root
    - mode: 0755
```

however the user, group and mode changes will not be applied when `/tmp/sub/some/dir/utf8-js-1.json` exists.

### `skip_verify`

{{< parameter required=0 type=boolean default="false" >}}

If set to true, tacoscript won't verify the hash of the source file from the `source_hash` field. The `source_hash` will
be checked against the target file if it maches, the task will be skipped. Tacoscript will download source to a temp folder,
calculate it's sha256 hash and compare to the sha256 hash of the target file. If they don't match, file will be replaced
or created. If set to false, tacoscript will check if `source_hash` matches to the hash of the source location.
If not, the script will fail with an exception. Further `source_hash` will be checked for the target file. If not matched,
file will be replaced/created and skipped otherwise.

```yaml
another-url:
  file.managed:
    - name: /tmp/sub/utf8-js-1.json
    - source: https://raw.githubusercontent.com/mathiasbynens/utf8.js/master/package.json
    - source_hash: sha256=40c5219fc82b478b1704a02d66c93cec2da90afa62dc18d7af06c6130d9966ed
    - skip_verify: true
```

In this script, the file `/tmp/sub/utf8-js-1.json` will be created/replaced only if sha256 hash of source
`https://raw.githubusercontent.com/mathiasbynens/utf8.js/master/package.json` doesn't match with `/tmp/sub/utf8-js-1.json`.

### `contents`

{{< parameter required=0 type=string >}}

UTF-8 encoded string which expected to be the content of the target file. This value excludes the usage of `source`
field as tacoscript uses data either from source or contents field. Additionally `source_hash` and `skip_verify` are
ignored, if `contents` field is provided.

```yaml
another-file:
  file.managed:
    - name: my-file-win1251.txt
    - contents: |
        goes here
        Funny file
```

In this example we take the contents of the `my-file-win1251.txt` file and compare it with the `contents` field line
by line. If they matched, no content modification will be done. If not, the target file `my-file-win1251.txt` will
contain `goes here Funny file`, respecting multiline format.

Additionally, you will see in logs something like (assuming that `my-file-win1251.txt` is empty)

```text
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

### `mode`

{{< parameter required=0 type=integer default="0774" >}}

This field shows the desired filemode for the target file. This value will be ignored in Windows. If mode is not set and
file exists, no file mode will be changed. If mode is not set and file is created, the default mode 0774 will be set to it.

```yaml
another-file:
  file.managed:
    - name: /tmp/myfile.txt
    - contents: one
    - mode: 0777
```

As a result of this execution, file `/tmp/myfile.txt` will have `rwxrwxrwx` rights
(see <https://en.wikipedia.org/wiki/File-system_permissions>)

### `user`

{{< parameter required=0 type=string >}}

This field shows the desired owner of the target file. This value will be ignored in Windows. By default, all files are
created with the ownership of the current user. However, if `user` field is specified, tacoscript will try to change the
ownership of the target file (no matter if it was created or updated) to the desired username. Of course this would require
running tacoscript with sufficient rights (e.g. as a `root` user).

Additionally, tacoscript will not apply ownership changes to the target file if `onlyif`, `unless`, `creates` conditions
failed or hash of the target file matches with `source_hash` value. If `skip_verify` is true and hash of target and source
file matched or there was no diff between the `contents` field and the contents of the target file, tacoscript will change
the ownership of the target file.

### `group`

{{< parameter required=0 type=string >}}

This field shows the desired group of the target file. This value will be ignored in Windows. By default, all files are
created with the ownership of the current user and his default group. However if `group` field is specified, tacoscript
will try to change the ownership of the target file (no matter if it was created or updated) to the desired group name.
Of course this would require running tacoscript with sufficient rights (e.g. as a `root` user).
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

```text
total 10536
drwxrwxrwt  31 root        wheel      992 Sep  6 21:10 .
drwxr-xr-x   6 root        wheel      192 Aug 29 09:42 ..
rwxr-xr-x   2 root        wheel       64 Sep  0 10:18 myfile.txt
```

### `encoding`

{{< parameter required=0 type=string >}}

This field shows the desired encoding for the content of the target file. This field can be only used in combination
with the `contents` field. Tacoscript accepts yaml file only in UTF-8 format. However, users can specify a different
value in `encoding` field. If target file exists, tacoscript will read it and convert from the specified encoding to UTF-8.
Then the `contents` script value will be compared with the decoded contents of the target file. If target file is empty,
or contents didn't match, tacoscript will convert the `contents` value to the `encoding` and write the result to the
target file.

```yaml
another-file:
  file.managed:
    - name: my-file-win1251.txt
    - contents: One
    - encoding: windows1251
```

After this script, the file `my-file-win1251.txt` will be saved in windows1251 encoding.

{{< expand "list of supported encoding names" "click to expand..." >}}

- codepage037
- codepage1047
- codepage1140
- codepage437
- codepage850
- codepage852
- codepage855
- codepage858
- codepage860
- codepage862
- codepage863
- codepage865
- codepage866
- iso8859_1
- iso8859_10
- iso8859_13
- iso8859_14
- iso8859_15
- iso8859_16
- iso8859_2
- iso8859_3
- iso8859_4
- iso8859_5
- iso8859_6
- iso8859_7
- iso8859_8
- iso8859_9
- koi8r
- koi8u
- macintosh
- macintoshcyrillic
- windows1250
- windows1251
- windows1252
- windows1253
- windows1254
- windows1255
- windows1256
- windows1257
- windows1258
- windows874
- gb18030
- gbk
- big5
- eucjp
- iso2022jp
- shiftJIS
- euckr
- utf16be
- utf16le
- utf8
- utf-8
  {{< /expand >}}
  Tacoscript will fail, if an unsupported encoding is provided.

## `file.replace`

The task `file.replace` allows you to replace the contents of files using regular expressions.

`file.replace` has following format:

```yaml
maintain-my-file:
  file.replace:
    - name: ~/.aws/config
    - pattern: region = us-east-1
    - repl: region = ap-southeast-1
    - onlyif:
      ~/.aws/config
```

We can interpret this script as:

1. Only perform this task if the file `~/.aws/config` exists
2. Using `~/.aws/config`, replace `region = us-east-1` with `region = ap-southeast-1`
3. No backup will be made as the `backup` property is not set

Here is another `file.replace` task:

```yaml
another-file:
  file.replace:
    - name: ~/.aws/config
    - pattern: region = us-east-2(.*)$
    - repl: region = us-west-2$1
    - backup: "bak"
    - onlyif:
      ~/.aws/config
```

We can read it as following:

1. Only perform this task if the file `~/.aws/config` exists
2. Using `~/.aws/config` then using a regex pattern, replace the `us-east-2` region with `ap-southeast-2` but keep the same availability zone
3. Make a backup with the extension `.bak`

{{< heading-supported-parameters >}}

### `name`

{{< parameter required=1 type=string >}}

`name` is the file path of the target file. `file.replaced` will search the contents of this file for the `pattern` provided.

### `pattern`

{{< parameter required=1 type=string >}}

Contains the regular expression to search for in the target file. The regular expressions supported use the golang regexp engine.
Group replacements are supported using the standard `()` and `$` notations.

### `replace`

{{< parameter required=1 type=string >}}

This is the text that will replace any text matching the `pattern` specified in the file indicated by `name`.

### `count`

{{< parameter type=string >}}

The number of replacements in the target file can be limited by using the `count` parameter.

### `append_if_not_found`

{{< parameter type=string >}}

If set to `true` then either the text specified by the `not_found_content` will be appended to the target file. If the
`not_found_context` parameter is empty then any content in the `repl` parameter will be used.

### `prepend_if_not_found`

{{< parameter type=string >}}

If set to `true` then either the text specified by the `not_found_content` will be prepended to the target file. If the
`not_found_context` parameter is empty then any content in the `repl` parameter will be used.

### `not_found_content`

{{< parameter type=string >}}

The text specified by this property will Used when either the `append_if_not_found` or `prepend_if_not_found` parameters
re set to `true`. The text will be appended or prepended to the target file accordingly.

### `backup`

{{< parameter type=string >}}

If set then this value will be used as an filename extension to be added to a backup copy of the target file. If not set
then no backup file will be created.

### `max_file_size`

{{< parameter type=string default="500k">}}

If set then target files whose size is greater will be skipped.
