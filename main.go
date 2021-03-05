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
		filepath.Join(etcPath, "孕159疾病-20201.1.14-莹硕.xlsx"),
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

var (
	geneListMap        = make(map[string]bool)
	functionExcludeMap = make(map[string]bool)
	diseaseDb          = make(map[string]map[string]string)
	geneInheritance    = make(map[string]string)
	localDb            = make(map[string]map[string]string)
	dropListMap        = make(map[string][]string)
	genderMap          = make(map[string]string)
	// BatchCnv : array of batch cnv map
	BatchCnv []map[string]string
	// BatchCnvTitle : titles of batch cnv
	BatchCnvTitle []string
	// SampleGeneInfo : sampleID -> GeneSymbol -> *GeneInfo
	SampleGeneInfo                      = make(map[string]map[string]*GeneInfo)
	limsInfo                            map[string]map[string]string
	qcMap                               map[string]string
	formalStyleID, supplementaryStyleID int
	//checkStyleID int
	formalCheckStyleID, supplementaryCheckStyleID int
)

type SampleInfo struct {
	sampleID string
	gender   string
	p0       string
	p1       string
	p2       string
	p3       string
}

func newSampleInfo(item map[string]string) SampleInfo {
	return SampleInfo{
		sampleID: item["sampleID"],
		gender:   item["gender_analysed"],
		p0:       item["P0"],
		p1:       item["P1"],
		p2:       item["P2"],
		p3:       item["P3"],
	}
}

var sampleInfos = make(map[string]SampleInfo)

var colorRGB map[string]string
var (
	// 验证位点
	checkColor string
	// 正式报告
	formalRreportColor string
	// 补充报告
	supplementaryReportColor string
)

var err error

func main() {
	version.LogVersion()
	// flag
	flag.Parse()
	if *prefix == "" {
		if *batch != "" {
			*prefix = *batch
		} else {
			flag.Usage()
			log.Println("-prefix are required!")
			os.Exit(1)
		}
	}

	if osUtil.FileExists(*gender) {
		log.Printf("load gender map from %s", *gender)
		genderMap = simpleUtil.HandleError(textUtil.File2Map(*gender, "\t", false)).(map[string]string)
	}

	var (
		localDb      = make(chan bool, 1)
		runAe        = make(chan bool, 1)
		runAvd       = make(chan bool, 1)
		runDmd       = make(chan bool, 1)
		runQC        = make(chan bool, 1)
		saveMain     = make(chan bool, 1)
		saveBatchCnv = make(chan bool, 1)
	)

	// load local db
	{
		localDb <- true
		loadLocalDb(localDb)
	}

	loadDb()

	limsInfo = loadLimsInfo()

	if *batchCNV != "" {
		loadBatchCNV(*batchCNV)
	}

	var excel = simpleUtil.HandleError(excelize.OpenFile(*template)).(*excelize.File)
	styleInit(excel)

	if *bamPath != "" {
		for i, path := range textUtil.File2Array(*bamPath) {
			var axis = simpleUtil.HandleError(excelize.CoordinatesToCellName(1, i+1)).(string)
			simpleUtil.CheckErr(
				excel.SetCellStr(
					*bamPathSheetName,
					axis,
					path,
				),
			)
		}
	}

	// QC
	if *qc != "" {
		runQC <- true
		WriteQC(excel, runQC)
	}

	// CNV
	// QC -> DMD
	{
		runQC <- true
		runDmd <- true
		WriteDmd(excel, runDmd)
	}

	// 补充实验
	{
		runAe <- true
		WriteAe(excel, runAe)
	}

	// All variant data
	{
		localDb <- true
		runAvd <- true
		WriteAvd(excel, runDmd, runAvd, *all)
	}

	// drug
	if *drugResult != "" {
		log.Println("Start load Drug")
		var drugDb = make(map[string]map[string]map[string]string)
		var drug, _ = textUtil.File2MapArray(*drugResult, "\t", nil)
		for _, item := range drug {
			var sampleID = item["样本编号"]
			item["SampleID"] = sampleID
			updateABC(item)
			var drugName = item["药物名称"]
			var sampleDrug, ok1 = drugDb[sampleID]
			if !ok1 {
				sampleDrug = make(map[string]map[string]string)
				drugDb[sampleID] = sampleDrug
			}
			if _, ok := sampleDrug[drugName]; !ok {
				item["SampleID"] = sampleID
				sampleDrug[drugName] = item
			}
		}
	}

	{
		runAvd <- true
		saveBatchCnv <- true
		go writeBatchCnv(saveBatchCnv)
	}

	{
		runAe <- true
		saveMain <- true
		go func() {
			log.Printf("excel.SaveAs(\"%s\")\n", *prefix+".xlsx")
			simpleUtil.CheckErr(excel.SaveAs(*prefix + ".xlsx"))
			log.Println("Save main Done")
			<-saveMain
		}()
	}

	// waite excel write done
	{
		saveMain <- true
		saveBatchCnv <- true
	}
	log.Println("All Done")
}
