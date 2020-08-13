package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/goUtil/xlsxUtil"
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

	var templateXlsx = xlsxUtil.OpenFile(*template)
	var excel = xlsxUtil.NewFile()
	for _, sheet := range templateXlsx.File.Sheets {
		excel.AppendSheet(*sheet, sheet.Name)
	}

	var sheet = excel.File.Sheet[*avdSheetName]
	var avdTitle = xlsxUtil.GetRowArray(0, sheet)
	fmt.Printf("%+v\n", avdTitle)
	var avd, _ = textUtil.Files2MapArray(strings.Split(*avdDataFiles, ","), "\t", nil)
	for _, item := range avd {
		xlsxUtil.AddMap2Row(item, avdTitle, sheet.AddRow())
	}
}
