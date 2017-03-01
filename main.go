package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"runtime"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/unei66/modelq/drivers"
	"github.com/unei66/modelq/gmq"
)
var TypeNullMap=make(map[string]string)

func main() {
	var targetDb, tableNames, packageName string
	var tmplName string
	var driver, schemaName string
	var touchTimestamp bool
	var pCount int
	var structOnly bool
	flag.StringVar(&targetDb, "db", "", "Target database source string: e.g. root@tcp(127.0.0.1:3306)/test?charset=utf-8")
	flag.StringVar(&tableNames, "tables", "", "You may specify which tables the models need to be created, e.g. \"user,article,blog\"")
	flag.StringVar(&packageName, "pkg", "", "Go source code package for generated models")
	flag.StringVar(&driver, "driver", "mysql", "Current supported drivers include mysql, postgres")
	flag.StringVar(&schemaName, "schema", "", "Schema for postgresql, database name for mysql")
	flag.BoolVar(&touchTimestamp, "dont-touch-timestamp", false, "Should touch the datetime fields with default value or on update")
	flag.StringVar(&tmplName, "template", "", "Passing the template to generate code, or use the default one")
	flag.IntVar(&pCount, "p", 4, "Parallell running for code generator")
	flag.BoolVar(&gmq.Debug, "debug", false, "Debug on/off")
	flag.BoolVar(&structOnly, "struct-only", false, "generate struct only")
	flag.Parse()

	runtime.GOMAXPROCS(pCount)

	if targetDb == "" {
		fmt.Println("Please provide the target database source.")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		return
	}
	if packageName == "" {
		printUsages("Please provide the go source code package name for generated models.")
		return
	}
	if driver != "mysql" && driver != "postgres" {
		printUsages("Current supported drivers include mysql, postgres.")
		return
	}
	if schemaName == "" {
		printUsages("Please provide the schema name.")
		return
	}

	dbSchema, err := drivers.LoadDatabaseSchema(driver, targetDb, schemaName, tableNames)
	if err != nil {
		log.Println("Cannot load table schemas from database.")
		log.Fatal(err)
	}

	codeConfig := &CodeConfig{
		packageName:    packageName,
		touchTimestamp: touchTimestamp,
		template:       tmplName,
		structOnly:     structOnly,
	}

	TypeNullMap["string"]="sql.NullString"
	TypeNullMap["int64"]="sql.NullInt64"
	TypeNullMap["int32"]="sql.NullInt64"
	TypeNullMap["int16"]="sql.NullInt64"
	TypeNullMap["int8"]="sql.NullInt64"
	TypeNullMap["int"]="sql.NullInt64"
	TypeNullMap["bool"]="sql.NullBool"
	TypeNullMap["float64"]="sql.NullFloat64"
	TypeNullMap["time.Time"]="mysql.NullTime"

	codeConfig.MustCompileTemplate()
	generateModels(schemaName, dbSchema, *codeConfig)
	formatCodes(packageName)
}

func formatCodes(pkg string) {
	log.Println("Running gofmt *.go")
	var out bytes.Buffer
	cmd := exec.Command("gofmt", "-w", pkg)
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		log.Println(out.String())
		log.Fatalf("Fail to run gofmt package, %s", err)
	}
}

func printUsages(message ...interface{}) {
	for _, x := range message {
		fmt.Println(x)
	}
	fmt.Println("\nUsage:")
	flag.PrintDefaults()
}
