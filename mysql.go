package main

import (
	"database/sql"
	"gopkg.in/gorp.v1"
	_ "github.com/ziutek/mymysql/godrv"
	"strconv"
)

type MysqlDbMap struct {
	Conn   *gorp.DbMap
	DbName string
	Tables map[string]*Table
}

func ConnectMysql() *MysqlDbMap {

	mysql := SystemConfig.Database.Mysql
	db, err := sql.Open(
		"mymysql",
		"tcp:" + mysql.Host + ":" + mysql.Port + "*" + mysql.DbName + "/" + mysql.Username + "/" + mysql.Password)
	if err != nil {
		panic(err)
	}

	return &MysqlDbMap{Conn:&gorp.DbMap{Db: db}, DbName:mysql.DbName}
}

func (dbmap *MysqlDbMap) Close() {
	dbmap.Conn.Db.Close()
}

func (dbmap *MysqlDbMap) GetConnection() *gorp.DbMap {
	return dbmap.Conn
}

func (dbmap *MysqlDbMap) GetTables() map[string]*Table {

	if len(dbmap.Tables) > 0 {
		return dbmap.Tables
	}

	tables := []string{}
	dbmap.Tables = make(map[string]*Table)
	dbmap.GetConnection().Select(&tables, "SHOW TABLES")
	for _, table := range tables {
		dbmap.Tables[table] = &Table{dbmap:dbmap, name:table}
	}

	return dbmap.Tables
}

func (dbmap *MysqlDbMap) GetTableColumns(table *Table) ColumnList {

	if table.columns != nil {
		return table.columns
	}

	table.columns = ColumnList{}
	dbmap.GetConnection().Select(&table.columns, "SHOW columns FROM " + table.name)

	foreign_keys := table.GetForeignKeys()

	for _, column := range table.columns {
		column.Table = table
		for _, foreign_key := range foreign_keys {
			if foreign_key.ColumnName == column.Field {
				column._isForeignKey = true
				column.ForeignKey = foreign_key
			}
		}
	}

	return table.columns
}

func (dbmap *MysqlDbMap) GetTableForeignKeys(table *Table) ForeignKeyList {
	foreign_keys := ForeignKeyList{}
	dbmap.GetConnection().Select(&foreign_keys, "SELECT TABLE_NAME `TableName`,COLUMN_NAME `ColumnName`," +
		"CONSTRAINT_NAME `ConstraintName`,REFERENCED_TABLE_NAME `ReferenceTableName`," +
		"REFERENCED_COLUMN_NAME `ReferenceColumnName` FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE " +
		"WHERE TABLE_NAME = '" + table.name + "' AND TABLE_SCHEMA = '" + dbmap.DbName + "'" +
		"AND REFERENCED_TABLE_NAME IS NOT Null")

	for _, foreign_key := range foreign_keys {
		foreign_key.Table = table
		foreign_key.ReferenceTable = dbmap.Tables[foreign_key.ReferenceTableName]
	}

	return foreign_keys
}

func (dbmap *MysqlDbMap) GetRows(table *Table, where string, limit int, offset int) []map[string]string {

	query := "SELECT * FROM `" + table.name + "` "
	if where != "" {
		query += "WHERE " + where
	}

	if limit != -1 {
		if offset != -1 {
			query += " LIMIT " + strconv.Itoa(offset) + ", " + strconv.Itoa(limit)
		} else {
			query += " LIMIT " + strconv.Itoa(limit)
		}
	}

	result, _ := dbmap.Conn.Db.Query(query)

	response := make([]map[string]string, 0)
	for result.Next() {

		params := make([]interface{}, 0)

		data := map[string]*string{}
		data_tamiz := map[string]string{}

		for _, column := range table.columns {
			var temp string
			params = append(params, &temp)
			data[column.Field] = &temp
		}

		result.Scan(params...)

		for key, value := range data {
			data_tamiz[key] = *value
		}
		response = append(response, data_tamiz)

	}
	return response
}

func (dbmap *MysqlDbMap) GetRowsCount(table *Table) int64 {
	count, _ := dbmap.Conn.SelectInt("SELECT Count(*) FROM `" + table.name + "`")
	return count
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Table struct {
	dbmap       *MysqlDbMap
	name        string
	foreignKeys ForeignKeyList
	columns     ColumnList
}

func (table *Table) GetColumns() ColumnList {
	return table.dbmap.GetTableColumns(table)
}

func (table *Table) GetForeignKeys() ForeignKeyList {
	if table.foreignKeys == nil {
		table.foreignKeys = table.dbmap.GetTableForeignKeys(table)
	}
	return table.foreignKeys
}

func (table *Table) GetRows(limit int, offset int) []map[string]string {
	return table.dbmap.GetRows(table, "", limit, offset)
}

func (table *Table) GetRowsCount() int64 {
	return table.dbmap.GetRowsCount(table)
}

func (table *Table) GetFilteredRows(where string, limit int, offset int) []map[string]string {
	return table.dbmap.GetRows(table, where, limit, offset)
}

func (table *Table) getPaginated(limit int) *Paginate {
	return NewPaginated(int(table.GetRowsCount()), limit)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////


type Columns struct {
	Field         string
	Type          string
	Null          string
	Key           string
	Extra         string
	_isForeignKey bool
	Table         *Table
	ForeignKey    *ForeignKey
}

type ColumnList []*Columns

func (columns ColumnList) GetPrimaryColumn() *Columns {
	for _, column := range columns {
		if column.IsPrimary() {
			return column
		}
	}
	return nil
}

func (column *Columns) IsPrimary() bool {
	if column.Key == "PRI" {
		return true
	}
	return false
}

func (column *Columns) IsAutoIncrement() bool {
	if column.Extra == "auto_increment" {
		return true
	}
	return false
}

func (column *Columns) IsUnique() bool {
	if column.Key == "UNI" {
		return true
	}
	return false
}

func (coulmn *Columns) IsForeignKey() bool {
	return coulmn._isForeignKey
}

func (column *Columns) GetForeignKey() *ForeignKey {
	if (!column.IsForeignKey()) {
		panic("this column has no foreign key!")
	}
	return column.ForeignKey
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ForeignKey struct {
	TableName           string
	ColumnName          string
	Table               *Table
	ConstraintName      string
	ReferenceTableName  string
	ReferenceColumnName string
	ReferenceTable      *Table
}

type ForeignKeyList []*ForeignKey


////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Paginate struct {
	TotalPages  int
	CurrentPage int
	Limit       int
	TotalItems  int
}

func NewPaginated(totalItems int, limit int) *Paginate {

	numberOfPages := totalItems / limit

	if totalItems % limit > 0 {
		numberOfPages++
	}

	return &Paginate{TotalItems:totalItems, Limit:limit, TotalPages:numberOfPages, CurrentPage:0}
}

func (paginated *Paginate) Next() bool {
	paginated.CurrentPage++
	if paginated.CurrentPage > paginated.TotalPages {
		return false
	}
	return true
}

func (paginated *Paginate) getLimitOffset() (int, int) {
	if paginated.CurrentPage == 0 {
		panic("you must call next iterable !")
	}
	return paginated.Limit, (paginated.CurrentPage - 1) * paginated.Limit
}