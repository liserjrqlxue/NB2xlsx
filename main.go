package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
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
	output = flag.String(
		"output",
		"",
		"output file path",
	)
	template = flag.String(
		"template",
		filepath.Join(templatePath, "NBS-final.result-批次号_产品编号.xlsx"),
		"template to be used",
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
)

func main() {
	version.LogVersion()
	flag.Parse()
	if *output == "" || *avdDataFiles == "" {
		flag.Usage()
		log.Println("-output and -avd are required!")
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

	var excel = simpleUtil.HandleError(excelize.OpenFile(*template)).(*excelize.File)

	// All variant data
	if *avdDataFiles != "" {
		var sheetName = *avdSheetName
		var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
		var title = rows[0]
		var rIdx = len(rows)
		for _, fileName := range strings.Split(*avdDataFiles, ",") {
			var avd, _ = textUtil.File2MapArray(fileName, "\t", nil)
			for _, item := range avd {
				if filterAvd(item) {
					rIdx++
					for j, k := range title {
						var axis = simpleUtil.HandleError(excelize.CoordinatesToCellName(j+1, rIdx)).(string)
						simpleUtil.CheckErr(excel.SetCellValue(*avdSheetName, axis, item[k]))
					}
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
				for j, k := range title {
					var axis = simpleUtil.HandleError(excelize.CoordinatesToCellName(j+1, rIdx)).(string)
					simpleUtil.CheckErr(excel.SetCellValue(*dmdSheetName, axis, item[k]))
				}
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
			info["SampleID"] = sampleID
			info["地贫_QC"] = item["QC"]
			info["β地贫_chr11"] = item["chr11"]
			info["α地贫_chr16"] = item["chr16"]
			db[sampleID] = info
		}
	}
	var rows = simpleUtil.HandleError(excel.GetRows(*aeSheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)
	for _, item := range db {
		rIdx++
		for j, k := range title {
			var axis = simpleUtil.HandleError(excelize.CoordinatesToCellName(j+1, rIdx)).(string)
			simpleUtil.CheckErr(excel.SetCellValue(*aeSheetName, axis, item[k]))
		}
	}

	log.Printf("excel.SaveAs(\"%s\")\n", *output)
	simpleUtil.CheckErr(excel.SaveAs(*output))
}

func filterAvd(item map[string]string) bool {
	var gene = item["Gene Symbol"]
	var function = item["Function"]
	return geneListMap[gene] && functionExcludeMap[function]
}
