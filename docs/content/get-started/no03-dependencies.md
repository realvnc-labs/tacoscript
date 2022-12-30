---
title: "Dependencies"
weight: 3
slug: dependencies
---
{{< toc >}}

## `require`

{{< parameter required=0 type=string|array >}}

*`require` parameter can have string or array type.*

```yaml
backup-data:
  cmd.run:
  - name: echo 123
  - require:
    - script1
    - script2
script1:
    ...
script2:
    ...
```

or

```yaml
backup-data:
  cmd.run:
  - name: echo 123
  - require: script1
script1: ...
```

Require section contains the names of the scripts which should be executed before the current one. Imagine you want to
download some remote zip file and unzip it to some location. You have to be sure, that unzip is installed in the system
and the file is downloaded to a required location. In this case unzip would require unzip utilities to be installed and
remove file to be downloaded:

```yaml
install-prereq:
   cmd.run:
     - name: apt install -y unzip
download-file:
   file.managed:
     - name: /tmp/myfile.zip
     - source: https://someremoteurl.com/somefile.zip
     - source_hash: sha256=40c5219fc82b478b1704a02d66c93cec2da90afa62dc18d7af06c6130d9966ed
     - makedirs: true
unzip-file:
   cmd.run:
     - name: unzip /tmp/myfile.zip
     - require:
         - install-prereq
         - download-file>
```

Since the `unzip-file` requires `install-prereq` and `download-file`, so the tacoscript will make sure that they are
executed before the `unzip-file`.
