# PgQuartz Connection endpoint configuration
The `Connections` block can be used to define one or more PostgreSQL connection endpoints for processing the SQL queries.

## Configuration options

### Connection type
Currently, a type can be set, and `postgresql` is usually set as the type, but setting this field has no effect.
In future releases we might add other connection endpoint types, such as MySQL, MongoDB, etc. if we receive requests from the community.
Please specify an [Issue](https://github.com/MannemSolutions/PgQuartz/issues) to request other endpoint types.

### Role
PgQuartz has the option to only run Queries on Connection endpoints if they have a specific role.
One of the following roles can be specified:
- standby: Only run if the endpoint is a standby (`SELECT pg_is_in_recovery()` returns `true`)
- primary: Only run if the endpoint is a standby (`SELECT pg_is_in_recovery()` returns `false`)
- all: Don't worry, all roles are fine

PgQuartz has 2 types of behavior on Connection Roles:
- When the job configuration option `runOnRoleError=false` (default)
  - PgQuartz errors out if the endpoint does not meet expectations (e.a. configured and actual role differ for a connection)
  - [role configuration on Commands](./COMMANDS.md#role) can be set to skip a certain Command depending on the role of the endpoint
- When the job configuration option `runOnRoleError=true`
  - PgQuartz continues if the endpoint does not meet expectations (e.a. configured and actual role differ)
  - [role configuration on Commands](./COMMANDS.md#role) can still be set to skip a certain Command depending on the role of the endpoint
  - Role configuration on the Connection acts as a default for [roles on Commands](./COMMANDS.md#role)
See [runOnRoleError](./JOBS.md#runonroleerror) for more info.

> **_Note_** that PgQuartz uses [jackc/pgx/v4](https://github.com/jackc/pgx/tree/v4), which also supports setting `target_session_attrs` to target a specific role for the connection.

### Connection parameters
PgQuartz uses [jackc/pgx/v4](https://github.com/jackc/pgx/tree/v4), which also supports setting many [libpq Parameter Key Words](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-PARAMKEYWORDS) including:
- host: Set host name(s) to connect to
- port: Set port to connect to
- user: Set User to connect as
- password: Specify a password; **_NOTE:_** there are far more secure options than specifying a clear text password in a config file, probably maintained in a git repo!!!
- `target_session_attrs` to target a specific role for the connection.
But more parameters can be applied.

## Example
We make the 'Connections concept' more tangible with an example:

### Example config
```
connections:
  pg:
    type: postgresql
    role: all
    conn_params:
      host: /tmp
      port: 5432
      user: postgres
      password: supassword
```