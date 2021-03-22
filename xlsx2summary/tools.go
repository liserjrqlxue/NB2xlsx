package main

import (
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
)

func WriteCellStr(excel *excelize.File, sheetName string, col, row int, value string) {
	simpleUtil.CheckErr(
		excel.SetCellStr(
			sheetName,
			simpleUtil.HandleError(excelize.CoordinatesToCellName(col, row)).(string),
			value,
		),
	)
}

func WriteCellValue(excel *excelize.File, sheetName string, col, row int, value interface{}) {
	simpleUtil.CheckErr(
		excel.SetCellValue(
			sheetName,
			simpleUtil.HandleError(excelize.CoordinatesToCellName(col, row)).(string),
			value,
		),
	)
}
