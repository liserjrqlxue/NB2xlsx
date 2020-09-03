package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize/v2"

	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/version"
)

// os
var (
	ex, _        = os.Executable()
	exPath       = filepath.Dir(ex)
	etcPath      = filepath.Join(exPath, "etc")
	templatePath = filepath.Join(exPath, "template")
)

// flag
var (
	perfix = flag.String(
		"perfix",
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
	avdDataFiles = flag.String(
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
	dmdFiles = flag.String(
		"dmd",
		"",
		"DMD result file list, comma as sep",
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
)

var (
	geneListMap        = make(map[string]bool)
	functionExcludeMap = make(map[string]bool)
	diseaseDb          = make(map[string]map[string]string)
	dropListMap        = make(map[string][]string)
)

func main() {
	version.LogVersion()
	flag.Parse()
	if *perfix == "" {
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
	var diseaseExcel = simpleUtil.HandleError(excelize.OpenFile(*diseaseExcel)).(*excelize.File)
	var diseaseSlice = simpleUtil.HandleError(diseaseExcel.GetRows(*diseaseSheetName)).([][]string)
	diseaseDb, _ = simpleUtil.Slice2MapMapArrayMerge(diseaseSlice, "基因", "/")

	// load drop list
	for k, v := range simpleUtil.HandleError(textUtil.File2Map(*dropList, "\t", false)).(map[string]string) {
		dropListMap[k] = strings.Split(v, ",")
	}

	var excel = simpleUtil.HandleError(excelize.OpenFile(*template)).(*excelize.File)

	// All variant data
	var avdFiles []string
	if *avdDataFiles != "" {
		avdFiles = strings.Split(*avdDataFiles, ",")
	}
	if *avdList != "" {
		avdFiles = append(avdFiles, textUtil.File2Array(*avdList)...)
	}
	if len(avdFiles) > 0 {
		var sheetName = *avdSheetName
		var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
		var title = rows[0]
		var rIdx = len(rows)
		for _, fileName := range avdFiles {
			var avd, _ = textUtil.File2MapArray(fileName, "\t", nil)
			for _, item := range avd {
				if filterAvd(item) {
					rIdx++
					updateAvd(item, rIdx)
					writeRow(excel, sheetName, item, title, rIdx)
				}
			}
		}
	}

	// CNV
	if *dmdFiles != "" {
		var sheetName = *dmdSheetName
		var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
		var title = rows[0]
		var rIdx = len(rows)
		for _, fileName := range strings.Split(*dmdFiles, ",") {
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
			var qc, a_result, b_result string
			if item["QC"] != "pass" {
				qc = "_等验证"
			}
			if item["chr11"] == "N" {
				b_result = "阴性"
			} else {
				b_result = item["chr11"]
			}
			if item["chr16"] == "N" {
				a_result = "阴性"
			} else {
				a_result = item["chr16"]
			}
			info["SampleID"] = sampleID
			info["地贫_QC"] = item["QC"]
			info["β地贫_chr11"] = item["chr11"]
			info["α地贫_chr16"] = item["chr16"]
			info["β地贫_最终结果"] = b_result + qc
			info["α地贫_最终结果"] = a_result + qc
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
			var result, qc, qc_result string
			var Categorization = item["SMN1_ex7_cn"]
			var QC = item["qc"]
			if Categorization == "1.5" || Categorization == "1" || QC != "1" {
				qc_result = "_等验证"
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
			info["SMN1 EX7 del最终结果"] = result + qc_result
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

	log.Printf("excel.SaveAs(\"%s\")\n", *perfix+".xlsx")
	simpleUtil.CheckErr(excel.SaveAs(*perfix + ".xlsx"))
}

var formulaTitle = map[string]bool{
	"解读人": true,
	"审核人": true,
}

func filterAvd(item map[string]string) bool {
	if !geneListMap[item["Gene Symbol"]] {
		return false
	}
	if functionExcludeMap[item["Function"]] {
		return false
	}
	var af, err = strconv.ParseFloat(item["GnomAD AF"], 64)
	if err == nil && af > 0.01 {
		return false
	}
	return true
}

func updateAvd(item map[string]string, rIdx int) {
	var gene = item["Gene Symbol"]
	var db, ok = diseaseDb[gene]
	if ok {
		item["疾病中文名"] = db["疾病"]
		item["遗传模式"] = db["遗传模式"]
	}
	item["解读人"] = fmt.Sprintf("=INDEX('任务单（空sheet）'!O:O,MATCH(D%d&MID($C%d,1,6),'任务单（空sheet）'!$R:$R,0),1)", rIdx, rIdx)
	item["审核人"] = fmt.Sprintf("=INDEX('任务单（空sheet）'!P:P,MATCH(D%d&MID($C%d,1,6),'任务单（空sheet）'!$R:$R,0),1)", rIdx, rIdx)
}

func updateDmd(item map[string]string, rIdx int) {
	item["#sample"] = item["#Sample"]
	item["OMIM"] = item["Disease"]
	if item["Significant"] != "YES" {
		item["Significant"] = "NO"
	}
	/*
		item["depth_rate"]=item["batch_control"]
		item["others_rate"]=item["all_control"]
		item["G/H"]=item["Mean_Ratio"]
		item["Zscore"]=item["Median_Ratio"]
		var omimWebsite="http://omim.org/search/?index=entry&start=1&limit=10&sort=score+desc%2C+prefix_sort+desc&search="+item["gene"]
		item["omimWebsite"]=omimWebsite
	*/
	// primerDesign
	var exId = item["exon"]
	var cdsId = item["exon"]
	var ratioVal, err = strconv.ParseFloat(item["Mean_Ratio"], 64)
	if err != nil {
		ratioVal = 0
	}
	if ratioVal >= 1.3 && ratioVal < 1.8 {
		item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exId + " DUP; - ;" + exId + "; " + cdsId + "; Het"
	} else if ratioVal >= 1.8 {
		if item["chr"] == "chrX" {
			item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exId + " DUP; - ;" + exId + "; " + cdsId + "; Hemi"
		} else {
			item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exId + " DUP; - ;" + exId + "; " + cdsId + "; Hom"
		}
	} else if ratioVal >= 0.2 && ratioVal <= 0.75 {
		item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exId + " DEL; - ;" + exId + "; " + cdsId + "; Het"
	} else if ratioVal < 0.2 {
		if item["chr"] == "chrX" {
			item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exId + " DEL; - ;" + exId + "; " + cdsId + "; Hemi"
		} else {
			item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exId + " DEL; - ;" + exId + "; " + cdsId + "; Hom"
		}
	} else {
		item["primerDesign"] = "-"
	}
	item["解读人"] = fmt.Sprintf("=INDEX('任务单（空sheet）'!O:O,MATCH(D%d&MID($C%d,1,6),'任务单（空sheet）'!$R:$R,0),1)", rIdx, rIdx)
	item["审核人"] = fmt.Sprintf("=INDEX('任务单（空sheet）'!P:P,MATCH(D%d&MID($C%d,1,6),'任务单（空sheet）'!$R:$R,0),1)", rIdx, rIdx)
}

func updateAe(item map[string]string, rIdx int) {
	item["F8int1h-1.5k&2k最终结果"] = "检测范围外"
	item["F8int22h-10.8k&12k最终结果"] = "检测范围外"
	item["解读人"] = fmt.Sprintf("=INDEX('任务单（空sheet）'!O:O,MATCH(D%d&MID($C%d,1,6),'任务单（空sheet）'!$R:$R,0),1)", rIdx, rIdx)
	item["审核人"] = fmt.Sprintf("=INDEX('任务单（空sheet）'!P:P,MATCH(D%d&MID($C%d,1,6),'任务单（空sheet）'!$R:$R,0),1)", rIdx, rIdx)
}

func writeRow(excel *excelize.File, sheetName string, item map[string]string, title []string, rIdx int) {
	for j, k := range title {
		var axis = simpleUtil.HandleError(excelize.CoordinatesToCellName(j+1, rIdx)).(string)
		if formulaTitle[k] {
			simpleUtil.CheckErr(excel.SetCellFormula(sheetName, axis, item[k]))
		} else {
			simpleUtil.CheckErr(excel.SetCellValue(sheetName, axis, item[k]))
		}
		var list, ok = dropListMap[k]
		if ok {
			var dvRange = excelize.NewDataValidation(true)
			dvRange.Sqref = axis
			simpleUtil.CheckErr(dvRange.SetDropList(list))
			simpleUtil.CheckErr(excel.AddDataValidation(sheetName, dvRange))
		}
	}
}
