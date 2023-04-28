package main

import (
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/xuri/excelize/v2"
	"log"
	"path/filepath"
)

func createExcelIM(sheetNames []string) *excelize.File {
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

func createExcel(mode string) *excelize.File {
	switch mode {
	case "NBSIM":
		return createExcelIM(imSheetList)
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

func initExcel(excel *excelize.File, mode string) {
	excel = createExcel(mode)
	styleInit(excel)
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
