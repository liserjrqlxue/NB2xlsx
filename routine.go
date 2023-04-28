package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
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
func LoadDmd4Sheet(excel *excelize.File, sheetName string, dmdArray []string) (dmdResult []map[string]string) {
	log.Println("Load DMD Start")
	if *dmdFiles != "" {
		dmdArray = strings.Split(*dmdFiles, ",")
	}
	if *dmdList != "" {
		dmdArray = append(dmdArray, textUtil.File2Array(*dmdList)...)
	}
	if len(dmdArray) > 0 {
		dmdResult = loadDmd(excel, sheetName, dmdArray)
	} else {
		log.Println("Load DMD Skip")
	}
	log.Println("Load DMD Done")

	return
}

func loadDmd(excel *excelize.File, sheetName string, dmdArray []string) (dmdResult []map[string]string) {
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	//var title = rows[0]
	var rIdx = len(rows)
	for _, fileName := range dmdArray {
		var dmd, _ = textUtil.File2MapArray(fileName, "\t", nil)
		for _, item := range dmd {
			rIdx++
			updateDmd(item)
			var (
				sampleID = item["#Sample"]
				gene     = item["gene"]
				CN       = strings.Split(item["CopyNum"], ";")[0]
				cn       float64
				err      error
			)
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
			if *im {
				addDatabase2Cnv(item)
				updateColumns(item, sheetTitleMap[sheetName])
			}
			dmdResult = append(dmdResult, item)
			//writeRow(excel, sheetName, item, title, rIdx)
		}
	}
	return
}

// WriteAe write AE sheet to excel
func WriteAe(excel *excelize.File, sheetName string, throttle chan<- bool) {
	log.Println("Write AE Start")
	var db = make(map[string]map[string]string)
	if *dipinResult != "" {
		var dipin, _ = textUtil.File2MapArray(*dipinResult, "\t", nil)
		for _, item := range dipin {
			updateDipin(item, db)
		}
	}
	if *smaResult != "" {
		var smaXlsx = excelize.NewFile()
		var smaTitle = textUtil.File2Array(filepath.Join(etcPath, "title.sma.txt"))
		var sma, _ = textUtil.File2MapArray(*smaResult, "\t", nil)
		writeTitle(smaXlsx, "Sheet1", smaTitle)
		for i, item := range sma {
			updateSma(item, db)
			writeRow(smaXlsx, "Sheet1", item, smaTitle, i+2)
		}
		simpleUtil.CheckErr(smaXlsx.SaveAs(*prefix + ".SMA_result.xlsx"))
	}
	if *sma2Result != "" {
		var sma, _ = textUtil.File2MapArray(*sma2Result, "\t", nil)
		for _, item := range sma {
			updateSma2(item, db)
		}
	}
	if len(db) > 0 {
		writeAe(excel, sheetName, db)
	} else {
		log.Println("Write AE Skip")
	}
	log.Println("Write AE Done")
	holdChan(throttle)
}

func writeAe(excel *excelize.File, sheetName string, db map[string]map[string]string) {
	if *im {
		sheetName = "THAL CNV"
	}
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)
	for _, item := range db {
		rIdx++
		updateAe(item)
		updateINDEX(item, "D", rIdx)
		var sampleID = item["SampleID"]
		item["sampleID"] = sampleID
		if *im {
			updateInfo(item, sampleID)
			updateGender(item, sampleID)
			for _, s := range []string{"THAL CNV", "SMN1 CNV"} {
				updateColumns(item, sheetTitleMap[s])
				writeRow(excel, s, item, sheetTitle[s], rIdx)
			}
		} else {
			if *cs {
				item["sex"] = item["Sex"]
				updateInfo(item, sampleID)
			} else if *wgs {
				updateGender(item, sampleID)
				updateInfo(item, sampleID)
			} else {
				updateABC(item, sampleID)
			}
			writeRow(excel, sheetName, item, title, rIdx)
		}
	}
}

// WriteQC write QC sheet to excel
func WriteQC(excel *excelize.File, sheetName, path string) {
	if path == "" {
		log.Printf("skip [%s] for absence", sheetName)
		return
	}

	log.Println("Write QC Start")
	if *cs {
		var qcMaps, _ = textUtil.File2MapArray(path, "\t", nil)
		writeQC(excel, sheetName, qcMaps)
	} else {
		writeQC(excel, sheetName, loadQC(path))
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

func goWriteBatchCnv(sheetName string, batchCnvDb []map[string]string, throttle chan<- bool) {
	var bcExcel = simpleUtil.HandleError(excelize.OpenFile(bcTemplate)).(*excelize.File)

	writeData2Sheet(bcExcel, sheetName, batchCnvDb, updateBatchCNV)

	simpleUtil.CheckErr(bcExcel.SaveAs(*prefix+".batchCNV.xlsx"), "bcExcel.SaveAs Error!")

	holdChan(throttle)
}
