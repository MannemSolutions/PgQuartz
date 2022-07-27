# Commands
Commands are the basic building blocks out of which [Steps](./STEPS.md) are defined.
Every [Step](./STEPS.md) is defined as a series of Commands, and every [Instance](./INSTANCES.md) runs this series of Commands until one fails.

## Limit your output
> **_Note_** that PgQuartz basically keeps track of all StdOut and StdErr, and that can require a lot of memory.
> PgQuartz has no special handling, flushing to disk capabilities, or other options.
> Therefore it is crucial for developers of PgQuartz jobs to make sure output is limited options like tail and LIMIT!!!

## Configuration options

### Name
Every Command has a name, which is used in output and logging.
**Note** that emptystring (default) is accepted as Name as well.

### Role
Roles can be set per [Connection](./CONNECTIONS.md), but can be overruled per Command.
For more information, please refer to [Roles on connections](./CONNECTIONS.md#Role).

### BatchMode
As a convenience option SQL Queries (Command bodies run against PostgreSQL connections) can be run in `batchMode`, which means they are split by semicolons and then run one at the time.
This splitting is done in a very crude manner, and as such below examples also act as separator characters
- such semicolons within `SQL strings` (like "My text with ;")
- SQL Names ('my name with ;') 
- semicolons used inside PL/PgSQL and other (SQL) code blocks

**Note** that specifying `batchMode: true` is not encouraged because:
- As mentioned before, the implementation is very crude and can easily mess up commands
- there is no technical downfall to specifying every Query as a separate Command
- there is upside to specifying every Query as a separate Command, because all other configuration option like `Name` and `Role` can be set differently for every separate Commands

### Command types
The `type` field of a command can have 2 types of values:
1. `shell` (default), which means 'execute this command in a terminal shell'
2. Any name of a [Connection](CONNECTIONS.md), which means run this command as a SQL Query on the specified connection.

**Note** that PgQuartz verifies the Job definition before running the job, and errors out if anything else is specified as a Command type.

### Inline or file
Command bodies can be either specified inline, and the effect depends on the type of command:
- Inline type bodies against PostgreSQL connections are directly run against the connection
- File type bodies against PostgreSQL connections are read from file into memory and then run against the connection
- Inline type bodies of type shell are written to a file from memory and the file is executed in a shell
- File type bodies of type shell are directly run in a shell

## Example
We make the 'Commands concept' more tangible with an example:

### Example config
```
steps:
  step 1:
    commands:
      - name: Run command 1.1
        type: pg
        inline: |
          CREATE TABLE IF NOT EXISTS t1 (id int, txt text);
          CREATE TABLE IF NOT EXISTS t2 (id int, id2 int);
      - name: Run command 1.2
        type: shell
        inline: "echo 'Done that' > /tmp/beenhere.txt"
      - name: Run command 1.3
        type: pg
        file: ./sql/step_1.3.sql
      - name: Run command 1.4
        type: shell
        file: ./bash/step_1.4.sh
connections:
  pg:
    type: postgresql
    conn_params:
      host: /tmp
```

### What does it do?
When running a job with a specification as shown in the [example](#example-config), PgQuartz will do the following:
1. PgQuartz will create a work queue and add only one [Instance](./INSTANCES.md) of `step 1` to that queue.
2. PgQuartz will create at least one Runner.
3. the Runner starts processing the [Instance](./INSTANCES.md):
    - Command 1.1 is run against the PostgreSQL connection and 
      - (unless it already exists) the table t1 is created
      - (unless it already exists) the table t2 is created
    - If Command 1.1 succeeded, the runner will run Command 1.2 in a shell and add a line to the file
    - If Command 1.2 succeeded, the runner will run Command 1.3, read the actual command from the specified file and run it against the PostgreSQL connection
    - If Command 1.3 succeeded, the runner will run Command 1.3, read the actual command from the specified file and run it in a shell
