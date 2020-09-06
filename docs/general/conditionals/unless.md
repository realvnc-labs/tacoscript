# unless
[string] and [array] type

The `unless` is a reverted `onlyif` parameter which means the task will run only of the `unless` condition fails (returns a non-zero exit code). However if multiple `unless` parameters are used, tacoscript will execute the task only if **one** `unless` condition fails (the `onlyif` parameters should be all successful to unblock the task). 

The examples below show some use cases for this parameter. 
We execute the command `service myservice start` only if we detect that there are no critical logs. Without this condition tacoscript will try to start the service which has no chance to be started.

    start-myservice:
      cmd.run:
        - name: service myservice start
        - unless: test -e myservice_fatallogs.txt

    #or

The second example demonstrates a check for some required configs. If the tacoscript detects any missing config from the list, it will send an email to the server's adminstrator. 
    
    report-failure:
      cmd.run:
        - name: mail -s "Some configs are missing, check what is going on" admin@admin.com
        - unless:
            - test -e criticalConfigOne.txt
            - test -e criticalConfigTwo.txt
            - test -e criticalConfigThree.txt



