# creates
[string] and [array] type

`creates` parameter can have both string and array type e.g.

    backup-data:
      cmd.run:
      - name: echo 123
      - creates:
        - file1.txt
        - file2.txt

    #or
    backup-data:
      cmd.run:
      - name: echo 123
      - creates: file1.txt

The `creates` parameter identifies the files which should be missing if you want to run the current task. In other words if any of the files in the `creates` section exist, the task will never run. 

A typical example use case for this parameter would be an exclusive access to a file, e.g. we don't run the backup, if the file is locked by another process:

    backup-data:
      cmd.run:
      - names:
        - touch serviceALock.txt
        - tar cvf somedata.txt.tar somedata.txt
        - rm serviceALock.txt
      - creates: serviceALock.txt
      
In this situation we expect that the script is running periodically, so at some point the lock will be removed, and the backup-data script get a chance to make a backup.
