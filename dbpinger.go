package main

import (
  "fmt"
  "log"
	"os"
  "database/sql"
  "net/http"
  "flag"
  _ "github.com/go-sql-driver/mysql"
  "code.google.com/p/gcfg"
)

//Place here any values that you want to check
var dbStatusCheck = map[string]string {
//  "wsrep_cluster_state_uuid" This two things are hardcoded
//  "wsrep_local_state_uuid"
  "wsrep_cluster_status" : "Primary",
  "wsrep_connected" : "ON",
  "wsrep_ready" : "ON",
  "wsrep_local_state" : "4",
}

type Config struct {
  Main struct {
  	Listen string
		DBHost string
		DBPort string
		DBUser string
		DBPass string
	}
}

var confFile string
var cfg Config

func main() {

  flag.StringVar(&confFile, "c", "/etc/dbpinger.conf", "Galera health status checker")

  flag.Parse()

  err := gcfg.ReadFileInto(&cfg, confFile)

  if err != nil {
    fmt.Println("Failed to load config file " + confFile);
		fmt.Println(err)
		os.Exit(1)
  }
	
  http.HandleFunc("/ping", pingHandler)

  log.Println(http.ListenAndServe(":"+ cfg.Main.Listen, nil))
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
     if checkDB(cfg.Main.DBHost, cfg.Main.DBPort, cfg.Main.DBUser, cfg.Main.DBPass, "") {
       w.WriteHeader(200)
     } else {
       w.WriteHeader(500)
     }
}

func checkDB(dbhost, dbport, dbuser, dbpass, dbname string) bool {

  db,err := newDBConnection(dbhost, dbport, dbuser, dbpass, dbname)
  defer db.Close()

  if err != nil {
    log.Println("Connection failed !")
    return false
  }

  rows,err := db.Query("SHOW STATUS")

  if err != nil {
    log.Println("Query failed")
    return false
  }

  tempMap := map[string]int{} 

  for k,_ := range dbStatusCheck {
    tempMap[k] = 1
  }

  wsrep_cluster_state_uuid := "."
  wsrep_local_state_uuid := "-"

  defer rows.Close()
  for rows.Next() {
      var name string
      var value string
      err = rows.Scan(&name, &value)

      if err != nil {
        log.Println("Get Rows failed !")
        return false
      } 

      if !checkValue(name, value) {
        log.Println("Error: " + name + ": " + value)
        return false
      }

      if tempMap[name] != 0 {
        tempMap[name] = 2
      }

      if name == "wsrep_cluster_state_uuid" {
        wsrep_cluster_state_uuid = value
      }

      if name == "wsrep_local_state_uuid" {
        wsrep_local_state_uuid = value
      }
  }

  if wsrep_local_state_uuid != wsrep_cluster_state_uuid {
    log.Println("Error: wsrep_local_state_uuid and wsrep_cluster_state_uuid don't match or are missing")
    return false

  }

  for k,v := range tempMap {
    if v != 2 {
      log.Println("Error: " + k + " is missing")
      return false
    }
  }

  return true
}

func checkValue(name,value string) bool {

    //If is not in the list, we just ignore it (true)
    if(dbStatusCheck[name] == "") {
      return true
    }
    return dbStatusCheck[name] == value
}

//You must close the opened connection! 
func newDBConnection(dbhost, dbport, dbuser, dbpass, dbname string) (*sql.DB, error) {

  key := buildKey(dbhost, dbport, dbuser, dbpass, dbname)

  newDB, err := sql.Open("mysql", key + "?timeout=15s&wait_timeout=5000")

  if err != nil {
    return newDB, err
  }

  return newDB, nil
}

func buildKey(host, port, user, pass, name string) string {
  return fmt.Sprintf("%v:%v@tcp(%v:%v)/%v", user, pass, host, port, name)
}
