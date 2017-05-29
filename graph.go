package main

import (
	"github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

type Neo4jDbMap struct {
	Conn golangNeo4jBoltDriver.Conn
}

func Neo4jConnect() *Neo4jDbMap {
	graph := SystemConfig.Database.Graph
	driver := golangNeo4jBoltDriver.NewDriver()
	conn, err := driver.OpenNeo("bolt://" + graph.Username + ":" + graph.Password + "@" + graph.Host + ":" + graph.Port)

	if err != nil {
		panic(SystemConfig.Database.Graph.Host + err.Error())
	}

	return &Neo4jDbMap{Conn:conn}
}

func (dbmap *Neo4jDbMap) Close() {
	dbmap.Conn.Close()
}

