package main

import (
	"strings"
	"os"
	"fmt"
	"encoding/json"
)

type TableConfig struct {
	Name          string
	IsManyToMany  bool
	Label         string
	UniqueColumns []string
	SkipColumns   []string
}

type RelationConfig struct {
	Label          string
	Table          string
	ReferenceTable string
	Properties     map[string]string
}

type GraphConfig struct {
	Host     string
	Port     string
	Username string
	Password string
}

type MysqlConfig struct {
	Host     string
	Port     string
	DbName   string
	Username string
	Password string
}

type DatabaseConfig struct {
	Graph GraphConfig
	Mysql MysqlConfig
}

type Configuration struct {
	TablesConfig    []*TableConfig
	RelationsConfig []*RelationConfig
	Limit           int
	SkipTables      []string
	Database        *DatabaseConfig
	Debug           bool
}

type tempConfig struct {
	TablesConfig    []TableConfig
	RelationsConfig []RelationConfig
	QueryLimit      int
	SkipTables      []string
	Database        DatabaseConfig
	Debug           bool
}

var SystemConfig *Configuration

func (configuration *Configuration) GetTableConfig(table *Table) *TableConfig {
	for _, config := range configuration.TablesConfig {
		if config.Name == table.name {
			return config
		}
	}
	panic("Config Not Found !")
}

func (configuration *Configuration) GetRelationConfig(foreignKey *ForeignKey) *RelationConfig {
	for _, config := range configuration.RelationsConfig {
		if config.Table == foreignKey.Table.name && config.ReferenceTable == foreignKey.ReferenceTable.name {
			return config
		}
	}
	return &RelationConfig{}
}

func InitConfiguration() {
	SystemConfig = &Configuration{TablesConfig: []*TableConfig{}, RelationsConfig: []*RelationConfig{}}
	loadConfig()
}

func (configuration *Configuration) InitDbConfig(tables map[string]*Table) {
	for _, table := range tables {
		config := configuration.findTableConfig(table.name)
		config.setupConfig(table)
	}
}

func loadConfig() {
	adr, _ := os.Getwd()
	file, err := os.Open(adr + "/conf.json")
	if err != nil {
		fmt.Println("Cannot load Config File ! \n" + err.Error())
	} else {
		decoder := json.NewDecoder(file)
		temp := tempConfig{}
		err := decoder.Decode(&temp)
		if err != nil {
			panic("Invalid Config file! => error:" + err.Error())
		}
		for _, tempTable := range temp.TablesConfig {
			SystemConfig.TablesConfig = append(SystemConfig.TablesConfig, &TableConfig{
				Name:tempTable.Name,
				UniqueColumns:tempTable.UniqueColumns,
				IsManyToMany:tempTable.IsManyToMany,
				Label:tempTable.Label,
				SkipColumns:tempTable.SkipColumns,
			})
		}

		for _, tempRelation := range temp.RelationsConfig {
			SystemConfig.RelationsConfig = append(SystemConfig.RelationsConfig, &RelationConfig{
				Table:tempRelation.Table,
				Label:tempRelation.Label,
				ReferenceTable:tempRelation.ReferenceTable,
				Properties:tempRelation.Properties,
			})
		}

		if temp.QueryLimit == 0 {
			SystemConfig.Limit = 20
		} else {
			SystemConfig.Limit = temp.QueryLimit
		}

		SystemConfig.SkipTables = temp.SkipTables
		SystemConfig.Database = &temp.Database
		SystemConfig.Debug = temp.Debug
	}

}

func (configuration *Configuration) findTableConfig(table string) *TableConfig {
	for _, config := range configuration.TablesConfig {
		if config.Name == table {
			return config
		}
	}
	config := &TableConfig{}
	configuration.TablesConfig = append(configuration.TablesConfig, config)
	return config
}

func (tableConfig *TableConfig) setupConfig(table *Table) {
	if tableConfig.Name == "" {
		tableConfig.Name = table.name
	}

	if tableConfig.Label == "" {
		tableConfig.Label = strings.ToUpper(table.name)
	}

	if len(tableConfig.UniqueColumns) == 0 {
		unique := []string{}
		for _, column := range table.columns {
			if column.IsPrimary() || column.IsUnique() {
				unique = append(unique, column.Field)
			}
		}
		tableConfig.UniqueColumns = unique
	}

}