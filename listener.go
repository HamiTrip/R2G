package main

import (
	"fmt"
	"encoding/json"
	"os/exec"
	"os"
	"bufio"
)

func RunCommand(db *DbMap) {
	var cmd  *exec.Cmd
	mysql := SystemConfig.Database.Mysql
	if SystemConfig.Debug {
		var host = ""
		if mysql.Host == "127.0.0.1" {
			host = "172.17.0.1"
		} else {
			host = mysql.Host
		}
		cmd = exec.Command(
			"docker",
			"run",
			"osheroff/maxwell",
			"bin/maxwell",
			"--user=" + mysql.Username,
			"--password=" + mysql.Password,
			"--host=" + host,
			"--producer=stdout")
	} else {
		cmd = exec.Command(
			"/app/bin/maxwell",
			"--user=" + mysql.Username,
			"--password=" + mysql.Password,
			"--host=" + mysql.Host,
			"--producer=stdout")
	}

	cmdReader, err := cmd.StdoutPipe()

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			go handelData(scanner.Text(), db)
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
		os.Exit(1)
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
		os.Exit(1)
	}
}

//func StartChangeListener(db *DbMap) {
//
//	ln, err := net.Listen("tcp", ":8080")
//	defer ln.Close()
//
//	if err != nil {
//		// handle error
//	}
//
//	fmt.Println("Start listen over... :8080")
//
//	for {
//		conn, err := ln.Accept()
//		if err != nil {
//			// handle error
//		}
//		go handleConnection(conn, db)
//	}
//}
//
//func handleConnection(conn net.Conn, db *DbMap) {
//	ch := make(chan []byte)
//	eCh := make(chan error)
//
//	// Start a goroutine to read from our net connection
//	go func(ch chan []byte, eCh chan error) {
//		for {
//			// try to read the data
//			data := make([]byte, 512)
//			data_len, err := conn.Read(data)
//			if err != nil {
//				// send an error if it's encountered
//				eCh <- err
//				return
//			}
//			// send data if we read some.
//			ch <- data[:data_len]
//		}
//	}(ch, eCh)
//
//	ticker := time.Tick(time.Second)
//	// continuously read from the connection
//	for {
//		select {
//
//		case data := <-ch:
//			go handelData(data, db)
//
//		case err := <-eCh:
//			fmt.Println(err)
//			break;
//
//		case <-ticker:
//
//		}
//	}
//
//}


////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type DatabaseChange struct {
	Database       string
	Table          string
	TableStructure *Table
	Type           string
	Data           map[string]interface{}
	Old            map[string]interface{}
}

func handelData(data string, db *DbMap) {
	databaseChange := DatabaseChange{}
	if (DEBUG) {
		fmt.Println(data)
	}
	err := json.Unmarshal([]byte(data), &databaseChange)
	if err != nil {
		fmt.Println("cannot decode json !")
		fmt.Println(err)
		return
	}

	if databaseChange.Database != db.Mysql.DbName {
		return
	}

	if stringInSlice(databaseChange.Table, SystemConfig.SkipTables) {
		return
	}

	databaseChange.SetTableStructure(db)

	var query string

	switch databaseChange.Type {
	case "insert":
		query = databaseChange.createQuery()
		break
	case "update":
		query = databaseChange.updateQuery()
		break
	case "delete":
		query = databaseChange.deleteQuery()
		break
	}

	if (DEBUG) {
		fmt.Println(query)
	}

	Mutex.Lock()
	_, err = db.Graph.Conn.ExecNeo(query, make(map[string]interface{}))
	Mutex.Unlock()

	if err != nil {
		fmt.Println("Errorr", err)
	}
}

func (databaseChange *DatabaseChange)  SetTableStructure(db *DbMap) {
	databaseChange.TableStructure = db.Mysql.Tables[databaseChange.Table]
}

func (databaseChange *DatabaseChange) createQuery() string {
	var query string
	tableConfig := SystemConfig.GetTableConfig(databaseChange.TableStructure)
	if tableConfig.IsManyToMany {
		query = RelationQuery("nod", databaseChange.TableStructure, databaseChange.GenerateProperties())
	} else {
		query = MergeQuery("nod", databaseChange.TableStructure, databaseChange.GenerateProperties(), true)
	}
	return query
}

func (databaseChange *DatabaseChange) updateQuery() string {
	var query string
	tableConfig := SystemConfig.GetTableConfig(databaseChange.TableStructure)

	set, property := databaseChange.GetSetANdProperty("nod")
	fmt.Println("---------------------------------------------")
	fmt.Println(set)
	fmt.Println("==================================")
	fmt.Println(property)
	fmt.Println("---------------------------------------------")
	label := "nod"

	if tableConfig.IsManyToMany {
		if len(databaseChange.TableStructure.foreignKeys) != 2 {
			panic("ManyToMany tabels must has been only 2 foreign key !")
		}

		properties := databaseChange.GenerateProperties()
		sets, property := databaseChange.TableStructure.GetSetAndProperty(label, properties)

		table_from := databaseChange.TableStructure.foreignKeys[0]
		label_from := "nod" + RandStringRunes(5)
		rows_from := table_from.ReferenceTable.GetFilteredRows(table_from.ReferenceColumnName + "='" + properties[table_from.ColumnName] + "'", -1, -1)
		node_from := MergeQuery(label_from, table_from.ReferenceTable, rows_from[0], false)

		table_to := databaseChange.TableStructure.foreignKeys[1]
		label_to := "nod" + RandStringRunes(5)
		rows_to := table_to.ReferenceTable.GetFilteredRows(table_to.ReferenceColumnName + "='" + properties[table_to.ColumnName] + "'", -1, -1)
		node_to := MergeQuery(label_to, table_to.ReferenceTable, rows_to[0], false)

		updated_keys := []string{}
		for k := range databaseChange.Old {
			updated_keys = append(updated_keys, k)
		}
		if stringInSlice(table_from.ColumnName, updated_keys) || stringInSlice(table_to.ColumnName, updated_keys) {
			fmt.Println("change in relation must be design")
		}

		query := node_from
		query += "\n" + node_to + "\n"
		query += "MERGE (" + label_from + ")-[" + label + ":" + databaseChange.TableStructure.GetTag()
		query += property + " ]-(" + label_to + ")"
		if sets != "" {
			query += " SET " + sets
		}

	} else {
		query = "MERGE (" + label + ":" + databaseChange.TableStructure.GetTag() + " {" + property + "})"
		if set != "" {
			query += " SET" + set
		}
		if len(databaseChange.TableStructure.GetForeignKeys()) > 0 {
			HandleRelation(label, &query, databaseChange.TableStructure, databaseChange.GenerateProperties())
		}

	}

	return query
}

func (databaseChange *DatabaseChange) deleteQuery() string {

	tableConfig := SystemConfig.GetTableConfig(databaseChange.TableStructure)

	var query string
	property := " "
	for key, value := range databaseChange.Data {
		if databaseChange.TableStructure.IsSkipProperty(key) {
			continue
		}

		property += " n." + key + "='" + FixStringStyle(GetValue(value)) + "' AND"

	}
	if len(property) == 1 {
		return "RETURN Null"
	}

	if tableConfig.IsManyToMany {
		query = "MATCH ()-[n:" + databaseChange.TableStructure.GetTag() + "]-() WHERE " + property[:len(property) - 3] + " DELETE r"
	} else {
		query = "MATCH (n:" + databaseChange.TableStructure.GetTag() + ") WHERE " + property[:len(property) - 3] + " DETACH DELETE n"
	}

	return query
}

func (databaseChange *DatabaseChange) GetSetANdProperty(label string) (string, string) {
	set := " "
	property := " "

	if databaseChange.Type != "update" {
		for key, value := range databaseChange.Data {
			if databaseChange.TableStructure.IsSkipProperty(key) {
				continue
			}

			if databaseChange.TableStructure.IsUniqueProperty(key) {
				property += " " + key + ":'" + FixStringStyle(GetValue(value)) + "' ,"
			} else {
				set += " " + label + "." + key + "='" + FixStringStyle(GetValue(value)) + "',"
			}
		}

	} else {
		for key, value := range databaseChange.Data {
			if databaseChange.TableStructure.IsSkipProperty(key) {
				continue
			}
			set += " " + label + "." + key + "='" + FixStringStyle(GetValue(value)) + "',"

			if databaseChange.TableStructure.IsUniqueProperty(key) {
				if val, ok := databaseChange.Old[key]; ok {
					property += " " + key + ":'" + FixStringStyle(GetValue(val)) + "' ,"
				} else {
					property += " " + key + ":'" + FixStringStyle(GetValue(value)) + "' ,"
				}
			}
		}

	}

	return set[:len(set) - 1], property[:len(property) - 1]
}

func (databaseChange *DatabaseChange) GenerateProperties() map[string]string {
	properties := map[string]string{}

	for key, val := range databaseChange.Data {
		properties[key] = GetValue(val)
	}

	return properties
}