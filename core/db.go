package core

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/fztcjjl/ztool/config"
	_ "github.com/go-sql-driver/mysql"
)

// Index
type Index struct {
	Name       string `mysql:"INDEX_NAME"`
	ColumnName string `mysql:"COLUMN_NAME"`
	Comment    string `mysql:"INDEX_COMMENT"`
	Sequence   int    `mysql:"SEQ_IN_INDEX"`
	IsUnique   bool   `mysql:"NON_UNIQUE"`
}

// Column
type Column struct {
	Name            string   `mysql:"COLUMN_NAME"`
	DataType        string   `mysql:"DATA_TYPE"`
	Type            string   `mysql:"COLUMN_TYPE"`
	Default         string   `mysql:"COLUMN_DEFAULT"`
	Comment         string   `mysql:"COLUMN_COMMENT"`
	Length          int64    `mysql:"CHARACTER_MAXIMUM_LENGTH"`
	Precision       int64    `mysql:"NUMERIC_PRECISION"`
	Scale           int64    `mysql:"NUMERIC_SCALE"`
	Position        int      `mysql:"ORDINAL_POSITION"`
	IsPrimaryKey    bool     `mysql:"COLUMN_KEY"`
	IsAutoIncrement bool     `mysql:"EXTRA"`
	IsNullable      bool     `mysql:"IS_NULLABLE"`
	IsUnsigned      bool     `mysql:"COLUMN_TYPE"`
	Indexes         []*Index `mysql:"-"`
	UniqueIndexes   []*Index `mysql:"-"`
}

func (col *Column) String() string {
	return ""
}

// Table
type Table struct {
	Name        string
	Columns     []*Column
	Indexes     map[string]*Index
	StoreEngine string
	Charset     string
	Comment     string
}

func NewTable() *Table {
	return &Table{
		Columns: make([]*Column, 0),
		Indexes: make(map[string]*Index),
	}
}

type DB struct {
	db *sql.DB
}

var (
	once sync.Once
	db   *DB
)

func GetDB() *DB {
	once.Do(func() {
		db = &DB{}
		var err error
		conf := config.GetConfig()
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
			conf.User, conf.Password, conf.Host, conf.Port, conf.Database)

		db.db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Panic(err)
		}

	})
	return db
}

func (db *DB) GetSchema() (tables []*Table, err error) {
	tables, err = db.GetTables()
	if err != nil {
		return
	}

	for _, table := range tables {
		if err = db.loadTableInfo(table); err != nil {
			return nil, err
		}
	}
	return
}

func (db *DB) GetTables() ([]*Table, error) {
	args := []interface{}{config.GetConfig().Database}
	s := "SELECT `TABLE_NAME`, `ENGINE`, `TABLE_ROWS`, `AUTO_INCREMENT`, `TABLE_COMMENT` from " +
		"`INFORMATION_SCHEMA`.`TABLES` WHERE `TABLE_SCHEMA`=? AND (`ENGINE`='MyISAM' OR `ENGINE` = 'InnoDB' OR `ENGINE` = 'TokuDB')"

	rows, err := db.db.Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []*Table
	for rows.Next() {
		table := NewTable()
		var name, engine, tableRows, comment string
		var autoIncr *string
		err = rows.Scan(&name, &engine, &tableRows, &autoIncr, &comment)
		if err != nil {
			return nil, err
		}

		table.Name = name
		table.Comment = comment
		table.StoreEngine = engine
		tables = append(tables, table)
	}
	return tables, nil

}

func (db *DB) loadTableInfo(table *Table) error {
	cols, err := db.GetColumns(table.Name)
	if err != nil {
		return err
	}

	table.Columns = cols

	// TODO: index

	return nil
}

func (db *DB) GetColumns(tableName string) ([]*Column, error) {
	args := []interface{}{config.GetConfig().Database, tableName}
	s := "SELECT `COLUMN_NAME`, `IS_NULLABLE`, `COLUMN_DEFAULT`, `DATA_TYPE`, `COLUMN_TYPE`," +
		" `COLUMN_KEY`, `EXTRA`,`COLUMN_COMMENT` FROM `INFORMATION_SCHEMA`.`COLUMNS` WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?"

	rows, err := db.db.Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []*Column
	for rows.Next() {
		var columnName, isNullable, dataType, colType, colKey, extra, comment string
		var colDefault sql.NullString
		err = rows.Scan(&columnName, &isNullable, &colDefault, &dataType, &colType, &colKey, &extra, &comment)
		if err != nil {
			return nil, err
		}

		col := &Column{
			Name:            strings.Trim(columnName, "` "),
			DataType:        dataType,
			Type:            colType,
			Default:         strings.TrimSpace(colDefault.String),
			Comment:         comment,
			Length:          0,
			Precision:       0,
			Scale:           0,
			Position:        0,
			IsPrimaryKey:    colKey == "PRI",
			IsAutoIncrement: extra == "auto_increment",
			IsNullable:      "YES" == isNullable,
			IsUnsigned:      strings.Contains(colType, "unsigned"),
			//Indexes:         nil,
			//UniqueIndexes:   nil,
		}

		cols = append(cols, col)
		// TODO index
	}
	return cols, nil
}
