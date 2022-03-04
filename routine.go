package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/xuri/excelize/v2"
)

func getAvd(fileName string, dbChan chan<- []map[string]string, throttle, writeExcel chan bool, all bool) {
	log.Printf("load avd[%s]\n", fileName)
	var avd, _ = textUtil.File2MapArray(fileName, "\t", nil)
	var sampleID = filepath.Base(fileName)
	if len(avd) == 0 {
		if all {
			writeExcel <- true
			go func() {
				// all snv
				var allExcel = excelize.NewFile()
				allExcel.NewSheet(*allSheetName)
				var allTitle = textUtil.File2Array(*allColumns)
				writeTitle(allExcel, *allSheetName, allTitle)
				log.Printf("excel.SaveAs(\"%s\")\n", strings.Join([]string{*prefix, "all", sampleID, "xlsx"}, "."))
				simpleUtil.CheckErr(allExcel.SaveAs(strings.Join([]string{*prefix, "all", sampleID, "xlsx"}, ".")))
				<-writeExcel
			}()
		}
		dbChan <- avd
		<-throttle
		return
	}
	if avd[0]["SampleID"] != "" {
		sampleID = avd[0]["SampleID"]
	}

	var geneHash = make(map[string]string)
	var geneInfo, ok = SampleGeneInfo[sampleID]
	if !ok {
		geneInfo = make(map[string]*GeneInfo)
	}
	for _, item := range avd {
		updateAvd(item)
		updateFromAvd(item, geneHash, geneInfo, sampleID)
	}

	var filterAvd []map[string]string
	for _, item := range avd {
		if item["filterAvd"] == "Y" {
			var info, ok = geneInfo[item["Gene Symbol"]]
			if !ok {
				log.Fatalf("geneInfo build error:\t%+v\n", geneInfo)
			} else {
				item["Database"] = info.getTag(item)
			}
			item["遗传模式判读"] = geneHash[item["Gene Symbol"]]
			filterAvd = append(filterAvd, item)
		}
	}
	if all {
		writeExcel <- true
		go func() {
			// all snv
			var allExcel = excelize.NewFile()
			allExcel.NewSheet(*allSheetName)
			var allTitle = textUtil.File2Array(*allColumns)
			writeTitle(allExcel, *allSheetName, allTitle)
			var rIdx0 = 1
			for _, item := range avd {
				rIdx0++
				writeRow(allExcel, *allSheetName, item, allTitle, rIdx0)
			}
			log.Printf("excel.SaveAs(\"%s\")\n", strings.Join([]string{*prefix, "all", sampleID, "xlsx"}, "."))
			simpleUtil.CheckErr(allExcel.SaveAs(strings.Join([]string{*prefix, "all", sampleID, "xlsx"}, ".")))
			<-writeExcel
		}()
	}
	dbChan <- filterAvd
	<-throttle
}

// WriteAvd write AVD sheet to excel
func WriteAvd(excel *excelize.File, runDmd, runAvd chan bool, all bool) {
	log.Println("Write AVD Start")
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

		// wait runDmd done
		runDmd <- true

		var (
			runWrite   = make(chan bool, 1)
			throttle   = make(chan bool, *threshold)
			writeExcel = make(chan bool, *threshold)
			size       = len(avdArray)
			dbChan     = make(chan []map[string]string, size)
		)

		// goroutine writeAvd in case block getAvd since more avd than threshold
		runWrite <- true
		go writeAvd(excel, dbChan, size, runWrite)
		for _, fileName := range avdArray {
			throttle <- true
			go getAvd(fileName, dbChan, throttle, writeExcel, all)
		}
		// wait writeAvd done
		runWrite <- true
		for i := 0; i < *threshold; i++ {
			throttle <- true
			writeExcel <- true
		}
	} else {
		log.Println("Write AVD Skip")
	}
	log.Println("Write AVD Done")
	<-runAvd
}

func writeAvd(excel *excelize.File, dbChan chan []map[string]string, size int, throttle chan bool) {
	var sheetName = *avdSheetName
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)
	var count = 0
	for avd := range dbChan {
		for _, item := range avd {
			rIdx++
			updateINDEX(item, "D", rIdx)
			writeRow(excel, sheetName, item, title, rIdx)
		}
		count++
		if count == size {
			close(dbChan)
		}
	}
	<-throttle
}

// LoadDmd load DMD sheet to excel
func LoadDmd(excel *excelize.File, throttle chan bool) {
	log.Println("Load DMD Start")
	var dmdArray []string
	if *dmdFiles != "" {
		dmdArray = strings.Split(*dmdFiles, ",")
	}
	if *dmdList != "" {
		dmdArray = append(dmdArray, textUtil.File2Array(*dmdList)...)
	}
	if len(dmdArray) > 0 {
		loadDmd(excel, dmdArray)
	} else {
		log.Println("Load DMD Skip")
	}
	log.Println("Load DMD Done")
	<-throttle
}

func loadDmd(excel *excelize.File, dmdArray []string) {
	var sheetName = *dmdSheetName
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	//var title = rows[0]
	var rIdx = len(rows)
	for _, fileName := range dmdArray {
		var dmd, _ = textUtil.File2MapArray(fileName, "\t", nil)
		for _, item := range dmd {
			rIdx++
			updateDmd(item)
			var sampleID = item["#Sample"]
			var gene = item["gene"]
			var CN = strings.Split(item["CopyNum"], ";")[0]
			var cn float64
			if CN == ">4" {
				cn = 5
				log.Printf("treat CopyNum[%s] as 5\n", item["CopyNum"])
			} else {
				cn, err = strconv.ParseFloat(strings.Split(item["CopyNum"], ";")[0], 64)
				if err != nil {
					cn = 3
					log.Printf("treat CopyNum[%s] as 3:%+v\n", item["CopyNum"], err)
				}
			}
			updateSampleGeneInfo(cn, sampleID, gene)
			addDiseases2Cnv(item, multiDiseaseSep, gene)
			updateINDEX(item, "D", rIdx)
			DmdCnv = append(DmdCnv, item)
			//writeRow(excel, sheetName, item, title, rIdx)
		}
	}
}

func writeDmdCnv(excel *excelize.File, throttle chan bool) {
	var sheetName = *dmdSheetName
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)
	for _, item := range DmdCnv {
		rIdx++
		var sampleID = item["#Sample"]
		var gene = item["gene"]
		updateCnvTags(item, sampleID, gene)
		writeRow(excel, sheetName, item, title, rIdx)
	}
	<-throttle
}

// WriteAe write AE sheet to excel
func WriteAe(excel *excelize.File, throttle chan bool) {
	log.Println("Write AE Start")
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
	if len(db) > 0 {
		writeAe(excel, db)
	} else {
		log.Println("Write AE Skip")
	}
	log.Println("Write AE Done")
	<-throttle
}

func writeAe(excel *excelize.File, db map[string]map[string]string) {
	var rows = simpleUtil.HandleError(excel.GetRows(*aeSheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)
	for _, item := range db {
		rIdx++
		updateAe(item)
		updateINDEX(item, "D", rIdx)
		writeRow(excel, *aeSheetName, item, title, rIdx)
	}
}

// WriteQC wirte QC sheet to excel
func WriteQC(excel *excelize.File, throttle chan bool) {
	log.Println("Write QC Start")
	writeQC(excel, loadQC(*qc))
	log.Println("Write QC Done")
	<-throttle
}

func writeQC(excel *excelize.File, db []map[string]string) {
	var rows = simpleUtil.HandleError(excel.GetRows(*qcSheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)
	qcMap, err = textUtil.File2Map(*qcTitle, "\t", true)
	simpleUtil.CheckErr(err, "load qcTitle fail")
	for i, item := range db {
		rIdx++
		updateQC(item, qcMap, i)
		updateINDEX(item, "B", rIdx)
		writeRow(excel, *qcSheetName, item, title, rIdx)
	}
}

func updateQC(item, qcMap map[string]string, i int) {
	item["Order"] = strconv.Itoa(i + 1)
	for k, v := range qcMap {
		item[k] = item[v]
	}
	var inputGender = "null"
	if limsInfo[item["Sample"]]["SEX"] == "1" {
		inputGender = "M"
	} else if limsInfo[item["Sample"]]["SEX"] == "2" {
		inputGender = "F"
	} else {
		inputGender = "null"
	}
	if inputGender != genderMap[limsInfo[item["Sample"]]["MAIN_SAMPLE_NUM"]] {
		item["Gender"] = inputGender + "!!!Sequenced" + genderMap[limsInfo[item["Sample"]]["MAIN_SAMPLE_NUM"]]
	}
	//item["RESULT"]=item[""]
	item["产品编号"] = limsInfo[item["Sample"]]["PRODUCT_CODE"]
}

func updateINDEX(item map[string]string, col string, index int) {
	item["解读人"] = fmt.Sprintf("=INDEX(任务单!O:O,MATCH(%s%d,任务单!I:I,0),1)", col, index)
	item["审核人"] = fmt.Sprintf("=INDEX(任务单!P:P,MATCH(%s%d,任务单!I:I,0),1)", col, index)
}

var (
	multiDiseaseSep      = "\n"
	batchCnvDiseaseTitle = []string{
		"疾病中文名",
		"遗传模式",
	}
)

func writeBatchCnv(throttle chan bool) {
	BatchCnvTitle = append(BatchCnvTitle, "Database")
	BatchCnvTitle = append(BatchCnvTitle, batchCnvDiseaseTitle...)
	var sheetName = "Sheet1"
	var bcExcel = excelize.NewFile()
	writeTitle(bcExcel, sheetName, BatchCnvTitle)
	for i, item := range BatchCnv {
		var genes = strings.Split(item["gene"], ",")
		updateCnvTags(item, item["sample"], genes...)
		addDiseases2Cnv(item, multiDiseaseSep, genes...)
		writeRow(bcExcel, sheetName, item, BatchCnvTitle, i+2)
	}
	var lastCellName = simpleUtil.HandleError(excelize.CoordinatesToCellName(len(BatchCnvTitle), len(BatchCnv)+1)).(string)
	simpleUtil.CheckErr(bcExcel.AddTable(sheetName, "A1", lastCellName, `{"table_style":"TableStyleMedium9"}`), "bcExcel.AddTable Error!")
	simpleUtil.CheckErr(bcExcel.SaveAs(*prefix+".batchCNV.xlsx"), "bcExcel.SaveAs Error!")
	<-throttle
}
