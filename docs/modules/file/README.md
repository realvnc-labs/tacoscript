# file.managed

The task `file.managed` ensures the existence of a file in the local file system. It can download files from remote urls (currently http(s)/ftp protocols are supported) or copy a file from the local file system. It can verify the checksums processed files and show content diffs in the source and target files. 

`file.managed` has following format:

    maintain-my-file:
      file.managed:
        - name: C:\temp\progs\npp.7.8.8.Installer.x64.exe
        - source: https://github.com/notepad-plus-plus/notepad-plus-plus/releases/download/v7.8.8/npp.7.8.8.Installer.x64.exe
        - source_hash: md5=79eef25f9b0b2c642c62b7f737d4f53f
        - makedirs: true # default false
        - replace: false # default true
        - creates: 'C:\Program Files\notepad++\notepad++.exe'

We can interpret this script as:
Download a file from `https://github.com/notepad-plus-plus/notepad-plus-plus/releases/download/v7.8.8/npp.7.8.8.Installer.x64.exe` to some temp location.
Check if md5 hash of it matches `79eef25f9b0b2c642c62b7f737d4f53f`. If not, skip the task. Check if md5 hash of the target file `C:\temp\npp.7.8.8.Installer.x64.exe` matches `79eef25f9b0b2c642c62b7f737d4f53f`, if yes, it means the file exists and has the desired content, so the task will be skipped.
The tacoscript should make directory tree `C:\temp\progs`, if needed. If file at `C:\temp\npp.7.8.8.Installer.x64.exe` exists, it won't be replaced even if it has a different content. The task will be skipped if the file `C:\Program Files\notepad++\notepad++.exe` already exists.

Here is another `file.managed` task:


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

We can read it as following:
Copy contents `My file content\ngoes here\nFunny file` to the `/tmp/my-file.txt` file. Don't check the hashsum of it. Implicitly tacoscript will compare the contents of the target file with the provided content and show the differences. If the contents don't differ, the task will be skipped. If file doesn't exist, it will be created. Tacoscript will make sure, that the file `/tmp/my-file.txt` is owned by user `root`, `group` - `www-data`, and has file mode 0755. The target content will be encoded as `UTF-8`. The task will be skipped if the target system has no `apache2` installed. 

## Task parameters

### name

[string] type, required

Name is the file path of the target file. A `file.managed` will make sure that the file `/tmp/targetfile.txt` is created or has the expected content.

    create-file:
      file.managed:
        - name: /tmp/targetfile.txt
        
### source, default empty string

URL or local path of the source file which should be copied to the target file. Source can be HTTP, HTTPS or FTP URL or a local file path. See some examples below:

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


### source_hash

[string] type, default empty string

Contains the hash sum of the source file in format `[hash_algo]=[hash_sum]`. 


    another-url:
      file.managed:
        - name: /tmp/sub/utf8-js-1.json
        - source: https://raw.githubusercontent.com/mathiasbynens/utf8.js/master/package.json
        - source_hash: sha256=40c5219fc82b478b1704a02d66c93cec2da90afa62dc18d7af06c6130d9966ed


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

### makedirs

[bool] type, default false

If the file is located in a path without a parent directory, then the task will fail. If makedirs is set to true, then the parent directories will be created to facilitate the creation of the named file.

Here is an example:


    another-url:
      file.managed:
        - name: /tmp/sub/some/dir/utf8-js-1.json
        - makedirs: true
        

If `makedirs` was false and dir path at `/tmp/sub/some/dir/` doesn't exist, the task will fail. Otherwise tacoscript will first create directories tree `/tmp/sub/some/dir/` and then place file `utf8-js-1.json` in it.

### replace

[bool] type, default true

If set to false and the file already exists, the file will not be modified even if changes would otherwise be made. Permissions and ownership will still be enforced, however.
Practically both scripts will be equivalent:

    another-url:
      file.managed:
        - name: /tmp/sub/some/dir/utf8-js-1.json
        - makedirs: true
        - replace: false
        

    another-url:
      file.managed:
        - name: /tmp/sub/some/dir/utf8-js-1.json
        - makedirs: true
        - creates:  /tmp/sub/some/dir/utf8-js-1.json
