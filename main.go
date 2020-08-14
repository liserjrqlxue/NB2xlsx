package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize/v2"

	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
)

// os
var (
	ex, _        = os.Executable()
	exPath       = filepath.Dir(ex)
	templatePath = filepath.Join(exPath, "template")
)

// flag
var (
	output = flag.String(
		"output",
		"",
		"output file path",
	)
	template = flag.String(
		"template",
		filepath.Join(templatePath, "NBS-final.result-批次号_产品编号.xlsx"),
		"template to be used",
	)
	avdDataFiles = flag.String(
		"avd",
		"",
		"All variants data file list, comma as sep",
	)
	avdSheetName = flag.String(
		"avdSheetName",
		"All variants data",
		"All variants data sheet name",
	)
)

func main() {
	flag.Parse()
	if *output == "" || *avdDataFiles == "" {
		flag.Usage()
		fmt.Println("-avd is required!")
		os.Exit(1)
	}

	var excel, err1 = excelize.OpenFile(*template)
	simpleUtil.CheckErr(err1)
	var rows, err2 = excel.GetRows(*avdSheetName)
	simpleUtil.CheckErr(err2)
	var avdTitle = rows[0]
	var offset = len(rows)
	var avd, _ = textUtil.Files2MapArray(strings.Split(*avdDataFiles, ","), "\t", nil)
	for i, item := range avd {
		for j, k := range avdTitle {
			var axis, err = excelize.CoordinatesToCellName(j+1, i+1+offset)
			simpleUtil.CheckErr(err)
			simpleUtil.CheckErr(excel.SetCellValue(*avdSheetName, axis, item[k]))
		}
	}
	fmt.Printf("excel.SaveAs(\"%s\")\n", *output)
	simpleUtil.CheckErr(excel.SaveAs(*output))
}
