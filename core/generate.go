package core

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

const goTemplate = `
package {{ .PackageName }}

{{ $nlen := len .Imports}}
{{if gt $nlen 0}}
import (
	{{ range .Imports -}}
		"{{ . }}"
	{{ end }}
)
{{end}}

{{ range .Tables }}
{{- if  .Comment }}// {{ .Comment }}{{ end }} 
type {{ Mapper .Name }} struct {
	{{ $table := . }}
	{{ range .Columns -}}
		{{ $col := . }}
		{{ Mapper $col.Name }} {{ Type $col }} {{ Tag $col }}
		{{- if  .Comment }}// {{ .Comment }}{{ end }} 
	{{- end }}
}


func (_ *{{ Mapper .Name }}) TableName() string {
	return "{{ .Name }}"
}
{{ end }}
`

var goGenerator *template.Template

func init() {
	t := template.New("default")
	t.Funcs(GoLangTmpl.Funcs)

	goGenerator, _ = t.Parse(goTemplate)

}

type LangTmpl struct {
	Funcs      template.FuncMap
	Formater   func(string) (string, error)
	GenImports func([]*Table) map[string]string
}

var GoLangTmpl LangTmpl = LangTmpl{
	template.FuncMap{
		"Mapper": snakeToCamel,
		"Type":   db2GoType,
		"Tag":    tag,
	},
	formatGo,
	genGoImports,
}

func formatGo(src string) (string, error) {
	source, err := format.Source([]byte(src))
	if err != nil {
		return "", err
	}
	return string(source), nil
}

func genGoImports(tables []*Table) map[string]string {
	imports := make(map[string]string)

	for _, table := range tables {
		for _, col := range table.Columns {
			if db2GoType(col) == goTime {
				imports["time"] = "time"
			}
			if db2GoType(col) == softDelete {
				imports["soft_delete"] = "gorm.io/plugin/soft_delete"
			}
		}
	}
	return imports
}

func Generate(dest string) (err error) {
	genDir, err := filepath.Abs(dest)
	if err != nil {
		return
	}
	genDir = strings.Replace(genDir, "\\", "/", -1)
	packageName := path.Base(genDir)
	log.Println(packageName)
	if err = os.MkdirAll(genDir, os.ModePerm); err != nil {
		return
	}

	db := GetDB()
	tables, err := db.GetSchema()
	if err != nil {
		return
	}

	for _, table := range tables {
		source, err := genTable(table, packageName)
		if err != nil {
			return err
		}
		if source == "" {
			continue
		}
		w, err := os.Create(path.Join(genDir, table.Name+".go"))
		if err != nil {
			return err
		}
		if _, err := w.WriteString(source); err != nil {
			return err
		}
		w.Close()
	}
	return
}

func genTable(table *Table, packageName string) (string, error) {
	tbs := []*Table{table}
	imports := GoLangTmpl.GenImports(tbs)

	tmpl := &Tmpl{
		Tables:      tbs,
		Imports:     imports,
		PackageName: packageName,
	}

	bs := bytes.NewBufferString("")
	if err := goGenerator.Execute(bs, tmpl); err != nil {
		return "", err
	}

	tplContent, err := ioutil.ReadAll(bs)
	if err != nil {
		return "", err
	}
	var source string
	if GoLangTmpl.Formater != nil {
		source, err = GoLangTmpl.Formater(string(tplContent))
		if err != nil {
			source = string(tplContent)
		}
	} else {
		source = string(tplContent)
	}

	return source, nil
}

type Tmpl struct {
	Tables      []*Table
	Imports     map[string]string
	PackageName string
}

func tag(col *Column) string {
	var tags []string
	tags = append(tags, fmt.Sprintf("json:%q", col.Name))

	var res []string
	if col.IsPrimaryKey {
		res = append(res, "primaryKey")
	}
	res = append(res, fmt.Sprintf("column:%s", col.Name))
	t := fmt.Sprintf("type:%s", col.Type)
	if col.IsAutoIncrement {
		t += " auto_increment"
	}
	res = append(res, t)
	if !col.IsNullable && !col.IsPrimaryKey {
		res = append(res, "not null")
	}
	if col.Default != "" {
		res = append(res, fmt.Sprintf("default:%s", col.Default))
	}

	if len(res) > 0 {
		tags = append(tags, "gorm:\""+strings.Join(res, ";")+"\"")
	}
	if len(tags) > 0 {
		return "`" + strings.Join(tags, " ") + "`"
	} else {
		return ""
	}
}
