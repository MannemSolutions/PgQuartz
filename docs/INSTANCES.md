# Matrix and job instances
Steps can be defined once, and run multiple times with different arguments.
This is done by specifying all values of arguments in a matrix.
Internally, PgQuartz will convert the matrix of arguments into a list of combinations.

Example:
```
matrix:
  arg1: ["1", "2"]
  arg2: ["A", "B"]
```

Would be converted into the following instances:
```
{"arg1": "1", "arg2": "A"}
{"arg1": "1", "arg2": "B"}
{"arg1": "2", "arg2": "A"}
{"arg1": "2", "arg2": "B"}
```

PgQuartz then schedules and runs every instance separately.
As such, (with enough runners) the step would be run 4 times, with the arguments set accordingly.

> **Note** that without specifying a matrix, the step would be run only once, without any arguments being set.

## Passing arguments

### Shell scripts

When running commands as shell scrips, the arguments are passed as environment variables.
As an example:
```
{"arg1": "1", "arg2": "A"}
```
would be run as
```
PGQ_INSTANCE_ARG1=1 PGQ_INSTANCE_ARG2=A /path/to/step/command.sh
```

Inside the script the arguments can be accessed through the names of the environment variables, like:
```
echo "arg1: ${PGQ_INSTANCE_ARG1}, arg2: ${PGQ_INSTANCE_ARG2}"
```
Which would create a stdout value containing:
```
arg1: 1, arg2: A
```

### PostgreSQL scripts

When running a PostgreSQL command, PgQuartz does the following:
- When in batch mode, PgQuartz splits the command by ';' characters into multiple queries, and does all of this for every query.
  - When not in batch mode, PgQuartz expects the query to be one query and does all of this for the one query.
- PgQuartz scans the query for named arguments (e.a. `:argname`) and replaces them with positional arguments (e.a. `$1`) while maintaining a list of the arguments values
- PgQuartz runs the query with positional arguments while passing the arguments as a list of positional arguments

This does mean that:
- arguments can (only) be passed by name specifying `:argname` placeholders in your query as required
- PgQuartz runs them as positional arguments, so your queries ed up in PostgreSQL logs with `$n` placeholders instead
- But, at least the interface to both PostgreSQL scripts and bash scripts is the same (named arguments)

## Example config
An example of running just one step, but with 6 different combinations of arguments, 6 times in parallel
```
steps:
  step 1:
    commands:
      - name: Run step 1.1
        type: shell
        inline: 'touch "/tmp/${PGQ_INSTANCE_ARG1}_${PGQ_INSTANCE_ARG2}"'
      - name: Run step 1.2
        type: pg
        inline: insert into t1 (id, txt) values(:arg1::integer, :arg2)
    matrix:
      arg1: ["1", "2"]
      arg2: ["A", "B", "C"]
parallel: 6
```

### What does it do?
This job will create 6 runners, and each runner will run the step, but with different arguments.
All instances will run in parallel (**note that even with queuing, order is not enforced between instances).

Assuming that we have
- a /tmp location which is writable, and does not have the files already existing
- a t1 table which is empty and can hold the values (id integer, txt text)

We end up with:
1. The following (empty) files:
   - /tmp/1_A
   - /tmp/1_B
   - /tmp/1_C
   - /tmp/2_A
   - /tmp/2_B
   - /tmp/2_C
2. A table with six rows:
   - id=1, txt='A'
   - id=1, txt='B'
   - id=1, txt='C'
   - id=2, txt='A'
   - id=2, txt='B'
   - id=2, txt='C'
