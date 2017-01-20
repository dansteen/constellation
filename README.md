# constellation
Constellation is a tool to spin up a "constellation" of rkt pods (see what I did there?) in a controlled fashion.   It allows you to specify a list of containers to spin up, their success or failure conditions, and their interdependencies such that dependent containers are not spun up until the containers they depend on have encountered a "success" condition.  Constellation will also ensure that networking is set up between the containers so that dependent containers can talk to their dependencies.  This is the rkt equivilent of [Controlled-compose](https://github.com/dansteen/controlled-compose) for docker.

# Networking
Constellation creates a rkt "contained network" for each `projectName` (as defined below), and all containers run under that project are on the same "contained network" and have access to each other.   Ports specified in the container manifest will be exported to the local machine (via the --ports mechanism) and assigned a random port on the local machine.  These ports are printed out at the end of the constellation run. 

# Examples
These examples go in ascending order of complexity.
## A Simple Application
The simplest invocation would spin up a single application and use the default command baked into the container:
```yaml
api.app.local:
    image: aci-repo.example.com/api:af457b220597aa34b739bff13afc514ba72e8100
```
You would then run this using `sudo ./constellation run -c api.yml` and a single container would be spun up.

## A Simple Applications with some conditions
Lets add in some success and failure conditions
```yaml
containers:
  api.app.local:
    image: aci-repo.example.com/api:af457b220597aa34b739bff13afc514ba72e8100
    state_conditions:
      timeout:
        duration: 300
        status: failure
      output:
        - source: STDOUT
          regex: The server is now ready to accept connections
          status: success
        - source: STDERR
          regex: ERROR
          status: failure
```
This tells constellation to run the container with the default command, but also checks STDOUT for the string `The server is now ready to accept connections`, and STDERR for the string `ERROR`.  If it finds the STDOUT string, the container will be marked as having started successfully, if it finds the STDERR string, it will be marked as having failed.   Additionally, there is a timeout - if no other state_condition has occurred after 300 (seconds), then the container will be marked as having failed.   Note that the first state_condition to occur stops monitoring for other state conditions.  So once a success or failure condition has happened, no other conditions will change that.

## An application and its database
Now lets say our application requires a database. We can set that up as follows:
```yaml
containers:
  db.local:
    image: docker://postgres:9.6
    environment:
      POSTGRES_USER: appuser
      POSTGRES_DB: scratch
      POSTGRES_PASSWORD: password
    state_conditions:
      output:
        - source: STDOUT
          regex: PostgreSQL init process complete; ready for start up.
          status: success
      timeout:
        duration: 30
        status: failure
  api.app.local:
    image: aci-repo.example.com/api:af457b220597aa34b739bff13afc514ba72e8100
    state_conditions:
      timeout:
        duration: 300
        status: failure
      output:
        - source: STDOUT
          regex: The server is now ready to accept connections
          status: success
        - source: STDERR
          regex: ERROR
          status: failure
    depends_on:
      - db.local
```
This will spin up a postgres container, and monitor it for a success string.  If it has not found that string after 30 seconds, it will mark the postgres container as having failed, and will exit (no other containers will be started).  If it does find the success string it will then move on and start up our application.   The application can access the database by the name of the container (`db.local` in this case).  As an additional point, note that we are setting the environment variables that are passed into the postgres container.

## An application and its database and some database config
A common need is to prep a database via migration scripts or similar.   This can be done as well:
```yaml
containers:
  db.local:
    image: docker://postgres:9.6
    environment:
      POSTGRES_USER: appuser
      POSTGRES_DB: scratch
      POSTGRES_PASSWORD: password
    state_conditions:
      output:
        - source: STDOUT
          regex: PostgreSQL init process complete; ready for start up.
          status: success
      timeout:
        duration: 30
        status: failure
  api.migrate.tmp:
    image: aci-repo.example.com/api:af457b220597aa34b739bff13afc514ba72e8100
    exec: /opt/jdk/bin/java -jar /opt/api/api.jar db migrate /etc/api-config.yml
    state_conditions:
      exit:
        codes: [0]
        status: success
      timeout:
        duration: 300
        status: failure
    depends_on:
      - db.local
  api.app.local:
    image: aci-repo.example.com/api:af457b220597aa34b739bff13afc514ba72e8100
    state_conditions:
      timeout:
        duration: 300
        status: failure
      output:
        - source: STDOUT
          regex: The server is now ready to accept connections
          status: success
        - source: STDERR
          regex: ERROR
          status: failure
    depends_on:
      - api.migrate.tmp
```
This will spin up a postgres db, and then run migration scripts against it.  Constellation will look for an exit code of `0` from the migration scripts, and if found, will treat that as a successful run and will move on to the next container.  Any other exit will result in a failure, and taking longer than 300 seconds will also result in a failure.  Note that we chain our container dependencies so that startup will be in order db.local --> api.migrate.tmp --> api.app.local.  You may use arbitrarily complex dependency relationships, however, dependency loops will result in an error.

## Monitoring log files for state_conditions
It is also possible to monitor log files for regex, and use the results in state conditions:
```yaml
volumes:
  app-log-dir:
    kind: host
    path: /tmp/app-logs
    uid: 9998
    gid: 9998
    mode: 0755
    
containers:
  api.app.local:
    image: aci-repo.example.com/api:af457b220597aa34b739bff13afc514ba72e8100
    mounts:
      - volume: app-log-dir
        path: /var/log/api
    state_conditions:
      filemonitor:
        - file: /var/log/api/api_application.log
          regex: "org.eclipse.jetty.server.Server | Started"
          status: success
        - file: /var/log/api/api_application.log
          regex: ERROR
          status: failure
      timeout:
        duration: 300
        status: failure
```
This will start the application, and monitor the file `/var/log/api/api_application.log` for success and failure strings.   If it does not find one in 300 seconds it will register the application as having failed.   To understand the above config, it is important to realize that **file monitoring is done by Constellation outside of the container**.   This means that you have to first export the directory that contains the file you want to monitor.   The stanza:
```yaml
mounts:
 - volume: app-log-dir
   path: /var/log/api
```
tells rkt that you want to mount the external volume named `app-log-dir` into the container at the path `/var/log/api`.   Then, the stanzas:
```yaml
volumes:
  app-log-dir:
    kind: host
    path: /tmp/app-logs
    uid: 9998
    gid: 9998
    mode: 0755
```
define the `app-log-dir` volume and tell rkt which external folder to use when the volume `app-log-dir` is referenced in a mount.  Notice that `volumes` are defined constellation-wide and are not specific to a container.  The path in `file` should be the path to the file you want monitored *inside* the container.  Constellation will handle correct path adjustment to find the intended file.

## Spreading config across multiple files
It is common to have multiple applications that depend on each other.  We don't want to have to keep all of our config together, and we don't want to have to redefine the same config more than once.   Constellation allows the importing of other constellation configs:
```yaml
require:
  - postgres.yml

containers:           
  api.app.local:
    image: aci-repo.example.com/api:af457b220597aa34b739bff13afc514ba72e8100
    state_conditions:
      timeout:
        duration: 300
        status: failure
      output:
        - source: STDOUT
          regex: The server is now ready to accept connections
          status: success
        - source: STDERR
          regex: ERROR
          status: failure
    depends_on:
      - db.local
```

Notice that we do not define `db.local` in this constellation config file.  Rather we reference an additional config file `postgres.yml` that contains that definition.   When using `require` you will need to have the `require`d files either in the same directory as the file that requires them, or you will need to add the `-I` flag to your command invocation with a path to the location of the required file.

# Reference

## CLI
Constellation can be invoked with the following commands:
| Command | Description 
| --- | --- |
| run | Run the containers described in the config file
| stop | Stop the containers that are part of the Project Name defined with -p
| clean | Stop and remove the containers taht are part of the Project name defined with -p

The following flags are supported:

| Flag | Name | Description | Required
| ---- | ---- | ----------- | --------
| -c | Constellation Config | Path to the constellation config file that defines your applications | yes
| -p | Project Name | A unique name for this invocation.  Containers started are tagged with this name, and this is used to `stop` and `clean` the containers | yes
| -H | Hosts Entries | Extra entries for the /etc/hosts file in all containers.  Useful for external resources | no
| -i | Image Overrides | Overrides the versions of images in the config file | no
| -I | Include Directories | Directories to search for config files included using the `require` stanza | no
| -v | Volume Overrides | Overide the volumes defined in the config file. Must be an absolute path. | no

## Config Stanzas
The following config Stanzas are supported:

### Base Config
The following base stanzas are supported.  See below for more information about each of them.

| Stanza | Description | 
| ------ | ----------- |
| require | A list of constellation config files. File names provided here will be processed along with (prior to) the config file that includes them. Note that only filenames should be here not full paths.  Paths to files must be included in the `-I` CLI flag unless the file is in the same directory as the file that is calling it. |
| volumes | A hash of volume names.  Volumes named here can be referenced in the `mounts` stanza of the container definition and mounted into containers.  They can also be overriden using the `-v` flag. |
| containers | A hash of container definitions. The base stanza for our container definitions. |


#### Require
This is a list of config files to include.

#### Volumes
A hash of volume definitions.

| Parameter | Values | Description | Required |
| --------- | ------ | ----------- | -------- |
| kind | `host` \| `empty` | The type of volume this should be.  Note that only type `host` can be used with filemonitor state_conditions. | yes |
| path | `<filepath>` | the local path (external to the container) that you want to mount into the container | yes |
| uid  | numeric <uid> | the uid to set as the owner of `path` | yes |
| gid  | numeric <gid> | the gid to set as the owner of `path` | yes |
| mode | octal <mode> | the permissions to apply to `path` | yes |


#### Container Config
These Stanzas are available when defining containers:

| Parameter | Values | Description | Required |
| --------- | ------ | ----------- | -------- |
| image | `<image_path>` | The path to the image to use for this container.  Can be overriden by -i. | Yes |
| exec  | `<command>` | The command to run inside the container. If left out will run the default container command. | No |
| environment | Hash of environment values `ENV:value` | The environment values to pass into the container | No |
| mounts | See Below | A list of mount definitons for this container. | No |
| state_conditions | See Below | A hash of state conditions to determin success or failure for this container | No |
| depends_on | List of container definition names | The containers that this container depends on. | No |

##### Mounts
Mounts are used to mount folders on host machine into the container.  These stanzas are available when defining mounts:

| Parameters | Values | Description | Required |
| ---------- | ------ | ----------- | -------- |
| volume | `<volume_name>` | The name of the volume (as defined above) to mount | Yes |
| path | `<path>` | the path inside the container to mount `volume` on | 

##### State Conditions
State conditions are used to determin if a container has come up sucesfully or not.  There are several types of state conditions that we support:

###### timeout
This state condition will trigger after a certain amount of time has passed. It expects a hash with the following parameters:

| Parameters | Values | Description | Required |
| ---------- | ------ | ----------- | -------- |
| duration | `<int>` | The number of seconds to wait prior to triggering | Yes |
| status | `success` \| `failure` | The result to return if this state condition triggers | Yes

###### exit
This state condition will trigger when the container exits.   Note that, if no exit state condition is defined then the container is not expected to exit, and any exit will trigger a failure. It expects a hash with the following parameters:

| Parameters | Values | Description | Required |
| ---------- | ------ | ----------- | -------- |
| codes | `[ <int> ]` | Expects an array of exit codes.  These are the codes that will trigger the result defined in `status` | Yes |
| status | `success` \| `failure` | The result to return if the container exits with one of `codes`.  If any other exit code is returned the other status will be returned. | Yes |

###### output
This state condition will monitor the output of the container and trigger if a Regex is found.  It expects a list of hashes containing the following parameters: 

| Parameters | Values | Description | Required |
| ---------- | ------ | ----------- | -------- |
| source | `STDOUT`\|`STDERR` | The output to monitor | Yes |
| regex | /regex/ | The regular expression to monitor the `source` for | Yes |
| status | `success`\|`failure` | The status to return when `regex` is found | Yes |

###### filemonitor
This state condition will monitor the named files for the supplied Regex, and trigger if it is found.  Note that monitoring is done externally to the container, so any files must be exported via `mounts` and `volumes`.  It expects a list of hashes containting the following parameters:

| Parameters | Values | Description | Required |
| ---------- | ------ | ----------- | -------- |
| file | `<file_path>` | The file path to monitor.  **Note that this is the path to the file inside the container**. | Yes |
| regex | /regex/ | The regex to look for | Yes |
| status | `success`\|`failuire` | The status to return when `regex` is found | Yes |

### Full Config Example
This is an example of how to use all of the above config stanzas.

File 1: `/data/constellation_files/postgres.yml`
```yaml
containers:
  db.local:
    image: docker://postgres:9.6
    environment:
      POSTGRES_USER: appuser
      POSTGRES_DB: scratch
      POSTGRES_PASSWORD: password
    state_conditions:
      output:
        - source: STDOUT
          regex: PostgreSQL init process complete; ready for start up.
          status: success
```
File 2: `/data/api/api.yml`
```yaml
require:
  - postgres.yml
volumes:
  app-log-dir:
    kind: host
    path: /tmp/app-logs
    uid: 9998
    gid: 9998
    mode: 0755
containers:
  api.app.local:
    image: aci-repo.example.com/api:af457b220597aa34b739bff13afc514ba72e8100
    exec: envar /opt/jdk/bin/java -jar /opt/api/api.jar server /etc/api-config.yml
    environment:
      DATABASE: db.local
    mounts:
      - volume: app-log-dir
        path: /var/log/api
    state_conditions:
      filemonitor:
        - file: /var/log/api/api_application.log
          regex: "org.eclipse.jetty.server.Server | Started"
          status: success
        - file: /var/log/api/api_application.log
          regex: ERROR
          status: failure
      timeout:
        duration: 300
        status: failure
      exit:
        codes [ 1, 2 ]
        status: failure
      output:
        - source: STDOUT
          regex: Running .* application@[a-z0-9]*
          status: success
        - source: STDERR
          regex: Exited
          status: failure
    depends_on:
      - db.local
```

This would then be run with the following invocation:
`sudo constellation run -c /data/api/api.yml -I /data/constellation_files`


# Known Bugs
- Setting a Volume with Kind to "empty" will result in constellation not being able to monitor the files referenced in those volumes, 
  but also not reporting an error
- The "clean" command does not always remove all containers in a single run.  Multiple runs will fix this for now.
- When a project contains transient containers (containers that are expected to exit) re-running the project without running "clean" first will attempt to re-run those transient containers.  If they make changes that are not idempotent (e.g. db migrations etc) they are likely to fail.  I have a fix for this, but have not had a chance to implement it yet.

# TODO
- Clean up output - its a bit too verbose
