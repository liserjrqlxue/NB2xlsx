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
	fillChan(throttle)
}

func loadDmd(excel *excelize.File, dmdArray []string) {
	var sheetName = *dmdSheetName
	if *im {
		sheetName = "DMD CNV"
	}
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
			DmdCnv = append(DmdCnv, item)
			//writeRow(excel, sheetName, item, title, rIdx)
		}
	}
}

func goUpdateCNV(excel *excelize.File, throttle chan bool) {
	if *im {
		*dmdSheetName = "DMD CNV"
	}
	updateData2Sheet(excel, *dmdSheetName, DmdCnv, updateDMDCNV)
	fillChan(throttle)
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
		writeAe(excel, db)
	} else {
		log.Println("Write AE Skip")
	}
	log.Println("Write AE Done")
	fillChan(throttle)
}

func writeAe(excel *excelize.File, db map[string]map[string]string) {
	if *im {
		*aeSheetName = "THAL CNV"
	}
	var rows = simpleUtil.HandleError(excel.GetRows(*aeSheetName)).([][]string)
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
			writeRow(excel, *aeSheetName, item, title, rIdx)
		}
	}
}

// WriteQC write QC sheet to excel
func WriteQC(excel *excelize.File, throttle chan bool) {
	log.Println("Write QC Start")
	if *cs {
		var qcMaps, _ = textUtil.File2MapArray(*qc, "\t", nil)
		writeQC(excel, qcMaps)
	} else {
		writeQC(excel, loadQC(*qc))
	}
	log.Println("Write QC Done")
	emptyChan(throttle)
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

func goWriteBatchCnv(throttle chan bool) {
	var sheetName = "Sheet1"
	var bcExcel = simpleUtil.HandleError(excelize.OpenFile(bcTemplate)).(*excelize.File)

	updateData2Sheet(bcExcel, sheetName, BatchCnv, updateBatchCNV)

	//var lastCellName = simpleUtil.HandleError(excelize.CoordinatesToCellName(len(BatchCnvTitle), len(BatchCnv)+1)).(string)
	//simpleUtil.CheckErr(bcExcel.AddTable(sheetName, "A1", lastCellName, `{"table_style":"TableStyleMedium9"}`), "bcExcel.AddTable Error!")

	simpleUtil.CheckErr(bcExcel.SaveAs(*prefix+".batchCNV.xlsx"), "bcExcel.SaveAs Error!")

	fillChan(throttle)
}

func saveMainExcel(excel *excelize.File, path string, throttle chan bool) {
	log.Printf("excel.SaveAs(\"%s\")\n", path)
	simpleUtil.CheckErr(excel.SaveAs(path))
	log.Println("Save main Done")
	fillChan(throttle)
}
