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
    
    
## How to customize configuration:

#### TablesConfig:

For example (Not for many to many tables is relation):

    "TablesConfig": [
        {
          "name": "continent",
          "label": "CNT",
          "UniqueColumns": [
            "hid"
          ],
          "SkipColumns": [
            "id",
            "updated_at",
            "created_at"
          ]
        }
    ]
    
**name**: Mysql table name
**label**: Neo4j ***node*** label
**UniqueColumns**: The unique columns (set index or this properties in neo4j)
**SkipColumns**: skip this columns and don't save to node properties


-----------------------------------------------------------------
 
 If you want set many-to-many tables as relation in neo4j you can use the following config for table:
  
    "TablesConfig": [
        {
          "name": "user_orders",
          "label": "ORD_USR",
          "IsManyToMany": true,
          "UniqueColumns": [
            "id"
          ],
          "SkipColumns": [
            "updated_at",
            "created_at"
          ]
        }
    ]

this configuration is mostly like the first example.

**name**: Mysql table name
**label**: Neo4j ***relation*** label
**IsManyToMany**: set this table is many-to-many troth table
**UniqueColumns**: The unique columns (set index or this properties in neo4j)
**SkipColumns**: skip this columns and don't save to node properties


### RelationsConfig

For example (without relation properties):

    "RelationsConfig": [
        {
          "Label": "CTN_CTR",
          "Table": "country",
          "ReferenceTable": "continent"
        }
    ]

This method find relations from mysql foreign keys and you can customize the relations.

**Label**: Neo4j ***relation*** label
**Table**: The mysql table
**ReferenceTable**: The mysql reference table


-------------------------------------------------------------------

example with relation properties:

you can use table or reference table columns value as relation values

    "RelationsConfig": [
        {
          "Label": "CNT_CTR",
          "Table": "country",
          "ReferenceTable": "continent",
          "Properties": {
            "title": "country.title"
        }
    ]



