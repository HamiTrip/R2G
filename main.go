package main

import (
	"fmt"
	"gopkg.in/cheggaaa/pb.v1"
	"strconv"
	"math/rand"
	"os"
	"strings"
	"time"
)

const DEBUG = true

var ARGS_REQUESTED_TABLES = []string{}

func main() {
	InitConfiguration()
	db := ConnectDatabase()
	defer db.Close()
	//
	db.Init()
	SystemConfig.InitDbConfig(db.Mysql.Tables)
	//
	args := []string{}
	if len(os.Args) > 1 {
		args = os.Args[1:]
		for _, arg := range args {
			if strings.HasPrefix(arg, "tables=") {
				data := strings.Replace(arg, "tables=", "", 1)
				for _, value := range strings.Split(data, ",") {
					ARGS_REQUESTED_TABLES = append(ARGS_REQUESTED_TABLES, value)
				}
			}
		}

	} else {
		args = []string{"Initial", "Listener"}
	}

	if stringInSlice("Initial", args) {
		if stringInSlice("Listener", args) {
			go R2G(db)
		} else {
			R2G(db)
		}
	}

	if stringInSlice("Listener", args) {
		RunCommand(db)
	}
	time.Sleep(10000000000000)
}

func R2G(db *DbMap) {

	fmt.Println("****************************************************************************")
	fmt.Println("Start to create Indexes")

	bar := pb.StartNew(len(db.Mysql.Tables))
	bar.ShowBar = true
	bar.SetWidth(80)
	for _, table := range db.Mysql.Tables {

		bar.Increment()
		if stringInSlice(table.name, SystemConfig.SkipTables) || SystemConfig.GetTableConfig(table).IsManyToMany {
			continue
		}

		createIndex(db, table)
	}
	bar.FinishPrint("End of create tables index ")

	if len(ARGS_REQUESTED_TABLES) == 0 {
		for _, table := range db.Mysql.Tables {
			if stringInSlice(table.name, SystemConfig.SkipTables) {
				continue
			}
			insertData(db, table)
		}
	} else {
		for _, table := range db.Mysql.Tables {
			if stringInSlice(table.name, SystemConfig.SkipTables) {
				continue
			}

			if stringInSlice(table.name, ARGS_REQUESTED_TABLES) {
				insertData(db, table)
			}
		}
	}

}

func createIndex(dbmap *DbMap, table *Table) {
	queries := []string{}
	for _, column := range table.GetUniqueProperties() {
		queries = append(queries, "CREATE  INDEX ON :" + table.GetTag() + "(" + column + ")")
		//queries = append(queries, "CREATE INDEX ON :" + table.GetTag() + "(" + column + ")")
	}
	_, err := dbmap.Graph.Conn.ExecPipeline(queries, make([]map[string]interface{}, len(queries))...)
	if err != nil {
		fmt.Println(queries[0])
		panic(err)
	}
}

func insertData(dbmap *DbMap, table *Table) {
	paginated := table.getPaginated(SystemConfig.Limit)

	fmt.Println("****************************************************************************")
	fmt.Println("Start to insert Table " + table.name + " | number of rows:" + strconv.Itoa(paginated.TotalItems))

	bar := pb.StartNew(paginated.TotalPages)
	bar.ShowBar = true
	bar.SetWidth(80)
	for paginated.Next() {
		rows := table.GetRows(paginated.getLimitOffset())
		saveToGraph(dbmap, table, rows, true)
		bar.Increment()
	}
	bar.FinishPrint("End of insert Table " + table.name)
}

func saveToGraph(dbmap *DbMap, table *Table, rows []map[string]string, frist_try bool) {
	tableConfig := SystemConfig.GetTableConfig(table)
	if tableConfig.IsManyToMany {
		queries := []string{}
		for index, row := range rows {
			label := "nod" + strconv.Itoa(index)
			queries = append(queries, RelationQuery(label, table, row))
		}
		_, err := dbmap.Graph.Conn.ExecPipeline(queries, make([]map[string]interface{}, len(queries))...)
		if err != nil {
			fmt.Println("Samaple query :" + queries[0])
			fmt.Println("Error Wile Insert Retry ")
			if frist_try {
				fmt.Println(err)
				fmt.Println("Start to retry !")
				saveToGraph(dbmap, table, rows, false)
				return
			}
			panic(err)
		}
	} else {
		queries := []string{}
		for index, row := range rows {
			label := "nod" + strconv.Itoa(index)
			queries = append(queries, MergeQuery(label, table, row, true))
		}
		_, err := dbmap.Graph.Conn.ExecPipeline(queries, make([]map[string]interface{}, len(queries))...)
		if err != nil {
			fmt.Println("Samaple query :" + queries[0])
			fmt.Println("Error Wile Insert Retry ")
			if frist_try {
				fmt.Println(err)
				fmt.Println("Start to retry !")
				saveToGraph(dbmap, table, rows, false)
				return
			}
			panic(err)
		}
	}

}

func RelationQuery(label string, table *Table, properties map[string]string) string {
	if len(table.foreignKeys) != 2 {
		panic("ManyToMany tabels must has been only 2 foreign key !")
	}

	sets, property := table.GetSetAndProperty(label, properties)

	table_from := table.foreignKeys[0]
	label_from := "nod" + RandStringRunes(5)
	rows_from := table_from.ReferenceTable.GetFilteredRows(table_from.ReferenceColumnName + "='" + properties[table_from.ColumnName] + "'", -1, -1)
	node_from := MergeQuery(label_from, table_from.ReferenceTable, rows_from[0], false)

	table_to := table.foreignKeys[1]
	label_to := "nod" + RandStringRunes(5)
	rows_to := table_to.ReferenceTable.GetFilteredRows(table_to.ReferenceColumnName + "='" + properties[table_to.ColumnName] + "'", -1, -1)
	node_to := MergeQuery(label_to, table_to.ReferenceTable, rows_to[0], false)

	query := node_from
	query += "\n" + node_to + "\n"
	query += "MERGE (" + label_from + ")-[" + label + ":" + table.GetTag()
	query += "{ " + property + "} ]-(" + label_to + ")"
	if sets != "" {
		query += " SET " + sets
	}
	return query
}

func MergeQuery(label string, table *Table, properties map[string]string, insert_relation bool) string {

	sets, property := table.GetSetAndProperty(label, properties)

	data := "MERGE (" + label + ":" + table.GetTag() + " {" + property + "})"
	if sets != "" {
		data += " SET" + sets
	}
	if insert_relation && len(table.GetForeignKeys()) > 0 {
		HandleRelation(label, &data, table, properties)
	}

	return data
}

func HandleRelation(parent_label string, data *string, table *Table, properties map[string]string) {

	for _, relation := range table.GetForeignKeys() {
		rows := relation.ReferenceTable.GetFilteredRows(relation.ReferenceColumnName + "='" + properties[relation.ColumnName] + "'", -1, -1)

		for index, value := range rows {
			label := "for" + RandStringRunes(5) + strconv.Itoa(index)
			relLabel := "for" + RandStringRunes(5) + strconv.Itoa(index)

			set, property := relation.GetRelationSetAndProperty(relLabel, properties, value)
			*data += "\n" + MergeQuery(label, relation.ReferenceTable, value, false)
			*data += "\n" + "MERGE (" + parent_label + ")<-[ " + relLabel + ":" +
				relation.GetTag() + " " + property + " ]-(" + label + ") "
			if set != "" {
				*data += " SET " + set + " "
			}
		}

	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

