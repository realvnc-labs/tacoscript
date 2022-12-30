---
title: "Installation"
weight: 0
slug: quick-start
---
{{< toc >}}

{{< hint type=note title="Tacoscript and RPort">}}
If you have the rport client already installed, you very likely will have tacoscript installed as well because the
installer script installs it by default. On Windows check the existence of the `C:\Program Files\tacoscript` folder,
on Linux and macOS check `/usr/local/bin/tacpscript`.
{{< /hint >}}

Jump to [our release page](https://github.com/cloudradar-monitoring/tacoscript/releases/tag/latest) and download a binary
for your host OS. Don't forget to download a corresponding md5 file as well and compare the checksums.

## On macOS

```shell
curl -L https://download.rport.io/tacoscript/stable/?arch=Darwin_$(uname -m)|\
tar xzf - -C /usr/local/bin/ tacoscript
```

## On Linux

```shell
curl -L https://download.rport.io/tacoscript/stable/?arch=Linux_$(uname -m)|\
tar xzf - -C /usr/local/bin/ tacoscript
```

## On Windows

```powershell
iwr https://download.rport.io/tacoscript/stable/?arch=Windows_x86_64 `
-OutFile tacoscript.zip
$dest = "C:\Program Files\tacoscript"
mkdir $dest
mkdir "$($dest)\bin"
Expand-Archive -Path tacoscript.zip -DestinationPath $dest -force
mv "$($dest)\tacoscript.exe" "$($dest)\bin"

$ENV:PATH="$ENV:PATH;$($dest)\bin"

[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable(
        "Path", [EnvironmentVariableTarget]::Machine
    ) + ";$($dest)\bin",
    [EnvironmentVariableTarget]::Machine
)
& tacoscript --version
```

## Compile from sources

```shell
git clone https://github.com/cloudradar-monitoring/tacoscript.git
cd tacoscript
go build -o tacoscript main.go
./tacoscript --help
mv tacoscript /usr/local/bin/tacoscript
```
