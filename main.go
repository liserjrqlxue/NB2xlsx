package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/liserjrqlxue/acmg2015"

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
		filepath.Join(etcPath, "疾病简介和治疗-20200825.xlsx"),
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
)

var (
	geneListMap        = make(map[string]bool)
	functionExcludeMap = make(map[string]bool)
	diseaseDb          = make(map[string]map[string]string)
	localDb            = make(map[string]map[string]string)
	dropListMap        = make(map[string][]string)
)

func main() {
	version.LogVersion()
	flag.Parse()
	if *prefix == "" {
		flag.Usage()
		log.Println("-prefix are required!")
		os.Exit(1)
	}

	// load gene list
	for _, key := range textUtil.File2Array(*geneList) {
		geneListMap[key] = true
	}

	// load function exclude list
	for _, key := range textUtil.File2Array(*functionExclude) {
		functionExcludeMap[key] = true
	}

	// load disease database
	diseaseDb, _ = simpleUtil.Slice2MapMapArrayMerge(
		simpleUtil.HandleError(
			simpleUtil.HandleError(
				excelize.OpenFile(*diseaseExcel),
			).(*excelize.File).
				GetRows(*diseaseSheetName),
		).([][]string),
		"基因",
		"/",
	)

	// load 已解读数据库
	localDb, _ = simpleUtil.Slice2MapMapArray(
		simpleUtil.HandleError(
			simpleUtil.HandleError(
				excelize.OpenFile(*localDbExcel),
			).(*excelize.File).
				GetRows(*localDbSheetName),
		).([][]string),
		"Transcript", "cHGVS",
	)

	// load drop list
	for k, v := range simpleUtil.HandleError(textUtil.File2Map(*dropList, "\t", false)).(map[string]string) {
		dropListMap[k] = strings.Split(v, ",")
	}

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
		// all snv
		var allExcel = excelize.NewFile()
		allExcel.NewSheet(*allSheetName)
		var allTitle = textUtil.File2Array(*allColumns)
		writeTitle(allExcel, *allSheetName, allTitle)
		var rIdx0 = 1

		acmg2015.AutoPVS1 = *autoPVS1
		var acmgCfg = simpleUtil.HandleError(textUtil.File2Map(*acmgDb, "\t", false)).(map[string]string)
		for k, v := range acmgCfg {
			acmgCfg[k] = filepath.Join(dbPath, v)
		}
		acmg2015.Init(acmgCfg)
		var sheetName = *avdSheetName
		var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
		var title = rows[0]
		var rIdx = len(rows)
		for _, fileName := range avdArray {
			var avd, _ = textUtil.File2MapArray(fileName, "\t", nil)
			for _, item := range avd {
				rIdx0++
				updateAvd(item, rIdx)
				writeRow(allExcel, *allSheetName, item, allTitle, rIdx0)
				if filterAvd(item) {
					rIdx++
					writeRow(excel, sheetName, item, title, rIdx)
				}
			}
		}
		log.Printf("excel.SaveAs(\"%s\")\n", *prefix+".all.xlsx")
		simpleUtil.CheckErr(allExcel.SaveAs(*prefix + ".all.xlsx"))
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
		var sheetName = *dmdSheetName
		var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
		var title = rows[0]
		var rIdx = len(rows)
		for _, fileName := range dmdArray {
			var dmd, _ = textUtil.File2MapArray(fileName, "\t", nil)
			for _, item := range dmd {
				rIdx++
				updateDmd(item, rIdx)
				writeRow(excel, sheetName, item, title, rIdx)
			}
		}
	}

	// 补充实验
	var db = make(map[string]map[string]string)
	if *dipinResult != "" {
		var dipin, _ = textUtil.File2MapArray(*dipinResult, "\t", nil)
		for _, item := range dipin {
			var sampleID = item["sample"]
			var info, ok = db[sampleID]
			if !ok {
				info = item
			}
			var qc, aResult, bResult string
			if item["QC"] != "pass" {
				qc = "_等验证"
			}
			if item["chr11"] == "N" {
				bResult = "阴性"
			} else {
				bResult = item["chr11"]
			}
			if item["chr16"] == "N" {
				aResult = "阴性"
			} else {
				aResult = item["chr16"]
			}
			info["SampleID"] = sampleID
			info["地贫_QC"] = item["QC"]
			info["β地贫_chr11"] = item["chr11"]
			info["α地贫_chr16"] = item["chr16"]
			info["β地贫_最终结果"] = bResult + qc
			info["α地贫_最终结果"] = aResult + qc
			db[sampleID] = info
		}
	}
	if *smaResult != "" {
		var sma, _ = textUtil.File2MapArray(*smaResult, "\t", nil)
		for _, item := range sma {
			var sampleID = item["SampleID"]
			var info, ok = db[sampleID]
			if !ok {
				info = item
			}
			var result, qc, qcResult string
			var Categorization = item["SMN1_ex7_cn"]
			var QC = item["qc"]
			if Categorization == "1.5" || Categorization == "1" || QC != "1" {
				qcResult = "_等验证"
			}
			switch Categorization {
			case "0":
				result = "纯阳性"
			case "0.5":
				result = "纯合灰区"
			case "1":
				result = "杂合阳性"
			case "1.5":
				result = "杂合灰区"
			default:
				result = "阴性"
			}
			if QC == "1" {
				qc = "Pass"
			} else {
				qc = "Fail"
			}
			info["SMN1_检测结果"] = result
			info["SMN1_质控结果"] = qc
			info["SMN1 EX7 del最终结果"] = result + qcResult
		}
	}
	var rows = simpleUtil.HandleError(excel.GetRows(*aeSheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)
	for _, item := range db {
		rIdx++
		updateAe(item, rIdx)
		writeRow(excel, *aeSheetName, item, title, rIdx)
	}

	log.Printf("excel.SaveAs(\"%s\")\n", *prefix+".xlsx")
	simpleUtil.CheckErr(excel.SaveAs(*prefix + ".xlsx"))
}
