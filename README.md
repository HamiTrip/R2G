# Relational to Graph
 `This package under development` 
 With this package you can map your relational data from mysql to neo4j graph database 
  
## Features

* automatically map your RDBMS data to graph DB 
* By using observer pattern, keep update your graph data
 
 
## Dependencies 
 * [Maxwell](https://github.com/zendesk/maxwell) (as observer) 
 * Go 1.6+
 
## Run with docker

you must fist create a config json file:

    {
      "Database": {
        "Graph": {
          "Host": "172.17.0.1",
          "Port": "7687",
          "Username": "neo4j",
          "Password": "123456"
        },
        "Mysql": {
          "Host": "172.17.0.1",
          "Port": "3306",
          "DbName": "csp",
          "Username": "username",
          "Password": "password"
        }
      },
      "TablesConfig": [],
      "RelationsConfig": []
    }
    
and then run the following command :
  
    $ docker run -v /path/to/your/conf.json:/app/conf.json amirasaran/r2g 
