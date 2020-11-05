package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/liserjrqlxue/goUtil/osUtil"

	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/version"
)

// os
var (
	ex, _        = os.Executable()
	exPath       = filepath.Dir(ex)
	dbPath       = filepath.Join(exPath, "db")
	etcPath      = filepath.Join(exPath, "etc")
	templatePath = filepath.Join(exPath, "template")
)

// flag
var (
	prefix = flag.String(
		"prefix",
		"",
		"output to -prefix.xlsx",
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
		filepath.Join(etcPath, "疾病简介和治疗-20200925.xlsx"),
		"disease database excel",
	)
	diseaseSheetName = flag.String(
		"diseaseSheetName",
		"Sheet2",
		"sheet name of disease database excel",
	)
	localDbExcel = flag.String(
		"db",
		filepath.Join(etcPath, "已解读数据库-2020.09.10.xlsx"),
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
)

var (
	geneListMap        = make(map[string]bool)
	functionExcludeMap = make(map[string]bool)
	diseaseDb          = make(map[string]map[string]string)
	localDb            = make(map[string]map[string]string)
	dropListMap        = make(map[string][]string)
	genderMap          = make(map[string]string)
)

func main() {
	version.LogVersion()
	// flag
	flag.Parse()
	if *prefix == "" {
		flag.Usage()
		log.Println("-prefix are required!")
		os.Exit(1)
	}
	var (
		runAe  = make(chan bool, 1)
		runAvd = make(chan bool, 1)
		runDmd = make(chan bool, 1)
	)

	if osUtil.FileExists(*gender) {
		genderMap = simpleUtil.HandleError(textUtil.File2Map(*gender, "\t", false)).(map[string]string)
	}

	loadDb()

	var excel = simpleUtil.HandleError(excelize.OpenFile(*template)).(*excelize.File)

	SampleGeneInfo = make(map[string]map[string]*GeneInfo)

	// CNV
	{
		runDmd <- true
		go WriteDmd(excel, *dmdFiles, *dmdList, runDmd)
	}

	// 补充实验
	{
		runAe <- true
		go WriteAe(excel, *dipinResult, *smaResult, runAe)
	}

	// All variant data
	{
		runAvd <- true
		go WriteAvd(excel, runDmd, runAvd)
	}

	// drug
	if *drugResult != "" {
		log.Println("Start load Drug")
		var drugDb = make(map[string]map[string]map[string]string)
		var drug, _ = textUtil.File2MapArray(*drugResult, "\t", nil)
		for _, item := range drug {
			var sampleID = item["样本编号"]
			var drugName = item["药物名称"]
			var sampleDrug, ok1 = drugDb[sampleID]
			if !ok1 {
				sampleDrug = make(map[string]map[string]string)
				drugDb[sampleID] = sampleDrug
			}
			var drugInfo, ok2 = sampleDrug[drugName]
			if !ok2 {
				drugInfo = item
				sampleDrug[drugName] = drugInfo
				drugInfo["SampleID"] = sampleID
			} else {

			}
		}

	}

	// wait done
	runAe <- true
	runAvd <- true

	log.Printf("excel.SaveAs(\"%s\")\n", *prefix+".xlsx")
	simpleUtil.CheckErr(excel.SaveAs(*prefix + ".xlsx"))
	log.Println("Done")
}
