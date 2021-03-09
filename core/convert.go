package core

import (
	"strings"
	"unicode"
)

const (
	goString   = "string"
	goBytes    = "[]byte"
	goInt      = "int"
	goUint     = "uint"
	goInt32    = "int32"
	goUint32   = "uint32"
	goInt64    = "int64"
	goUint64   = "uint64"
	goFloat32  = "float32"
	goFloat64  = "float64"
	goBool     = "bool"
	goTime     = "time.Time"
	softDelete = "soft_delete.DeletedAt"
)

func db2GoType(col *Column) string {
	if col.Name == "deleted_at" {
		return softDelete
	}
	switch col.DataType {
	case "tinyint", "smallint", "mediumint":
		if col.IsUnsigned {
			return goUint
		}
		return goInt
	case "int", "integer":
		if col.IsUnsigned {
			return goUint
		}
		return goInt
	case "bigint":
		if col.IsUnsigned {
			return goUint64
		}
		return goInt64
	case "json", "enum", "set", "char", "varchar", "tinytext", "text", "mediumtext", "longtext", "decimal", "numeric":
		return goString
	case "year", "date", "datetime", "time", "timestamp":
		return goTime
	case "float":
		return goFloat32
	case "double", "real":
		return goFloat64
	case "bit", "binary", "varbinary", "tinyblob", "blob", "mediumblob", "longblob":
		return goBytes
	default:
		return "unknown"
	}
}

func snakeToCamel(s string) string {
	var result string
	words := strings.Split(s, "_")
	for _, word := range words {
		if len(word) > 0 {
			w := []rune(word)
			for i, _ := range w {
				if i == 0 {
					w[i] = unicode.ToUpper(w[i])
				} else {
					w[i] = unicode.ToLower(w[i])
				}
			}
			result += string(w)
		}
	}

	return result
}
