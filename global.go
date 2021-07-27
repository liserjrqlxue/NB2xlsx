package main

import (
	"os"
	"path/filepath"
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
