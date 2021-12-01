package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/version"
	"github.com/xuri/excelize/v2"
)

// os
var (
	ex, _        = os.Executable()
	exPath       = filepath.Dir(ex)
	etcPath      = filepath.Join(exPath, "..", "etc")
	templatePath = filepath.Join(exPath, "..", "template")
)

// flag
var (
	anno = flag.String(
		"anno",
		"",
		"anno excel, comma as sep",
	)
	prefix = flag.String(
		"prefix",
		"",
		"prefix of output, output -prefix.-tag.xlsx",
	)
	// option
	template = flag.String(
		"template",
		filepath.Join(templatePath, "新生儿基因筛查结果-159基因.xlsx"),
		"info dir",
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
	if *anno == "" || *prefix == "" {
		flag.Usage()
		log.Printf("-anno/-prefix required!")
		os.Exit(1)
	}
}

func main() {
	患病风险 = simpleUtil.HandleError(textUtil.File2Map(filepath.Join(etcPath, "患病风险.txt"), "\t", false)).(map[string]string)

	var outExcel, err = excelize.OpenFile(*template)
	if err != nil {
		log.Fatalf("can not load [%s]\n", *template)
	}

	var db = loadDB()

	// load sample info
	var infoSheetName = "样本信息"
	var strSlice = simpleUtil.HandleError(
		simpleUtil.HandleError(
			excelize.OpenFile(*anno),
		).(*excelize.File).GetRows(infoSheetName),
	).([][]string)
	log.Printf("load %d rows from %s", len(strSlice), infoSheetName)

	var sheetName = "159基因结果汇总"
	var appendColName = simpleUtil.HandleError(excelize.ColumnNumberToName(colLength + 1)).(string)
	simpleUtil.CheckErr(
		outExcel.SetColWidth(
			sheetName,
			appendColName,
			appendColName,
			20,
		),
	)
	//simpleUtil.CheckErr(
	//	outExcel.SetColStyle(
	//		sheetName,
	//		"B",
	//		simpleUtil.HandleError(outExcel.NewStyle(`{"fill":{"type":"pattern","color":["#E0EBF5"],"pattern":1}, "number_format": 14}`)).(int),
	//	),
	//)

	fillExcel2(strSlice, db, outExcel, sheetName)

	var outputPath = fmt.Sprintf("%s.%s.xlsx", *prefix, *tag)
	simpleUtil.CheckErr(outExcel.SaveAs(outputPath))
	log.Printf("信息：保存到[%s]\n", outputPath)
}
