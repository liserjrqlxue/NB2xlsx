package main

import (
	"path/filepath"

	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/xuri/excelize/v2"
)

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
