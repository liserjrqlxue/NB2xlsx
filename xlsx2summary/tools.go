package main

import (
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
)

func WriteCellInt(excel *excelize.File, sheetName string, col, row int, value int) {
	simpleUtil.CheckErr(
		excel.SetCellInt(
			sheetName,
			simpleUtil.HandleError(excelize.CoordinatesToCellName(col, row)).(string),
			value,
		),
	)
}

func WriteCellStr(excel *excelize.File, sheetName string, col, row int, value string) {
	simpleUtil.CheckErr(
		excel.SetCellStr(
			sheetName,
			simpleUtil.HandleError(excelize.CoordinatesToCellName(col, row)).(string),
			value,
		),
	)
}
