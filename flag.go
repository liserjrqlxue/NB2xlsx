package main

import (
	"flag"
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
