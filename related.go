package main

import (
	"strings"
	"encoding/json"
)

func (table *Table) GetTag() string {
	tableConfig := SystemConfig.GetTableConfig(table)
	return tableConfig.Label
}

func (table *Table) GetUniqueProperties() []string {
	tableConfig := SystemConfig.GetTableConfig(table)
	return tableConfig.UniqueColumns
}

func (table *Table) GetSkipProperties() []string {
	tableConfig := SystemConfig.GetTableConfig(table)
	return tableConfig.SkipColumns
}

func (table *Table) IsUniqueProperty(key string) bool {
	return stringInSlice(key, table.GetUniqueProperties())
}

func (table *Table) IsSkipProperty(key string) bool {
	return stringInSlice(key, table.GetSkipProperties())
}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func FixStringStyle(str string) string {
	str = strings.Replace(str, "\\", "", -1)
	return strings.Replace(str, "'", "\\'", -1)
}

func (table *Table) GetSetAndProperty(label string, properties map[string]string) (string, string) {
	set := " "
	property := " "
	for key, value := range properties {
		if table.IsSkipProperty(key) {
			continue
		}
		if err, js := isJSON(value); err == nil {
			if table.IsUniqueProperty(key) {
				for sub_key, sub_value := range js {
					property += " " + sub_key + ":'" + FixStringStyle(sub_value) + "' ,"
				}
			} else {
				for sub_key, sub_value := range js {
					set += " " + label + "." + sub_key + "='" + FixStringStyle(sub_value) + "',"
				}
			}
		} else {
			if table.IsUniqueProperty(key) {
				property += " " + key + ":'" + FixStringStyle(value) + "' ,"
			} else {
				set += " " + label + "." + key + "='" + FixStringStyle(value) + "',"
			}
		}

	}
	return set[:len(set) - 1], property[:len(property) - 1]
}

func isJSON(s string) (error, map[string]string) {
	var js map[string]string
	err := json.Unmarshal([]byte(s), &js)
	return err, js
}

func (foreign *ForeignKey) GetRelationSetAndProperty(label string, tableProperties map[string]string, referenceProperties map[string]string) (string, string) {
	set := " "
	property := " "

	relationConfigs := SystemConfig.GetRelationConfig(foreign)
	for key, value := range relationConfigs.Properties {
		data := strings.Split(value, ".")
		tableName := data[0]
		column := data[1]

		if foreign.ReferenceTable.name == tableName {
			set += " " + label + "." + key + "='" + FixStringStyle(referenceProperties[column]) + "',"
		} else if foreign.Table.name == tableName {
			set += " " + label + "." + key + "='" + FixStringStyle(tableProperties[column]) + "',"
		}
	}

	return set[:len(set) - 1], property[:len(property) - 1]
}

func (foreign *ForeignKey) GetTag() string {
	relationConfigs := SystemConfig.GetRelationConfig(foreign)
	if relationConfigs.Label != "" {
		return relationConfigs.Label
	}
	return foreign.ReferenceTable.GetTag() + "_" + foreign.Table.GetTag()
}