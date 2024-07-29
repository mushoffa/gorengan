# Liquibase
This particular package does not have a Go code base, rather than to demonstrate database migration using Liquibase alongside your Go project. The configuration file(s) and test(s) were performed on the Ubuntu operating system, please adjust accordingly to your OS.

## Official Documentation
* [Docs](https://docs.liquibase.com/start/home.html)
* [Getting Started](https://www.liquibase.org/get-started/quickstart)
* [Docker Image](https://hub.docker.com/r/liquibase/liquibase)

## Prerequisite
* Please ensure you have [Docker](https://docs.docker.com/engine/install/) installed on your system.
* Please ensure you have database install on your system (PostgreSQL, MySQL, MongoDB, etc).

## Run Migration
Please adjust your database environment variables accordingly before running the migration process (database url, username, password, port, etc). Run migration process using defined rule(s) on [Makefile](./Makefile) or you can write your own script.
### Localhost
This case is when you have your database installed on your system. Run the following makefile rule on your terminal to run the migration.
```shell
$ make migrate-local
```
> [!WARNING]
> This case has not been tested yet; please resolve any errors that occurred throughout the process on your end.
### Localhost - Docker
This case is when you have your database installed as docker container on your system. 

1. Please ensure your database container name.
```shell
$ docker ps
CONTAINER ID   IMAGE      COMMAND                  CREATED      STATUS        PORTS                                       NAMES
ec1a74c8676d   postgres   "docker-entrypoint.s…"   5 days ago   Up 19 hours   0.0.0.0:5440->5432/tcp, :::5440->5432/tcp   pg-projects
```

2. Please ensure your docker container network.
```shell
$ docker container inspect pg-projects | grep NetworkMode
            "NetworkMode": "postgres_default",
```

3. Adjust your container name and its network to the docker command, in this case my container name is ```pg-projects``` and network name is ```postgres_default```.
```shell
docker run --network=postgres_default --rm \
		-v $(ROOT_PATH)/migration:/liquibase/changelog liquibase/liquibase \
		--searchPath=/liquibase/changelog/ \
		--changeLogFile=changelog.xml \
		--url="jdbc:postgresql://pg-projects/test2?user=admin&password=admin&currentSchema=public&sslmode=disable" \
		update
```

4. Run the following makefile rule on your terminal to run the migration.
```shell
$ make migrate-docker
docker run --network=postgres_default --rm \        
        -v /media/mushoffa/DATA/1_REPOSITORIES/2_GITHUB/mushoffa/gorengan/liquibase/migration:/liquibase/changelog liquibase/liquibase \
        --searchPath=/liquibase/changelog/ \                                                           
        --changeLogFile=changelog.xml \
        --url="jdbc:postgresql://pg-projects/liquibase?user=admin&password=admin&currentSchema=public&sslmode=disable" \
        update
####################################################
##   _     _             _ _                      ##
##  | |   (_)           (_) |                     ##
##  | |    _  __ _ _   _ _| |__   __ _ ___  ___   ##
##  | |   | |/ _` | | | | | '_ \ / _` / __|/ _ \  ##
##  | |___| | (_| | |_| | | |_) | (_| \__ \  __/  ##
##  \_____/_|\__, |\__,_|_|_.__/ \__,_|___/\___|  ##
##              | |                               ##
##              |_|                               ##   
##                                                ## 
##  Get documentation at docs.liquibase.com       ##                                                                                                                                                                             
##  Get certified courses at learn.liquibase.com  ## 
##                                                ##
####################################################
Starting Liquibase at 19:36:09 using Java 17.0.12 (version 4.29.0 #3278 built at 2024-07-24 22:00+0000)
Liquibase Version: 4.29.0
Liquibase Open Source 4.29.0 by Liquibase
Running Changeset: 001.xml::001::Liquibase
Running Changeset: 002.xml::002::Liquibase
Running Changeset: 003.xml::002::Liquibase
Running Changeset: 004.xml::004::Liquibase

UPDATE SUMMARY
Run:                          4
Previously run:               0
Filtered out:                 0
-------------------------------
Total change sets:            4

Liquibase: Update has been successful. Rows affected: 4
Liquibase command 'update' was executed successfully.
```
### CI
In my opinion, i think it's possible to run the migration process during one of the CI job(s) or step(s) to your development, statging, or production database environment url.