package main

import (
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/xuri/excelize/v2"
	"log"
	"path/filepath"
)

func createMainExcel(path string, mode Mode, all bool) {
	var excel = initExcel(mode)
	fillExcel(excel, mode, all)

	simpleUtil.CheckErr(excel.SaveAs(path))
	log.Printf("excel.SaveAs(\"%s\"):\n", path)
	log.Println("Save main Done")
}

func initExcel(mode Mode) *excelize.File {
	var excel = newExcel(mode)
	styleInit(excel)
	return excel
}

func newExcel(mode Mode) *excelize.File {
	switch mode {
	case NBSP:
		return simpleUtil.HandleError(excelize.OpenFile(mainTemplate)).(*excelize.File)
	case NBSIM:
		return newExcelIM(imSheetList)
	case WGSNB:
		return simpleUtil.HandleError(excelize.OpenFile(wgsTemplate)).(*excelize.File)
	case WGSCS:
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

// init style
func styleInit(excel *excelize.File) {
	var (
		colorRGB = simpleUtil.HandleError(textUtil.File2Map(rgb, "\t", false)).(map[string]string)
		// 验证位点
		checkColor = colorRGB["red"]
		// 正式报告
		formalRreportColor = colorRGB["corn flower blue"]
		// 补充报告
		supplementaryReportColor = colorRGB["yellow green"]
		// style
		checkFont         = `"font":{"color":"` + checkColor + `"}`
		formalFill        = `"fill":{"type":"pattern","pattern":1,"color":["` + formalRreportColor + `"]}`
		supplementaryFill = `"fill":{"type":"pattern","pattern":1,"color":["` + supplementaryReportColor + `"]}`
	)
	formalStyleID = simpleUtil.HandleError(excel.NewStyle(`{` + formalFill + `}`)).(int)
	supplementaryStyleID = simpleUtil.HandleError(excel.NewStyle(`{` + supplementaryFill + `}`)).(int)
	//checkStyleID = simpleUtil.HandleError(excel.NewStyle(`{` + checkFont + `}`)).(int)
	formalCheckStyleID = simpleUtil.HandleError(excel.NewStyle(`{` + formalFill + `,` + checkFont + `}`)).(int)
	supplementaryCheckStyleID = simpleUtil.HandleError(excel.NewStyle(`{` + supplementaryFill + `,` + checkFont + `}`)).(int)
}

// fillExcel fill sheets
func fillExcel(excel *excelize.File, mode Mode, all bool) {
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
	case NBSIM:
		dmdSheetName = "DMD CNV"
		aeSheetName = "THAL CNV"
		avdSheetName = "SNV&INDEL"
	case WGSNB:
		dmdSheetName = "CNV-原始"
		wgsDmdSheetName = "CNV"
	case WGSCS:
		wgsDmdSheetName = "DMD CNV"
	}
	// Sample
	if mode == NBSIM {
		writeDataFile2Sheet(excel, sampleSheetName, *info, mode, updateSample)
	}
	// bam文件路径
	updateBamPath2Sheet(excel, bamPathSheetName, *bamPath, mode)
	// QC -> DMD
	WriteQC(excel, qcSheetName, *qc, mode)
	var dmdResult = LoadDmd4Sheet(
		excel,
		dmdSheetName,
		mode,
		loadFilesAndList(*dmdFiles, *dmdList),
	)
	// DMD -> All variant data
	writeAvd2Sheet(
		excel,
		avdSheetName,
		allSheetName,
		mode,
		loadFilesAndList(*avdFiles, *avdList),
		runAvd,
		all,
	)

	// CNV
	// write CNV after runAvd
	waitChan(runAvd)
	if mode != WGSCS {
		writeData2Sheet(excel, dmdSheetName, mode, dmdResult, updateDMDCNV)
	}
	// DMD-lumpy
	writeDataFile2Sheet(excel, wgsDmdSheetName, *lumpy, mode, updateLumpy)
	// DMD-nator
	writeDataFile2Sheet(excel, wgsDmdSheetName, *nator, mode, updateNator)
	// 补充实验
	go WriteAe(excel, aeSheetName, *dipinResult, *smaResult, mode, writeAeChan)
	// drug, no use
	writeDataFile2Sheet(excel, drugSheetName, *drugResult, mode, updateDrug)
	// 个特
	writeDataList2Sheet(excel, icSheetName, *featureList, mode, updateFeature)
	// 基因ID
	writeDataList2Sheet(excel, geneIDSheetName, *geneIDList, mode, updateGeneID)

	waitChan(writeAeChan)
}

type handleFunc func(map[string]string, Mode)

func writeData2Sheet(excel *excelize.File, sheetName string, mode Mode, db []map[string]string, fn handleFunc) {
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
		fn(item, mode)
		var sampleID = item["sampleID"]
		switch mode {
		case NBSIM:
			updateInfo(item, sampleID, mode)
			updateColumns(item, sheetTitleMap[sheetName])
		case WGSNB:
			updateInfo(item, sampleID, mode)
			updateGender(item, sampleID)
		}
		updateINDEX(item, "D", rIdx)
		writeRow(excel, sheetName, item, title, rIdx, mode)
	}
}

// writeDataFile2Sheet File2MapArray fill in sheet with fn
func writeDataFile2Sheet(excel *excelize.File, sheetName, path string, mode Mode, fn handleFunc) {
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
		fn(item, mode)
		updateINDEX(item, "D", rIdx)
		writeRow(excel, sheetName, item, title, rIdx, mode)
	}
}

// writeDataList2Sheet list of File2MapArray fill in sheet with fn
func writeDataList2Sheet(excel *excelize.File, sheetName, list string, mode Mode, fn handleFunc) {
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
			fn(item, mode)
			writeRow(excel, sheetName, item, title, rIdx, mode)
		}
	}
}
