# onlyif
[string] and [array] type

The `onlyif` parameter gives a list of commands which will be executed before the actual task and if any of them fails (returns non zero exit code), the task will be skipped. However, the failures won't stop the program execution. The `onlyif` checks are given to prove that the task should run, in case of failure the task will be completely ignored.

In the example below the command `kafka-console-producer.sh --broker-list localhost:9092 --topic my_topic --new-producer < my_file.txt` won't be triggered, if tacoscript detects, that kafka or zookeeper services are not running or `messages.txt` file doesn't exist.

`onlyif` parameter can be both string and array e.g.

    publish-kafka-message:
      cmd.run:
        - name: kafka-console-producer.sh --broker-list localhost:9092 --topic my_topic --new-producer < my_file.txt
        - onlyif: test -e my_file

    #or
    
    publish-kafka-message:
      cmd.run:
        - name: kafka-console-producer.sh --broker-list localhost:9092 --topic my_topic --new-producer < messages.txt
        - onlyif: 
            - service kafka status
            - service zookeeper status
            - test -e messages.txt

