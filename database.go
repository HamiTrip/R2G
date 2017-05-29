package main

type DbMap struct {
	Mysql *MysqlDbMap
	Graph *Neo4jDbMap
}

func ConnectDatabase() *DbMap {

	mysqlDbMap := ConnectMysql()
	neo4jDbMap := Neo4jConnect()
	dbmap := &DbMap{Mysql:mysqlDbMap,Graph:neo4jDbMap}

	return dbmap
}



func (db *DbMap) Close()  {
	db.Mysql.Close()
	db.Graph.Close()
}

func (db *DbMap) Init()  {
	for _, table := range db.Mysql.GetTables() {
		table.GetColumns()
	}
}

