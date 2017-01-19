# constellation
Constellatin is a tool to spin up a "constellation" of rkt pods (see what I did there) in a controlled fashion.   It allows you to specify a list of containers to spin up, their success or failure conditions, and their interdependencies such that dependant containers are not spun up until the containers they depend on have encountered a "success" condition.  Constellation will also ensure that networking is set up between the containers so that dependent containers can talk to their dependencies.

# Examples
These examples go in ascending order of complexity.
## A Simple Application
The simplest invocation would spin up a single application and use the default command baked into the container:
```
traitify-api.app.local:
    image: aci-repo.prod.awse.traitify.com/traitify-api:af457b220597aa34b739bff13afc514ba72e8100
```
You would then run this using `sudo ./constellation run -c api.yml` and a single container would be spun up.

## A Simple Applications with some conditions
Lets add in some success and failure conditions
```
containers:
  api.app.local:
    image: aci-repo.prod.awse.traitify.com/traitify-api:af457b220597aa34b739bff13afc514ba72e8100
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
This tells constelation to run the container with the default command, but also checks STDOUT for the string `The server is now ready to accept connections`, and STDERR for the string `ERROR`.  If it finds the STDOUT string, the container will be marked as having started succesfully, if it finds the STDERR string, it will be marked as having failed.   Additionally, there is a timeout - if no other state_condition has occured after 300 (seconds), then the container will be marked as having failed.   Note that the first state_condition to occur stops monitoring for other state conditions.  So once a success or failure condition has happened, no other conditions will change that.

## An application and its database
Now lets say our application requries a database. We can set that up as follows:
```
containers:
  db.local:
    image: docker://postgres:9.6
    environment:
      POSTGRES_USER: appuser
      POSTGRES_DB: scratch
      POSTGRES_PASSWORD: dev_password_1234
    state_conditions:
      output:
        - source: STDOUT
          regex: PostgreSQL init process complete; ready for start up.
          status: success
      timeout:
        duration: 30
        status: failure
  api.app.local:
    image: aci-repo.prod.awse.traitify.com/traitify-api:af457b220597aa34b739bff13afc514ba72e8100
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
```
containers:
  db.local:
    image: docker://postgres:9.6
    environment:
      POSTGRES_USER: appuser
      POSTGRES_DB: scratch
      POSTGRES_PASSWORD: dev_password_1234
    state_conditions:
      output:
        - source: STDOUT
          regex: PostgreSQL init process complete; ready for start up.
          status: success
      timeout:
        duration: 30
        status: failure
  api.migrate.tmp:
    image: aci-repo.prod.awse.traitify.com/traitify-api:af457b220597aa34b739bff13afc514ba72e8100
    exec: /opt/jdk/bin/java -jar /opt/api/api.jar db migrate /etc/traitify/api-config.yml
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
    image: aci-repo.prod.awse.traitify.com/traitify-api:af457b220597aa34b739bff13afc514ba72e8100
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
This will spin up a postrgres db, and then run migration scripts against it.  Constellation will look for an exit code of `0` from the migration scripts, and if found, will treat that as a sucessful run and will move on to the next container.  Any other exit will result in a failure, and taking longer than 300 seconds will also result in a failure.  Note that we chain our container dependencies so that startup will be in order db.local --> api.migrate.tmp --> api.app.local.  You may use arbitrarily complex dependnecy relationships, however, dependency loops will result in an error.

## Monitoring log files for state_conditions
It is also possible to monitor log files for regex, and use the results in state conditions:
```
volumes:
  app-log-dir:
    kind: host
    path: /tmp/app-logs
    uid: 9998
    gid: 9998
    mode: 0755
    
containers:
  api.app.local:
    image: aci-repo.prod.awse.traitify.com/traitify-api:af457b220597aa34b739bff13afc514ba72e8100
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
This will start the application, and monitor the file `/var/log/api/api_application.log` for success and failure strings.   If it does not find one in 300 seconds it will register the application as having failed.   To understand the above config, it is important to realize that *file monitoring is done by Constellation outside of the container*.   This means that you have to first export the directory that contains the file you want to monitor.   The stanza:
```
mounts:
 - volume: app-log-dir
   path: /var/log/api
```
tells rkt that you want to mount the external volume named `app-log-dir` into the container at the path `/var/log/api`.   Then, the stanzas:
```
volumes:
  app-log-dir:
    kind: host
    path: /tmp/app-logs
    uid: 9998
    gid: 9998
    mode: 0755
```
define the `app-log-dir` volume and tell rkt which external folder to use when the volume `app-log-dir` is referenced in a mount.  Notice that `volumes` are defined constellation-wide and are not specific to a container.  The path in `file` should be the path to the file you want monitored *inside* the container.  Constellation will handle correct path adjustment to find the intended file.

## Spreading config accross multiple files
It is common to have multiple applications that depend on each other.  We don't want to have to keep all of our config together, and we don't want to have to redefine the same config more than once.   Constellation allowes the importing of other constellation configs:
```
require:
  - postgres.yml

containers:           
  api.app.local:
    image: aci-repo.prod.awse.traitify.com/traitify-api:af457b220597aa34b739bff13afc514ba72e8100
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
Constellation can be invoked with the following flags:





# Notes
mention that -v paths must be absolute paths not relative

# Todo
# Known Bugs
- Setting a Volume with Kind to "empty" will result in constellation not being able to monitor the files referenced in those volumes, 
  but also not reporting an error
- The "clean" command does not always remove all containers in a single run.  Multiple runs will fix this for now.
