package main

import (
	"os"
	"path/filepath"
	"regexp"
)

// os
var (
	ex, _        = os.Executable()
	exPath       = filepath.Dir(ex)
	dbPath       = filepath.Join(exPath, "db")
	etcPath      = filepath.Join(exPath, "etc")
	templatePath = filepath.Join(exPath, "template")
)

// template
var (
	// template for prefix+".xlsx"
	mainTemplate = filepath.Join(templatePath, "NBS-final.result-批次号_产品编号.xlsx")
	// template for prefix+".batchCNV.xlsx"
	bcTemplate = filepath.Join(templatePath, "NB2xlsx.batchCNV.xlsx")
)

// etc
var (
	// ACMG2015 db list
	acmgDbList = filepath.Join(etcPath, "acmg.db.list.txt")
	// samples' all snv Excel sheet title
	allColumns = filepath.Join(etcPath, "avd.all.columns.txt")
	// drop list for main excel
	dropList = filepath.Join(etcPath, "drop.list.txt")
	// gene list to filter
	geneList = filepath.Join(etcPath, "gene.list.txt")
	// gene info : Transcript and EntrezID
	geneInfoList = filepath.Join(etcPath, "gene.info.txt")
	// exclude function list
	functionExclude = filepath.Join(etcPath, "function.exclude.txt")
	// i18n.txt
	i18nTxt = filepath.Join(etcPath, "i18n.txt")
	// variant db
	jsonAes   = filepath.Join(etcPath, "已解读数据库.json.aes")
	jsonAesIM = filepath.Join(etcPath, "已解读数据库.IM.json.aes")
)

var (
	geneInfoMap        = make(map[string]map[string]string)
	geneListMap        = make(map[string]bool)
	geneIMListMap      = make(map[string]bool)
	geneSubListMap     = make(map[string]bool)
	geneExcludeListMap = make(map[string]bool)

	functionExcludeMap = make(map[string]bool)

	diseaseSep = "$$"
	diseaseDb  = make(map[string]map[string]string)

	geneInheritance = make(map[string]string)
	localDb         = make(map[string]map[string]string)
	cnvDb           = make(map[string]map[string]string)
	dropListMap     = make(map[string][]string)
	genderMap       = make(map[string]string)

	// DmdCnv : array of DMD cnv map
	DmdCnv []map[string]string

	// SampleGeneInfo : sampleID -> GeneSymbol -> *GeneInfo
	SampleGeneInfo = make(map[string]map[string]*GeneInfo)
	limsInfo       = make(map[string]map[string]string)
	imInfo         map[string]map[string]string

	columnName    = "字段-中心实验室"
	sheetTitle    = make(map[string][]string)
	sheetTitleMap = make(map[string]map[string]string)

	formalStyleID, supplementaryStyleID int
	//checkStyleID int
	formalCheckStyleID, supplementaryCheckStyleID int
)

var sampleDetail = make(map[string]map[string]string)
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

var codeKey = "c3d112d6a47a0a04aad2b9d2d2cad266"

// tag

var tag7gene = make(map[string]bool)

//LOFofPLP : Lost Of Function for PLP
var LOFofPLP = map[string]bool{
	"nonsense":   true,
	"frameshift": true,
	"stop-gain":  true,
	"span":       true,
	"altstart":   true,
	"init-loss":  true,
	"splice-3":   true,
	"splice-5":   true,
}

var cdsList = map[string]bool{
	"cds-del":   true,
	"cds-ins":   true,
	"cds-indel": true,
	"stop-loss": true,
}

var spliceList = map[string]bool{
	"splice+10": true,
	"splice-10": true,
	"splice+20": true,
	"splice-20": true,
	"intron":    true,
}

var spliceCSList = map[string]bool{
	"splice+10":    true,
	"splice-10":    true,
	"splice+20":    true,
	"splice-20":    true,
	"intron":       true,
	"coding-synon": true,
}

var (
	isAD          = regexp.MustCompile(`AD`)
	isXLD         = regexp.MustCompile(`XLD`)
	isP           = regexp.MustCompile(`P`)
	isI           = regexp.MustCompile(`I`)
	isD           = regexp.MustCompile(`D`)
	isNeutral     = regexp.MustCompile(`neutral`)
	isDeleterious = regexp.MustCompile(`deleterious`)
	isPLPVUS      = regexp.MustCompile(`^P|^LP|^VUS`)

	af0List = map[string]bool{
		"ESP6500 AF":    true,
		"1000G AF":      true,
		"ExAC AF":       true,
		"ExAC EAS AF":   true,
		"GnomAD AF":     true,
		"GnomAD EAS AF": true,
	}
	afThreshold = 1e-4
)

// CS
var (
	top1kGene        = make(map[string]bool)
	repeatRegion     []*Region
	homologousRegion []*Region
)

// IM
var (
	i18n string
	I18n = make(map[string]map[string]string)

	thalNameMap map[string]map[string]string

	imSheetList = []string{
		"Sample",
		"QC",
		"SNV&INDEL",
		"DMD CNV",
		"THAL CNV",
		"SMN1 CNV",
	}
)

// HyperLink
var hyperLinkTitle = map[string]bool{
	"β地贫_chr11": true,
	"α地贫_chr16": true,

	"β地贫_最终结果": true,
	"α地贫_最终结果": true,

	"reads_picture": true,

	"P0": true,
	"P1": true,
	"P2": true,
	"P3": true,
}

// float format
var afFloatFormatArray = []string{
	"GnomAD AF",
	"GnomAD EAS AF",
	"ExAC AF",
	"ExAC EAS AF",
	"1000G AF",
	"1000G EAS AF",
	"ESP6500 AF",
	"PVFD AF",
	"dbSNP Allele Freq",
}

var qcFloatFormatArray = []string{
	"Q20(%)",
	"Q30(%)",
	"AverageDepth",
	"Depth>=30(%)",
	"GC(%)",
}

var deafnessGeneList = make(map[string]bool)

var exonCount map[string]string
