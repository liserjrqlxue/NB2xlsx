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
	threshold = flag.Int(
		"threshold",
		12,
		"threshold limit",
	)
)

// input
var (
	// sample info
	detail = flag.String(
		"detail",
		"",
		"sample info",
	)
	gender = flag.String(
		"gender",
		"F",
		"gender for all or gender map file",
	)
	info = flag.String(
		"info",
		"",
		"im info.txt",
	)
	lims = flag.String(
		"lims",
		"",
		"lims.info",
	)
	// all variants data
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
	// bam文件路径
	bamPath = flag.String(
		"bamPath",
		"",
		"bamList file",
	)
	// batchCNV
	batchCNV = flag.String(
		"batchCNV",
		"",
		"batchCNV result",
	)
	// 补充实验
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
	sma2Result = flag.String(
		"sma2",
		"",
		"sma result file",
	)
	// DMD CNV
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
	lumpy = flag.String(
		"lumpy",
		"",
		"DMD-lumpy data",
	)
	nator = flag.String(
		"nator",
		"",
		"DMD-nator data",
	)
	// 药物检测结果
	drugResult = flag.String(
		"drug",
		"",
		"drug result file",
	)
	// 个特
	featureList = flag.String(
		"feature",
		"",
		"个特 list",
	)
	// 基因ID
	geneIDList = flag.String(
		"geneID",
		"",
		"基因ID list",
	)
	// QC
	qc = flag.String(
		"qc",
		"",
		"qc excel",
	)
)

// output
var (
	prefix = flag.String(
		"prefix",
		"",
		"output to -prefix.xlsx,default is -batch.xlsx",
	)
	annoDir = flag.String(
		"annoDir",
		"",
		"output sample annotation to annoDir/[sampleID]_vcfanno.xlsx for CS",
	)

	// output sheet name
	aeSheetName = flag.String(
		"aeSheetName",
		"补充实验",
		"Additional Experiments sheet name",
	)
	allSheetName = flag.String(
		"allSheetName",
		"Sheet1",
		"all snv sheet name",
	)
	avdSheetName = flag.String(
		"avdSheetName",
		"All variants data",
		"All variants data sheet name",
	)
	bamPathSheetName = flag.String(
		"bamPathSheetName",
		"bam文件路径",
		"bamPath sheet name",
	)
	dmdSheetName = flag.String(
		"dmdSheetName",
		"CNV",
		"DMD result sheet name",
	)
	drugSheetName = flag.String(
		"drugSheetName",
		"药物检测结果",
		"drug sheet name",
	)
	qcSheetName = flag.String(
		"qcSheet",
		"QC",
		"qc sheet name",
	)
)

// config file
var (
	// etc
	acmgDb = flag.String(
		"acmgDb",
		filepath.Join(etcPath, "acmg.db.list.txt"),
		"acmg db list",
	)
	allColumns = flag.String(
		"allColumns",
		filepath.Join(etcPath, "avd.all.columns.txt"),
		"all snv sheet title",
	)
	dropList = flag.String(
		"dropList",
		filepath.Join(etcPath, "drop.list.txt"),
		"drop list for excel",
	)
	geneList = flag.String(
		"geneList",
		filepath.Join(etcPath, "gene.list.txt"),
		"gene list to filter",
	)
	geneInfoList = flag.String(
		"geneInfoList",
		filepath.Join(etcPath, "gene.info.txt"),
		"gene info:Transcript and EntrezID",
	)
	functionExclude = flag.String(
		"functionExclude",
		filepath.Join(etcPath, "function.exclude.txt"),
		"function list to exclude",
	)
	// template
	bcTemplate = flag.String(
		"bcTemplate",
		filepath.Join(templatePath, "NB2xlsx.batchCNV.xlsx"),
		"template to be used",
	)
	template = flag.String(
		"template",
		filepath.Join(templatePath, "NBS-final.result-批次号_产品编号.xlsx"),
		"template to be used",
	)
)

// boolean
var (
	acmg = flag.Bool(
		"acmg",
		false,
		"if re-calculate ACMG2015",
	)
	all = flag.Bool(
		"all",
		false,
		"if output all snv",
	)
	autoPVS1 = flag.Bool(
		"autoPVS1",
		false,
		"is use autoPVS1",
	)
	cs = flag.Bool(
		"cs",
		false,
		"if use for CS",
	)
	im = flag.Bool(
		"im",
		false,
		"if use for im",
	)
	wgs = flag.Bool(
		"wgs",
		false,
		"if use for wgs",
	)
)
