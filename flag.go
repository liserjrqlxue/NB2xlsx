package main

import (
	"flag"
	"path/filepath"
)

// flag
var (
	batch = flag.String(
		"batch",
		"",
		"batch name",
	)
	prefix = flag.String(
		"prefix",
		"",
		"output to -prefix.xlsx,default is -batch.xlsx",
	)
	template = flag.String(
		"template",
		filepath.Join(templatePath, "NBS-final.result-批次号_产品编号.xlsx"),
		"template to be used",
	)
	dropList = flag.String(
		"dropList",
		filepath.Join(etcPath, "drop.list.txt"),
		"drop list for excel",
	)
	avdList = flag.String(
		"avdList",
		"",
		"All variants data file list, one path per line",
	)
	avdFiles = flag.String(
		"avd",
		"",
		"All variants data file list, comma as sep",
	)
	avdSheetName = flag.String(
		"avdSheetName",
		"All variants data",
		"All variants data sheet name",
	)
	diseaseExcel = flag.String(
		"disease",
		filepath.Join(etcPath, "新生儿疾病库.xlsx"),
		"disease database excel",
	)
	diseaseSheetName = flag.String(
		"diseaseSheetName",
		"Sheet1",
		"sheet name of disease database excel",
	)
	localDbExcel = flag.String(
		"db",
		filepath.Join(etcPath, "已解读数据库.xlsx"),
		"已解读数据库",
	)
	localDbSheetName = flag.String(
		"dbSheetName",
		"Sheet1",
		"已解读数据库 sheet name",
	)
	acmgDb = flag.String(
		"acmgDb",
		filepath.Join(etcPath, "acmg.db.list.txt"),
		"acmg db list",
	)
	autoPVS1 = flag.Bool(
		"autoPVS1",
		false,
		"is use autoPVS1",
	)
	dmdFiles = flag.String(
		"dmd",
		"",
		"DMD result file list, comma as sep",
	)
	dmdList = flag.String(
		"dmdList",
		"",
		"DMD result file list, one path per line",
	)
	dmdSheetName = flag.String(
		"dmdSheetName",
		"CNV",
		"DMD result sheet name",
	)
	dipinResult = flag.String(
		"dipin",
		"",
		"dipin result file",
	)
	smaResult = flag.String(
		"sma",
		"",
		"sma result file",
	)
	aeSheetName = flag.String(
		"aeSheetName",
		"补充实验",
		"Additional Experiments sheet name",
	)
	drugResult = flag.String(
		"drug",
		"",
		"drug result file",
	)
	drugSheetName = flag.String(
		"drugSheetName",
		"药物",
		"drug sheet name",
	)
	geneList = flag.String(
		"geneList",
		filepath.Join(etcPath, "gene.list.txt"),
		"gene list to filter",
	)
	functionExclude = flag.String(
		"functionExclude",
		filepath.Join(etcPath, "function.exclude.txt"),
		"function list to exclude",
	)
	allSheetName = flag.String(
		"allSheetName",
		"Sheet1",
		"all snv sheet name",
	)
	allColumns = flag.String(
		"allColumns",
		filepath.Join(etcPath, "avd.all.columns.txt"),
		"all snv sheet title",
	)
	gender = flag.String(
		"gender",
		"F",
		"gender for all or gender map file",
	)
	threshold = flag.Int(
		"threshold",
		12,
		"threshold limit",
	)
	batchCNV = flag.String(
		"batchCNV",
		"",
		"batchCNV result",
	)
	all = flag.Bool(
		"all",
		false,
		"if output all snv",
	)
	lims = flag.String(
		"lims",
		"",
		"lims.info",
	)
	qc = flag.String(
		"qc",
		"",
		"qc excel",
	)
	qcTitle = flag.String(
		"qcTitle",
		filepath.Join(etcPath, "QC.txt"),
		"qc title map",
	)
	qcSheetName = flag.String(
		"qcSheet",
		"QC",
		"qc sheet name",
	)
	bamPath = flag.String(
		"bamPath",
		"",
		"bamList file",
	)
	bamPathSheetName = flag.String(
		"bamPathSheetName",
		"bam文件路径",
		"bamPath sheet name",
	)
)