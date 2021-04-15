package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/version"
)

// os
var (
	ex, _   = os.Executable()
	exPath  = filepath.Dir(ex)
	etcPath = filepath.Join(exPath, "..", "etc")
)

// flag
var (
	anno = flag.String(
		"anno",
		"",
		"anno excel, comma as sep",
	)
	input = flag.String(
		"input",
		"",
		"info dir",
	)
	prefix = flag.String(
		"prefix",
		"",
		"prefix of output, output -prefix.-tag.xlsx",
	)
	tag = flag.String(
		"tag",
		time.Now().Format("2006-01-02"),
		"tag of output, default is date[2006-01-02]",
	)
)

var 患病风险 map[string]string

var sampleCount = make(map[string]int)

func init() {
	version.LogVersion()
	flag.Parse()
	if *anno == "" || *input == "" {
		flag.Usage()
		log.Printf("-anno/-input required!")
		os.Exit(1)
	}
	if *prefix == "" {
		*prefix = *input
	}
}

func main() {
	患病风险 = simpleUtil.HandleError(textUtil.File2Map(filepath.Join(etcPath, "患病风险.txt"), "\t", false)).(map[string]string)

	var outExcel, err = excelize.OpenFile(*input)
	if err != nil {
		log.Fatalf("can not load [%s]\n", *input)
	}

	var db = loadDB()

	// load sample info
	var sheetName = "159基因结果汇总"
	var strSlice = simpleUtil.HandleError(
		simpleUtil.HandleError(
			excelize.OpenFile(*input),
		).(*excelize.File).GetRows(sheetName),
	).([][]string)

	var appendColName = simpleUtil.HandleError(excelize.ColumnNumberToName(colLength + 1)).(string)
	simpleUtil.CheckErr(
		outExcel.SetColWidth(
			sheetName,
			appendColName,
			appendColName,
			10,
		),
	)
	simpleUtil.CheckErr(
		outExcel.SetColStyle(
			sheetName,
			appendColName,
			simpleUtil.HandleError(outExcel.NewStyle(`{"fill":{"type":"pattern","color":["#E0EBF5"],"pattern":1}, "number_format": 14}`)).(int),
		),
	)

	fillExcel(strSlice, db, outExcel, sheetName)

	var outputPath = fmt.Sprintf("%s.%s.xlsx", *prefix, *tag)
	simpleUtil.CheckErr(outExcel.SaveAs(outputPath))
	log.Printf("信息：保存到[%s]\n", outputPath)
}
