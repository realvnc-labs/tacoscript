---
title: 'Windows Registry'
weight: 4
slug: registry
---

{{< toc >}}

## Preface

Tacoscript comes with functions to modify your windows registry.

## `win_reg.present`

The task `win_reg.present` ensures that the specified registry value is present in the registry.

`win_reg.present` has following format:

```yaml
maintain-my-windows-registry:
  win_reg.present:
    - reg_path: 'HKLM:\System\CurrentControlSet\Control\Terminal Server'
    - name: fDenyTSConnections
    - value: 0
    - type: REG_SZ
```

We can interpret this script as:

1. Using the terminal server registry path, make sure the value with the name `fDenyTSConnections` is set to a `0` string.

{{< heading-supported-parameters >}}

### `reg_path`

{{< parameter required=1 type=string >}}

This is the registry path to check for the desired `value` with the registry `name` specified below.

### `name`

{{< parameter required=1 type=string >}}

This is the `name` of the registry value to ensure is present at the registry path specified above.

### `value`

{{< parameter required=1 type=string >}}

This is the value that must be is present. If there is no value currently then a new value will be set.
If there is an existing value then it will be replaced.

### `type`

{{< parameter required=1 type=string >}}

This is the registry type of the value to be present. If there is an existing value with a different
type then both the value and type will be updated.

## `win_reg.absent`

The task `win_reg.absent` ensures that the specified registry value is present in the registry.

`win_reg.absent` has following format:

```yaml
maintain-my-windows-registry:
  win_reg.absent:
    - path: 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Run'
    - name: VMware User Process
```

We can interpret this script as:

1. Ensure that the registry value with the name `VMware User Process` at the registry path location
   specified is not set / absent

{{< heading-supported-parameters >}}

### `reg_path`

{{< parameter required=1 type=string >}}

This is the registry path to check that the registry value with the `name` specified below is not set.

### `name`

{{< parameter required=1 type=string >}}

This is the `name` of the registry value to ensure is absent at the registry path specified above.

## `win_reg.absent_key`

The task `win_reg.absent_key` ensures that the specified registry key is not in the registry.

`win_reg.absent_key` has following format:

```yaml
maintain-my-windows-registry:
  win_reg.absent_key:
    - reg_path: 'HKCU:\Software\GoProgrammingLanguage'
```

We can interpret this script as:

1. Completely remove the `GoProgrammingLanguage` registry key and all sub-keys and values

{{< heading-supported-parameters >}}

### `reg_path`

{{< parameter required=1 type=string >}}

This is the registry path to ensure is absent. Will recursively remove all sub-keys and values. Use with great caution.
