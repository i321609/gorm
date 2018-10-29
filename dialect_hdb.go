package gorm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

type hdb struct {
	commonDialect
}

func init() {
	RegisterDialect("hdb", &hdb{})
}

func (hdb) GetName() string {
	return "hdb"
}

// Get Data Type for hdb Dialect
func (s *hdb) DataTypeOf(field *StructField) string {
	var dataValue, sqlType, size, additionalType = ParseFieldStructForDialect(field, s)

	if sqlType == "" {
		switch dataValue.Kind() {
		case reflect.Bool:
			sqlType = "boolean"
		case reflect.Int8:
			if s.fieldCanAutoIncrement(field) {
				field.TagSettings["AUTO_INCREMENT"] = "AUTO_INCREMENT"
				sqlType = "smallint generated always as IDENTITY"
			} else {
				sqlType = "smallint"
			}
		case reflect.Uint8:
			//8-bit unsigned integer
			if s.fieldCanAutoIncrement(field) {
				field.TagSettings["AUTO_INCREMENT"] = "AUTO_INCREMENT"
				sqlType = "tinyint generated always as IDENTITY"
			} else {
				sqlType = "tinyint"
			}
		case reflect.Int16:
			if s.fieldCanAutoIncrement(field) {
				field.TagSettings["AUTO_INCREMENT"] = "AUTO_INCREMENT"
				sqlType = "smallint generated always as IDENTITY"
			} else {
				sqlType = "smallint"
			}
		case reflect.Uint16, reflect.Int, reflect.Int32:
			if s.fieldCanAutoIncrement(field) {
				field.TagSettings["AUTO_INCREMENT"] = "AUTO_INCREMENT"
				sqlType = "int generated always as IDENTITY"
			} else {
				sqlType = "int"
			}
		case reflect.Uint, reflect.Uint32, reflect.Uintptr, reflect.Int64:
			if s.fieldCanAutoIncrement(field) {
				field.TagSettings["AUTO_INCREMENT"] = "AUTO_INCREMENT"
				sqlType = "bigint generated always as IDENTITY"
			} else {
				sqlType = "bigint"
			}
		case reflect.Uint64: //18,446,744,073,709,551,615 (2^64 - 1)
			if s.fieldCanAutoIncrement(field) {
				field.TagSettings["AUTO_INCREMENT"] = "AUTO_INCREMENT"
				sqlType = "DECIMAL(20,0) generated always as IDENTITY"
			} else {
				sqlType = "DECIMAL(20,0)"
			}
		case reflect.Float32:
			sqlType = "real"
		case reflect.Float64:
			sqlType = "double"
		case reflect.String:
			if size > 0 && size <= 5000 {
				sqlType = fmt.Sprintf("varchar(%d)", size)
			} else {
				sqlType = "CLOB"
			}
		case reflect.Struct:
			if _, ok := dataValue.Interface().(time.Time); ok {
				sqlType = "timestamp"
			}
		default:
			if IsByteArrayOrSlice(dataValue) {
				if size > 0 && size <= 5000 {
					sqlType = fmt.Sprintf("varbinary(%d)", size)
				} else {
					sqlType = "blob"
				}
			}
		}
	}

	if sqlType == "" {
		panic(fmt.Sprintf("invalid sql type %s (%s) for hdb", dataValue.Type().Name(), dataValue.Kind().String()))
	}

	if strings.TrimSpace(additionalType) == "" {
		return sqlType
	}
	return fmt.Sprintf("%v %v", sqlType, additionalType)
}

func (s hdb) HasTable(tableName string) bool {
	var count int

	s.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM TABLES where SCHEMA_NAME=CURRENT_SCHEMA and TABLE_NAME='%s'",
		strings.ToUpper(tableName))).Scan(&count)

	return count > 0
}

func (s hdb) CurrentDatabase() (name string) {
	s.db.QueryRow("select DATABASE_NAME from SYS.M_DATABASES").Scan(&name)
	return
}

func (s hdb) HasColumn(tableName string, columnName string) bool {
	var count int

	s.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM SYS.COLUMNS WHERE TABLE_NAME='%s' AND COLUMN_NAME='%s' AND SCHEMA_NAME=CURRENT_SCHEMA",
		strings.ToUpper(tableName), strings.ToUpper(columnName))).Scan(&count)
	return count > 0
}

func (s hdb) HasForeignKey(tableName string, foreignKeyName string) bool {
	var count int

	s.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM REFERENTIAL_CONSTRAINTS WHERE SCHEMA_NAME=CURRENT_SCHEMA AND TABLE_NAME='%s' AND COLUMN_NAME='%s'",
		strings.ToUpper(tableName), strings.ToUpper(foreignKeyName))).Scan(&count)
	return count > 0
}

func (s hdb) HasIndex(tableName string, indexName string) bool {
	var count int
	s.db.QueryRow(fmt.Sprintf("SELECT count(*) FROM  INDEXES WHERE SCHEMA_NAME=CURRENT_SCHEMA AND TABLE_NAME='%s' AND INDEX_NAME='%s'",
		strings.ToUpper(tableName), strings.ToUpper(indexName))).Scan(&count)
	return count > 0
}

func (hdb) Quote(key string) string {
	//return fmt.Sprintf(`"%s"`, key)
	return fmt.Sprintf("%s", key)
}

func (s hdb) AddColumn(tableName string, columnName string, typ string) error {
	_, err := s.db.Exec(fmt.Sprintf("ALTER TABLE %v ADD (%v %v)", tableName, columnName, typ))
	return err
}

func (s hdb) ModifyColumn(tableName string, columnName string, typ string) error {
	_, err := s.db.Exec(fmt.Sprintf("ALTER TABLE %v ALTER (%v %v)", tableName, columnName, typ))
	return err
}

func (hdb) DefaultValueStr() string {
	return "''"
}

func (s hdb) LastInsertId(result sql.Result, querySql string, scope *Scope) (int64, error) {
	var primaryValue int64
	scope.SQLDB().QueryRow(querySql).Scan(&primaryValue)

	return primaryValue, nil
}

