package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
)

func getAvd(fileName string, dbChan chan<- []map[string]string, throttle, writeExcel chan bool) {
	log.Printf("load avd[%s]\n", fileName)
	var avd, _ = textUtil.File2MapArray(fileName, "\t", nil)
	var sampleID = filepath.Base(fileName)
	if len(avd) == 0 {
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
		if item["filterAvd"] == "Y" {
			var info, ok = geneInfo[item["Gene Symbol"]]
			if !ok {
				info = new(GeneInfo).new(item)
				geneInfo[item["Gene Symbol"]] = info
			} else {
				info.count(item)
			}
			if *gender == "M" || genderMap[sampleID] == "M" {
				updateGeneHash(geneHash, item, "M")
			} else if *gender == "F" || genderMap[sampleID] == "F" {
				updateGeneHash(geneHash, item, "F")
			}
		}
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
	dbChan <- filterAvd
	<-throttle
}

func WriteAvd(excel *excelize.File, runDmd, runAvd chan bool) {
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
			go getAvd(fileName, dbChan, throttle, writeExcel)
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
			updateINDEX(item, rIdx)
			writeRow(excel, sheetName, item, title, rIdx)
		}
		count++
		if count == size {
			close(dbChan)
		}
	}
	<-throttle
}

func WriteDmd(excel *excelize.File, throttle chan bool) {
	log.Println("Write DMD Start")
	var dmdArray []string
	if *dmdFiles != "" {
		dmdArray = strings.Split(*dmdFiles, ",")
	}
	if *dmdList != "" {
		dmdArray = append(dmdArray, textUtil.File2Array(*dmdList)...)
	}
	if len(dmdArray) > 0 {
		writeDmd(excel, dmdArray)
	} else {
		log.Println("Write DMD Skip")
	}
	log.Println("Write DMD Done")
	<-throttle
}

func writeDmd(excel *excelize.File, dmdArray []string) {
	var sheetName = *dmdSheetName
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	var title = rows[0]
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
			updateINDEX(item, rIdx)
			writeRow(excel, sheetName, item, title, rIdx)
		}
	}
}

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
		updateINDEX(item, rIdx)
		writeRow(excel, *aeSheetName, item, title, rIdx)
	}
}

func updateINDEX(item map[string]string, rIdx int) {
	item["解读人"] = fmt.Sprintf("=INDEX('任务单（空sheet）'!O:O,MATCH(D%d&MID($C%d,1,6),'任务单（空sheet）'!$R:$R,0),1)", rIdx, rIdx)
	item["审核人"] = fmt.Sprintf("=INDEX('任务单（空sheet）'!P:P,MATCH(D%d&MID($C%d,1,6),'任务单（空sheet）'!$R:$R,0),1)", rIdx, rIdx)
}

var (
	batcnCnvSep          = "\n"
	batchCnvDiseaseTitle = []string{
		"包装疾病分类",
		"基因",
		"疾病",
		"遗传模式",
		"发病年龄",
		"疾病简介",
		"疾病治疗",
		"治疗药物",
		"中国上市",
		"国家医保",
		"出生缺陷救助项目",
		"可及治疗梯队",
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
		addDiseases2Cnv(item, batchCnvDiseaseTitle, batcnCnvSep, genes...)
		writeRow(bcExcel, sheetName, item, BatchCnvTitle, i+2)
	}
	var lastCellName = simpleUtil.HandleError(excelize.CoordinatesToCellName(len(BatchCnvTitle), len(BatchCnv)+1)).(string)
	simpleUtil.CheckErr(bcExcel.AddTable(sheetName, "A1", lastCellName, `{"table_style":"TableStyleMedium9"}`))
	simpleUtil.CheckErr(bcExcel.SaveAs(*prefix + ".batchCNV.xlsx"))
	<-throttle
}
