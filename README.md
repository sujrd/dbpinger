# Introduction

The dbPinger is a simple daemon that checks the health of Galera cluster nodes. A load balancer can then query the dbPinger using simple HTTP GET requests to check the health of the Galera node.

## Usage

Create a configuration file that contains the following:

```
[main]
listen = 4146
dbhost = localhost
dbport = 3306
dbuser = <user>
dbpass = <pass>
```

then run the daemon with:

```
dbpinger -c file.conf
```

After the daemon is running you can configure your load balancer (e.g. Amazon ELB) to query the following:

```
http://<db server>:4146/ping
```

this returns 200 OK if the database is running and the Galera node is in sync and part of a cluster.

