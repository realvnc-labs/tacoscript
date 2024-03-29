---
title: "Template Engine"
weight: 4
slug: template-engine
---
{{< toc >}}

## Templates rendering

You can use [golang template language](https://golang.org/pkg/text/template/) in your scripts.
The templates are evaluated before parsing the yaml format.
Templating allows you to use conditions and variables.

## Predefined variables

Here is the list of variables and example values that you can use in your tacoscript templates:

`.taco_os_kernel`
: windows, linux, freebsd, [Read more](https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63#a-list-of-valid-goos-values) for possible other values

`.taco_os_family`
: darwin, debian, redhat, windows

`.taco_os_platform`
: darwin, ubuntu, centos, debian, alpine, windows

`.taco_os_name`
: mac os x, ubuntu, centos linux, debian gnu/linux, alpine linux, windows server 2019 standard

`.taco_os_version`
: mac os x, ubuntu, centos linux, debian gnu/linux, alpine linux, windows server 2019 standard

`.taco_architecture`
: x86_64, 386, aarch

The values in the predefined variables are always lowercase, so make sure that you use lowercase values in comparison operators e.g. "redhat" rather than "Redhat".

You can use predefined variables in your templates as:

```yaml
template:
  cmd.run:
{{ if eq .taco_os_family "redhat" }}
    - name: yum --version
{{ else }}
    - name: apt --version
{{ end }}

```

In this case if you execute this script in a "redhat" centos linux environment the template will be converted to the following script:

```yaml
template:
  cmd.run:
    - name: yum --version
```

## Conditions

```yaml
installVim:
  cmd.run:
{{ if eq .taco_os_family "redhat" }}
    - name: yum install vim
{{ else if eq .taco_os_family "debian" }}
    - name: apt install vim
{{ else if eq .taco_os_platform "alpine" }}
    - name: apk add vim
{{ else }}
    - name: echo "Unsupported platform {{ .taco_os_family }}"
{{ end }}
```

## Variables

```yaml
{{$file := "test.txt"}}
template:
  cmd.run:
    - name: touch {{ $file }}
    - creates:
        - {{ $file }}
```

By defining $file variable we avoid code repetition like:

```yaml
template:
  cmd.run:
    - name: touch test.txt
    - creates:
        - test.txt
```

so if you would like to change the file name, you can do it in just one place.

Support for loops and predefined functions would be provided in the next releases.
