---
title: "Conditionals"
weight: 2
slug: conditionals
---
{{< toc >}}

## `creates`

{{< parameter required=0 type=string|array >}}

*`creates` parameter can have both string and array type.*

```yaml
backup-data:
  cmd.run:
    - name: echo 123
    - creates:
        - file1.txt
        - file2.txt
```

or

```yaml
    backup-data:
      cmd.run:
        - name: echo 123
        - creates: file1.txt
```

The `creates` parameter identifies the files which should be missing if you want to run the current task. In other words
if any of the files in the `creates` section exist, the task will never run.

A typical example use case for this parameter would be an exclusive access to a file, e.g. we don't run the backup, if
the file is locked by another process:

```yaml
backup-data:
  cmd.run:
  - names:
    - touch serviceALock.txt
    - tar cvf somedata.txt.tar somedata.txt
    - rm serviceALock.txt
  - creates: serviceALock.txt
```

In this situation we expect that the script is running periodically, so at some point the lock will be removed, and the
backup-data script get a chance to make a backup.

## `onlyif`

{{< parameter required=0 type=string|array >}}

The `onlyif` parameter gives a list of commands which will be executed before the actual task and if any of them fails (
returns non zero exit code), the task will be skipped. However, the failures won't stop the program execution.
The `onlyif` checks are given to prove that the task should run, in case of failure the task will be completely ignored.

In the example below the
command `kafka-console-producer.sh --broker-list localhost:9092 --topic my_topic --new-producer < my_file.txt` won't be
triggered, if tacoscript detects, that kafka or zookeeper services are not running or `messages.txt` file doesn't exist.

`onlyif` parameter can be both string and array e.g.

```yaml
publish-kafka-message:
  cmd.run:
    - name: kafka-console-producer.sh --broker-list localhost:9092 --topic my_topic --new-producer < my_file.txt
    - onlyif: test -e my_file
```

or

```yaml
publish-kafka-message:
  cmd.run:
    - name: kafka-console-producer.sh --broker-list localhost:9092 --topic my_topic --new-producer < messages.txt
    - onlyif: 
        - service kafka status
        - service zookeeper status
        - test -e messages.txt
```

## `unless`

{{< parameter required=0 type=string|array >}}

The `unless` is a reverted `onlyif` parameter which means the task will run only of the `unless` condition fails (
returns a non-zero exit code). However if multiple `unless` parameters are used, tacoscript will execute the task only
if **one** `unless` condition fails (the `onlyif` parameters should be all successful to unblock the task).

The examples below show some use cases for this parameter. We execute the command `service myservice start` only if we
detect that there are no critical logs. Without this condition tacoscript will try to start the service which has no
chance to be started.

```yaml
start-myservice:
  cmd.run:
    - name: service myservice start
    - unless: test -e myservice_fatallogs.txt
```

The second example demonstrates a check for some required configs. If the tacoscript detects any missing config from the
list, it will send an email to the server's adminstrator.

```yaml
report-failure:
  cmd.run:
    - name: mail -s "Some configs are missing, check what is going on" admin@admin.com
    - unless:
        - test -e criticalConfigOne.txt
        - test -e criticalConfigTwo.txt
        - test -e criticalConfigThree.txt
```
