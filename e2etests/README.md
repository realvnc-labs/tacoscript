## End-to-end tests

The function `TestTacoScript` takes all yaml files in the `e2e_test` folder and executes the included tacoscript
and compares the results with you defined expectations.

### Define Tacoscript and expected results

```yaml
Run:
  zchange shell:
    cmd.run:
      - name: print("hallo python")
      - shell: /usr/bin/env python3
On:
  - darwin
  - linux
Expect:
  Summary:
    Succeeded: 1
    Changes: 1
    Failed: 0
    Aborted: 0
    TotalTasksRun: 1

  TaskResults:
    - ID: zchange shell
      Result: true
      ChangesContains:
        - "stdout: hallo python"
```

Optionally, you can run a shell command (`/bin/sh` or PowerShell) before and after the Tascoscript execution. 
This way you can prepare the test environment, clean up or do further validations.

The [`./debian-pkgs.yaml`](./debian-pkgs.yaml) script demonstrates the `PreExec`, `PostExec` and `OnlyIf` options.


### Run test
All e2e tests are run together with
```shell
go test -v -run TestTacoScript ./e2etests/
```

### Run a single e2e test

```shell
go test -v -run TestTacoScript ./e2etests/ -script=write-file-unix.yaml
```