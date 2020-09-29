package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/goUtil/osUtil"

	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/version"
)

// os
var (
	ex, _        = os.Executable()
	exPath       = filepath.Dir(ex)
	dbPath       = filepath.Join(exPath, "db")
	etcPath      = filepath.Join(exPath, "etc")
	templatePath = filepath.Join(exPath, "template")
)

// flag
var (
	prefix = flag.String(
		"prefix",
		"",
		"output to -prefix.xlsx",
	)
	template = flag.String(
		"template",
		filepath.Join(templatePath, "NBS-final.result-批次号_产品编号.xlsx"),
		"template to be used",
	)
	dropList = flag.String(
		"dropList",
		filepath.Join(etcPath, "drop.list.txt"),
		"drop list for excel",
	)
	avdList = flag.String(
		"avdList",
		"",
		"All variants data file list, one path per line",
	)
	avdFiles = flag.String(
		"avd",
		"",
		"All variants data file list, comma as sep",
	)
	avdSheetName = flag.String(
		"avdSheetName",
		"All variants data",
		"All variants data sheet name",
	)
	diseaseExcel = flag.String(
		"disease",
		filepath.Join(etcPath, "疾病简介和治疗-20200925.xlsx"),
		"disease database excel",
	)
	diseaseSheetName = flag.String(
		"diseaseSheetName",
		"Sheet2",
		"sheet name of disease database excel",
	)
	localDbExcel = flag.String(
		"db",
		filepath.Join(etcPath, "已解读数据库-2020.09.10.xlsx"),
		"已解读数据库",
	)
	localDbSheetName = flag.String(
		"dbSheetName",
		"Sheet1",
		"已解读数据库 sheet name",
	)
	acmgDb = flag.String(
		"acmgDb",
		filepath.Join(etcPath, "acmg.db.list.txt"),
		"acmg db list",
	)
	autoPVS1 = flag.Bool(
		"autoPVS1",
		false,
		"is use autoPVS1",
	)
	dmdFiles = flag.String(
		"dmd",
		"",
		"DMD result file list, comma as sep",
	)
	dmdList = flag.String(
		"dmdList",
		"",
		"DMD result file list, one path per line",
	)
	dmdSheetName = flag.String(
		"dmdSheetName",
		"CNV",
		"DMD result sheet name",
	)
	dipinResult = flag.String(
		"dipin",
		"",
		"dipin result file",
	)
	smaResult = flag.String(
		"sma",
		"",
		"sma result file",
	)
	aeSheetName = flag.String(
		"aeSheetName",
		"补充实验",
		"Additional Experiments sheet name",
	)
	geneList = flag.String(
		"geneList",
		filepath.Join(etcPath, "gene.list.txt"),
		"gene list to filter",
	)
	functionExclude = flag.String(
		"functionExclude",
		filepath.Join(etcPath, "function.exclude.txt"),
		"function list to exclude",
	)
	allSheetName = flag.String(
		"allSheetName",
		"Sheet1",
		"all snv sheet name",
	)
	allColumns = flag.String(
		"allColumns",
		filepath.Join(etcPath, "avd.all.columns.txt"),
		"all snv sheet title",
	)
	gender = flag.String(
		"gender",
		"F",
		"gender for all or gender map file",
	)
	threshold = flag.Int(
		"threshold",
		12,
		"threshold limit",
	)
)

var (
	geneListMap        = make(map[string]bool)
	functionExcludeMap = make(map[string]bool)
	diseaseDb          = make(map[string]map[string]string)
	localDb            = make(map[string]map[string]string)
	dropListMap        = make(map[string][]string)
	genderMap          = make(map[string]string)
)

var (
	throttle   = make(chan bool, 1)
	writeExcel = make(chan bool, 1)
	dbChan     = make(chan []map[string]string, 1)
)

func main() {
	version.LogVersion()
	// flag
	flag.Parse()
	if *prefix == "" {
		flag.Usage()
		log.Println("-prefix are required!")
		os.Exit(1)
	}
	dbChan = make(chan []map[string]string, *threshold)
	throttle = make(chan bool, *threshold+1)
	writeExcel = make(chan bool, *threshold+1)

	if osUtil.FileExists(*gender) {
		genderMap = simpleUtil.HandleError(textUtil.File2Map(*gender, "\t", false)).(map[string]string)
	}

	loadDb()

	var excel = simpleUtil.HandleError(excelize.OpenFile(*template)).(*excelize.File)

	// All variant data
	var avdArray []string
	if *avdFiles != "" {
		avdArray = strings.Split(*avdFiles, ",")
	}
	if *avdList != "" {
		avdArray = append(avdArray, textUtil.File2Array(*avdList)...)
	}
	if len(avdArray) > 0 {
		log.Println("Start load AVD")
		// acmg
		acmg2015.AutoPVS1 = *autoPVS1
		var acmgCfg = simpleUtil.HandleError(textUtil.File2Map(*acmgDb, "\t", false)).(map[string]string)
		for k, v := range acmgCfg {
			acmgCfg[k] = filepath.Join(dbPath, v)
		}
		acmg2015.Init(acmgCfg)

		throttle <- true
		go writeAvd(excel, dbChan, len(avdArray), throttle)
		for _, fileName := range avdArray {
			throttle <- true
			go getAvd(fileName, dbChan, throttle, writeExcel)
		}
	}

	// CNV
	var dmdArray []string
	if *dmdFiles != "" {
		dmdArray = strings.Split(*dmdFiles, ",")
	}
	if *dmdList != "" {
		dmdArray = append(dmdArray, textUtil.File2Array(*dmdList)...)
	}
	if len(dmdArray) > 0 {
		throttle <- true
		go writeDmd(excel, dmdArray, throttle)
	}

	// 补充实验
	log.Println("Start load 补充实验")
	var db = make(map[string]map[string]string)
	if *dipinResult != "" {
		var dipin, _ = textUtil.File2MapArray(*dipinResult, "\t", nil)
		for _, item := range dipin {
			updateDipin(item, db)
		}
	}
	if *smaResult != "" {
		var sma, _ = textUtil.File2MapArray(*smaResult, "\t", nil)
		for _, item := range sma {
			updateSma(item, db)
		}
	}
	throttle <- true
	go writeAe(excel, db, throttle)

	for i := 0; i <= *threshold; i++ {
		throttle <- true
	}
	for i := 0; i <= *threshold; i++ {
		writeExcel <- true
	}

	log.Printf("excel.SaveAs(\"%s\")\n", *prefix+".xlsx")
	simpleUtil.CheckErr(excel.SaveAs(*prefix + ".xlsx"))
}
