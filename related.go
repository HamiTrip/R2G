package main

import (
	"strings"
	"encoding/json"
	"reflect"
	"strconv"
	"sync"
)


var Mutex = new(sync.Mutex)

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
	var js map[string]interface{}
	err := json.Unmarshal([]byte(s), &js)
	result := map[string]string{}
	if err == nil {
		for key, value := range js {
			result[key] = GetValue(value)
		}
	}
	return err, result
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

func GetValue(data interface{}) string {
	val := reflect.ValueOf(data)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10)
	case reflect.Float32, reflect.Float64:
		data_int := int(val.Float())
		if (val.Float() - float64(data_int)) == 0 {
			return strconv.FormatInt(int64(val.Float()), 10)
		}
		return strconv.FormatFloat(val.Float(), 'f', 6, 64)
	case reflect.Bool:
		if val.Bool() {
			return "true"
		}
		return "false"
	case reflect.String:
		return val.String()

	case reflect.Map:
		b, _ := json.Marshal(val.Interface())
		return string(b)
	case reflect.Slice:
		b, _ := json.Marshal(val.Interface())
		return string(b)
	}
	return ""
}