package main

import (
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/xuri/excelize/v2"
	"log"
	"path/filepath"
)

func createMainExcel(excel *excelize.File, path, mode string, all bool, throttle chan<- bool) {
	initExcel(excel, mode)
	fillExcel(excel, mode, all)

	simpleUtil.CheckErr(excel.SaveAs(path))
	log.Printf("excel.SaveAs(\"%s\"):\n", path)
	log.Println("Save main Done")

	holdChan(throttle)
}

func initExcel(excel *excelize.File, mode string) {
	excel = newExcel(mode)
	styleInit(excel)
}

func newExcel(mode string) *excelize.File {
	switch mode {
	case "NBSIM":
		return newExcelIM(imSheetList)
	case "NBSP":
		return simpleUtil.HandleError(excelize.OpenFile(mainTemplate)).(*excelize.File)
	case "WGSNB":
		return simpleUtil.HandleError(excelize.OpenFile(wgsTemplate)).(*excelize.File)
	case "WGSCS":
		return simpleUtil.HandleError(excelize.OpenFile(csTemplate)).(*excelize.File)
	default:
		log.Fatalf("mode [%s] not suppoort!", mode)
	}
	return nil
}

func newExcelIM(sheetNames []string) *excelize.File {
	var excel = excelize.NewFile()
	for _, s := range sheetNames {
		excel.NewSheet(s)
		var titleMaps, _ = textUtil.File2MapArray(filepath.Join(templatePath, s+".txt"), "\t", nil)
		var title []string
		for _, m := range titleMaps {
			title = append(title, m[columnName])
		}
		writeTitle(excel, s, title)
	}
	excel.DeleteSheet("Sheet1")
	return excel
}

func styleInit(excel *excelize.File) {
	colorRGB = simpleUtil.HandleError(textUtil.File2Map(filepath.Join(etcPath, "color.RGB.tsv"), "\t", false)).(map[string]string)
	checkColor = colorRGB["red"]
	formalRreportColor = colorRGB["corn flower blue"]
	supplementaryReportColor = colorRGB["yellow green"]
	var checkFont = `"font":{"color":"` + checkColor + `"}`
	var formalFill = `"fill":{"type":"pattern","pattern":1,"color":["` + formalRreportColor + `"]}`
	var supplementaryFill = `"fill":{"type":"pattern","pattern":1,"color":["` + supplementaryReportColor + `"]}`
	formalStyleID = simpleUtil.HandleError(excel.NewStyle(`{` + formalFill + `}`)).(int)
	supplementaryStyleID = simpleUtil.HandleError(excel.NewStyle(`{` + supplementaryFill + `}`)).(int)
	//checkStyleID = simpleUtil.HandleError(excel.NewStyle(`{` + checkFont + `}`)).(int)
	formalCheckStyleID = simpleUtil.HandleError(excel.NewStyle(`{` + formalFill + `,` + checkFont + `}`)).(int)
	supplementaryCheckStyleID = simpleUtil.HandleError(excel.NewStyle(`{` + supplementaryFill + `,` + checkFont + `}`)).(int)
}

// fillExcel fill sheets
func fillExcel(excel *excelize.File, mode string, all bool) {
	var (
		// local sheet names
		// Additional Experiments sheet name
		aeSheetName = "补充实验"
		// All samples' snv Excel sheet name
		allSheetName = "Sheet1"
		// All variants data sheet name
		avdSheetName = "All variants data"
		// Bam path sheet name
		bamPathSheetName = "bam文件路径"
		// DMD result sheet name
		dmdSheetName = "CNV"
		// WGS DMD result sheet name
		wgsDmdSheetName = "CNV"
		// Drug sheet name
		drugSheetName = "药物检测结果"
		// QC sheet name
		qcSheetName = "QC"
		// Sample sheet name
		sampleSheetName = "Sample"
		// individual characteristics sheet name
		icSheetName = "个特"
		// Gene ID sheet name
		geneIDSheetName = "基因ID"
		// batchCNV Excel sheet name

		// un-block channel bool
		runAvd      = make(chan bool, 1)
		writeAeChan = make(chan bool, 1)
	)
	switch mode {
	case "NBSIM":
		dmdSheetName = "DMD CNV"
		aeSheetName = "THAL CNV"
		avdSheetName = "SNV&INDEL"
	case "WGSNB":
		dmdSheetName = "CNV-原始"
		wgsDmdSheetName = "CNV"
	case "WGSCS":
		wgsDmdSheetName = "DMD CNV"
	}
	// Sample
	if mode == "NBSIM" {
		writeDataFile2Sheet(excel, sampleSheetName, *info, updateSample)
	}
	// bam文件路径
	updateBamPath2Sheet(excel, bamPathSheetName, *bamPath)
	// QC -> DMD
	WriteQC(excel, qcSheetName, *qc)
	var dmdResult = LoadDmd4Sheet(
		excel,
		dmdSheetName,
		loadFilesAndList(*dmdFiles, *dmdList),
	)
	// 补充实验
	WriteAe(excel, aeSheetName, writeAeChan)
	// DMD -> All variant data
	writeAvd2Sheet(
		excel,
		avdSheetName,
		allSheetName,
		loadFilesAndList(*avdFiles, *avdList),
		runAvd,
		all,
	)

	// CNV
	// write CNV after runAvd
	waitChan(runAvd)
	if mode != "WGSCS" {
		writeData2Sheet(excel, dmdSheetName, dmdResult, updateDMDCNV)
	}
	// DMD-lumpy
	writeDataFile2Sheet(excel, wgsDmdSheetName, *lumpy, updateLumpy)
	// DMD-nator
	writeDataFile2Sheet(excel, wgsDmdSheetName, *nator, updateNator)
	// drug, no use
	writeDataFile2Sheet(excel, drugSheetName, *drugResult, updateDrug)
	// 个特
	writeDataList2Sheet(excel, icSheetName, *featureList, updateFeature)
	// 基因ID
	writeDataList2Sheet(excel, geneIDSheetName, *geneIDList, updateGeneID)

	waitChan(writeAeChan)
}

type handleFunc func(map[string]string)

func writeData2Sheet(excel *excelize.File, sheetName string, db []map[string]string, fn handleFunc) {
	if len(db) == 0 {
		log.Printf("skip update [%s] for empty db", sheetName)
		return
	}

	log.Printf("update [%s]", sheetName)
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)

	for _, item := range db {
		rIdx++
		fn(item)
		if *im {
			updateInfo(item, item["sampleID"])
			updateColumns(item, sheetTitleMap[sheetName])
		}
		if *wgs {
			updateInfo(item, item["sampleID"])
			updateGender(item, item["sampleID"])
		}
		writeRow(excel, sheetName, item, title, rIdx)
	}
}

// writeDataFile2Sheet File2MapArray fill in sheet with fn
func writeDataFile2Sheet(excel *excelize.File, sheetName, path string, fn handleFunc) {
	if path == "" {
		log.Printf("skip update [%s] for no path", sheetName)
		return
	}
	var db, _ = textUtil.File2MapArray(path, "\t", nil)
	if len(db) == 0 {
		log.Printf("skip update [%s] for empty path [%s]", sheetName, path)
		return
	}

	log.Printf("update [%s]", sheetName)
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)

	for _, item := range db {
		rIdx++
		fn(item)
		updateINDEX(item, "D", rIdx)
		writeRow(excel, sheetName, item, title, rIdx)
	}
}

// writeDataList2Sheet list of File2MapArray fill in sheet with fn
func writeDataList2Sheet(excel *excelize.File, sheetName, list string, fn handleFunc) {
	if list == "" {
		log.Printf("skip update [%s] for no list", sheetName)
		return
	}

	var lists = textUtil.File2Array(list)
	if len(lists) == 0 {
		log.Printf("skip update [%s] for empty list [%s]", sheetName, list)
		return
	}

	log.Printf("update [%s]", sheetName)
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)

	for _, path := range lists {
		var db, _ = textUtil.File2MapArray(path, "\t", nil)
		for _, item := range db {
			rIdx++
			fn(item)
			writeRow(excel, sheetName, item, title, rIdx)
		}
	}
}
