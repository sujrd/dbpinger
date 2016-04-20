# Introduction

The dbPinger is a simple daemon that checks the health of Galera cluster nodes. A load balancer can then query the dbPinger using simple HTTP GET requests to check the health of the Galera node.

Usually to check the health of a database node in a master-slave or master-master configuration a load balancer simply tries to open a TCP connection to the database and perform a simple SELECT query. If the query returns a result the load balancer assumes the database node is healthy and forwards requests to it.

Unfortunately for a Galera cluster node a simple SELECT query is not enough to assert that the node is healthy. It is possible that the database node is accessible and that it responds to queries but still be disconnected from the cluster, be out of sync etc. 

This dbPinger daemon performs additional checks to ensure that the cluster node is alive and that it actually belongs to the cluster. Only when all these checks pass the node is considered healthy.

The checks the dbPinger perform currently are:

  - Ensure wsrep_cluster_status equals "Primary"
  - Ensure wsrep_connected is "ON"
  - Ensure wsrep_ready is "ON"
  - Ensure wsrep_local_state equals 4
  - Ensure that wsrep_cluster_state_uuid equals wsrep_local_state_uuid.

## Compilation

This dbPinger utility is written in Go so you need a Go development environment in order to compile it. If you are using Ubuntu/Debian is as simple as:

    sudo apt-get install golang

Go uses a standard workspace to develop all libraries and applications. In order to build and run the dbPinger daemon we must first create such workspace on our machines:

```
mkdir $HOME/go
mkdir $HOME/go/src
mkdir $HOME/go/bin
mkdir $HOME/go/pkg
mkdir $HOME/go/src/git.skillupjapan.net  # Here we keep all R&D go source files.
export GOPATH=$HOME/go                   # Choose any place you prefer
export PATH=$GOPATH/bin:$PATH
```````

You may want to add the GOPATH and PATH variables to your bash_profiles (MAC) or bashrc (Linux). Afterwards we need to install dependencies:

    go get github.com/go-sql-driver/mysql
    go get gopkg.in/gcfg.v1

finally we can build the dbPinger by downloading the source code and compiling it with:

    go build dbpinger

## Installation

Create a configuration file that contains the following:

```
[main]
listen = "4146"       # Listening port the load balancer uses to request health check
dbhost = "localhost"  # Cluster node connection host.
dbport = "3306"       # Cluster node port.
dbuser = "<user>"     # Cluster node username.
dbpass = "<pass>"     # Cluster node password.
```

then run the daemon with:

```
dbpinger -c file.conf
```

##  Systemd Daemon

If you want the dbPinger to run as a SystemD daemon simply create a file */etc/init/dbpinger.conf* that contains the following:

```
description "DbPinger Galera Health Checker"
author "rdadmin@allm.net"

start on started mountall
stop on shutdown

respawn

setuid mysql

exec /sbin/dbpinger -c /etc/dbpinger.conf
```````

This will allow the dbPinger to run as a daemon and be restarted when the database node restarts.

## Verify

You can check if dbPinger is running by making an HTTP GET request to the dbPinger listening port:

```
curl -v http://localhost:4146/ping
* Hostname was NOT found in DNS cache
*   Trying 127.0.0.1...
* Connected to localhost (127.0.0.1) port 4146 (#0)
> GET /ping HTTP/1.1
> User-Agent: curl/7.35.0
> Host: localhost:4146
> Accept: */*
> 
< HTTP/1.1 200 OK
< Date: Wed, 25 Mar 2015 10:44:57 GMT
< Content-Length: 0
< Content-Type: text/plain; charset=utf-8
< 
* Connection #0 to host localhost left intact
```

if the command returns 200 OK it means the database is running and the Galera node is in sync and part of a cluster. Any other HTTP response code indicates an issue with the configuration of Galera or of dbPinger itself.

## Configuring Amazon ELB

You can use an internal ELB to load balance the database connections to the nodes in your Galera cluster. Simply add the following URL as health check:

```
http://<db server>:<port>/ping
```

replace <db server> with the IP of the machine were dbPinger is running and <port> with the value you set for *listen* in the dbpinger.conf file.

## Configuring HAProxy

When using HAProxy to load balance database connections you can use a configuration like this:

    listen fup-db localhost:3306
    mode tcp
    balance leastconn
    option tcplog
      option httpchk GET /ping
      server db01 192.168.11.3:3306 maxconn 100 check port 4146 backup
      server db02 192.168.11.4:3306 maxconn 100 check port 4146 backup
      server db03 192.168.11.5:3306 maxconn 100 check port 4146

this configuration will use HTTP GET checks to the /ping path on port 4146 on each of the servers in the cluster. If any of these checks returns a response other that 200 then that node is removed from the balancer set.
