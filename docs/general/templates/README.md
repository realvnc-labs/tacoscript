# Templates rendering
You can use [golang template language](https://golang.org/pkg/text/template/) in your scripts. 
Template is evaluated before parsing the yaml format.
Templating allows you to use conditions and variables.

## Predefined variables
Here is the list of variables that you can use in your tacoscript templates:

<table>
    <tr>
    <th>Variable</th>
    <th>Possible values</th>
    </tr>
    <tr>
    <td>.taco_os_kernel</td>
    <td>windows, linux, freebsd, see https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63#a-list-of-valid-goos-values for possible other values
    </td>
    </tr>
    <tr>
    <td>.taco_os_family</td>
    <td>darwin, debian, redhat, windows
    </td>
    </tr>
    <tr>
    <td>.taco_os_platform</td>
    <td>darwin, ubuntu, centos, debian, alpine, windows</td>
    </tr>
    <tr>
    <td>.taco_os_name</td>
    <td>mac os x, ubuntu, centos linux, debian gnu/linux, alpine linux, windows server 2019 standard</td>
    </tr>
    <tr>
    <td>.taco_os_version</td>
    <td>10.15.7, 20.04.1 LTS (Focal Fossa), 8 (Core), 10 (buster), 10.0</td>
    </tr>
    <tr>
    <td>.taco_architecture</td>
    <td>x86_64, 386</td>
    </tr>
</table>

The values in the predefined variables are always lowercase, so make sure that you use lowercase values in comparison operators e.g. "redhat" rather than "Redhat".

You can use predefined variables in your templates as:

    template:
      cmd.run:
    {{ if eq .taco_os_family "redhat" }}
        - name: yum --version
    {{ else }}
        - name: apt --version
    {{ end }}

In this case if you execute this script in a "redhat" centos linux environment the template will be converted to the following script:

    template:
      cmd.run:
        - name: yum --version

## Conditions


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


## Variables

    {{$file := "test.txt"}}
    template:
      cmd.run:
        - name: touch {{ $file }}
        - creates:
            - {{ $file }}
            
By defining $file variable we avoid code repetition like:


    template:
      cmd.run:
        - name: touch test.txt
        - creates:
            - test.txt

so if you would like to change the file name, you can do it in just one place.


Support for loops and predefined functions would be provided in the next releases.
