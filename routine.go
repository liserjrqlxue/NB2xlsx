package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/xuri/excelize/v2"
)

func addChr(chr string) string {
	return "Chr" + strings.Replace(
		strings.Replace(chr, "chr", "", 1),
		"Chr", "", 1,
	)
}

// LoadDmd4Sheet load DMD sheet to excel
func LoadDmd4Sheet(excel *excelize.File, sheetName string, mode Mode, dmdArray []string) (dmdResult []map[string]string) {
	log.Println("Load DMD Start")
	if len(dmdArray) > 0 {
		dmdResult = loadDmd(excel, sheetName, mode, dmdArray)
	} else {
		log.Println("Load DMD Skip")
	}
	log.Println("Load DMD Done")

	return
}

func loadDmd(excel *excelize.File, sheetName string, mode Mode, dmdArray []string) (dmdResult []map[string]string) {
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	//var title = rows[0]
	var rIdx = len(rows)
	for _, fileName := range dmdArray {
		var dmd, _ = textUtil.File2MapArray(fileName, "\t", nil)
		for _, item := range dmd {
			rIdx++
			updateDmd(item, mode)
			dmdResult = append(dmdResult, item)
		}
	}
	return
}

// WriteAe write AE sheet to excel
func WriteAe(excel *excelize.File, sheetName, dipinResult, smaResult string, mode Mode, throttle chan<- bool) {
	log.Println("Write AE Start")
	var db = make(map[string]map[string]string)
	if dipinResult != "" {
		var dipin, _ = textUtil.File2MapArray(dipinResult, "\t", nil)
		for _, item := range dipin {
			updateDipin(item, db, mode)
		}
	}
	if smaResult != "" {
		var smaXlsx = excelize.NewFile()
		var smaTitle = textUtil.File2Array(titleSMA)
		var sma, _ = textUtil.File2MapArray(smaResult, "\t", nil)
		writeTitle(smaXlsx, "Sheet1", smaTitle)
		for i, item := range sma {
			updateSma(item, db, mode)
			writeRow(smaXlsx, "Sheet1", item, smaTitle, i+2, mode)
		}
		simpleUtil.CheckErr(smaXlsx.SaveAs(*prefix + ".SMA_result.xlsx"))
	}
	if *sma2Result != "" {
		var sma, _ = textUtil.File2MapArray(*sma2Result, "\t", nil)
		for _, item := range sma {
			updateSma2(item, db, mode)
		}
	}
	if len(db) > 0 {
		writeAe(excel, sheetName, mode, db)
	} else {
		log.Println("Write AE Skip")
	}
	log.Println("Write AE Done")
	holdChan(throttle)
}

func writeAe(excel *excelize.File, sheetName string, mode Mode, db map[string]map[string]string) {
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)
	for _, item := range db {
		rIdx++
		updateAe(item, mode)
		updateINDEX(item, "D", rIdx)
		var sampleID = item["SampleID"]
		item["sampleID"] = sampleID
		switch mode {
		case NBSP:
			updateABC(item, sampleID)
			writeRow(excel, sheetName, item, title, rIdx, mode)
		case NBSIM:
			updateInfo(item, sampleID, mode)
			updateGender(item, sampleID)
			for _, s := range []string{"THAL CNV", "SMN1 CNV"} {
				updateColumns(item, sheetTitleMap[s])
				writeRow(excel, s, item, sheetTitle[s], rIdx, mode)
			}
		case WGSNB:
			updateGender(item, sampleID)
			updateInfo(item, sampleID, mode)
			writeRow(excel, sheetName, item, title, rIdx, mode)
		case WGSCS:
			item["sex"] = item["Sex"]
			updateInfo(item, sampleID, mode)
			writeRow(excel, sheetName, item, title, rIdx, mode)
		}
	}
}

// WriteQC write QC sheet to excel
func WriteQC(excel *excelize.File, sheetName, path string, mode Mode) {
	if path == "" {
		log.Printf("skip [%s] for absence", sheetName)
		return
	}

	log.Println("Write QC Start")
	if mode == WGSCS {
		var qcMaps, _ = textUtil.File2MapArray(path, "\t", nil)
		writeQC(excel, sheetName, mode, qcMaps)
	} else {
		writeQC(excel, sheetName, mode, loadQC(path))
	}
	log.Println("Write QC Done")
}

func updateINDEX(item map[string]string, col string, index int) {
	item["解读人"] = fmt.Sprintf("=INDEX(任务单!O:O,MATCH(%s%d,任务单!I:I,0),1)", col, index)
	item["审核人"] = fmt.Sprintf("=INDEX(任务单!P:P,MATCH(%s%d,任务单!I:I,0),1)", col, index)
}

var (
	multiDiseaseSep = "\n"
	//batchCnvDiseaseTitle = []string{
	//	"疾病中文名",
	//	"遗传模式",
	//}
)
