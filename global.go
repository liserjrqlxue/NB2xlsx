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

var (
	geneListMap        = make(map[string]bool)
	geneSubListMap     = make(map[string]bool)
	geneExcludeListMap = make(map[string]bool)
	geneInfoMap        = make(map[string]map[string]string)
	functionExcludeMap = make(map[string]bool)
	diseaseDb          = make(map[string]map[string]string)
	geneInheritance    = make(map[string]string)
	localDb            = make(map[string]map[string]string)
	dropListMap        = make(map[string][]string)
	genderMap          = make(map[string]string)
	// DmdCnv : array of DMD cnv map
	DmdCnv []map[string]string
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

var err error

var codeKey = "c3d112d6a47a0a04aad2b9d2d2cad266"

// tag

// GeneInfo : struct info of gene
type GeneInfo struct {
	基因                      string
	遗传模式                    string
	性别                      string
	PLP, hetPLP, VUS, HpVUS int
	cnv, cnv0               bool
	tag3                    string
	tag4                    bool
}

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
