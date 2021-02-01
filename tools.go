package main

import (
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
)

func loadDb() {
	log.Println("Load Database Start")
	// load gene list
	for _, key := range textUtil.File2Array(*geneList) {
		geneListMap[key] = true
	}
	// load function exclude list
	for _, key := range textUtil.File2Array(*functionExclude) {
		functionExcludeMap[key] = true
	}

	// load disease database
	log.Println("Load Disease Start")
	diseaseDb, _ = simpleUtil.Slice2MapMapArrayMerge(
		simpleUtil.HandleError(
			simpleUtil.HandleError(
				excelize.OpenFile(*diseaseExcel),
			).(*excelize.File).
				GetRows(*diseaseSheetName),
		).([][]string),
		"基因",
		"/",
	)
	for gene, info := range diseaseDb {
		geneInheritance[gene] = info["遗传模式"]
	}

	// load drop list
	log.Println("Load DropList Start")
	for k, v := range simpleUtil.HandleError(textUtil.File2Map(*dropList, "\t", false)).(map[string]string) {
		dropListMap[k] = strings.Split(v, ",")
	}
	log.Println("Load Database Done")
}

func loadLocalDb(throttle chan bool) {
	// load 已解读数据库
	log.Println("Load LocalDb Start")
	localDb, _ = simpleUtil.Slice2MapMapArray(
		simpleUtil.HandleError(
			simpleUtil.HandleError(
				excelize.OpenFile(*localDbExcel),
			).(*excelize.File).
				GetRows(*localDbSheetName),
		).([][]string),
		"Transcript", "cHGVS",
	)
	log.Println("Load LocalDb Done")
	<-throttle
}

var formulaTitle = map[string]bool{
	"解读人": true,
	"审核人": true,
}

var hyperLinkTitle = map[string]bool{
	"β地贫_最终结果": true,
	"α地贫_最终结果": true,
}

var (
	isClinVar = map[string]bool{
		"Pathogenic":                   true,
		"Likely_pathogenic":            true,
		"Pathogenic/Likely_pathogenic": true,
	}
	notClinVar = map[string]bool{
		"Benign":               true,
		"Likely_benign":        true,
		"Benign/Likely_benign": true,
	}
	notClinVar2 = map[string]bool{
		"Benign":                 true,
		"Likely_benign":          true,
		"Benign/Likely_benign":   true,
		"Uncertain_significance": true,
		"Conflicting_interpretations_of_pathogenicity": true,
	}
	isHGMD = map[string]bool{
		"DM":     true,
		"DM?":    true,
		"DM/DM?": true,
	}
)

func gt(s string, tv float64) bool {
	var af, err = strconv.ParseFloat(s, 64)
	if err == nil && af > tv {
		return true
	}
	return false
}

var avdAfList = []string{
	"ESP6500 AF",
	"1000G AF",
	"ExAC AF",
	"GnomAD AF",
	"ExAC EAS AF",
	"GnomAD EAS AF",
}

func filterAvd(item map[string]string) bool {
	var mainKey = item["Transcript"] + "\t" + item["cHGVS"]
	if _, ok := localDb[mainKey]; ok {
		return true
	}
	if !geneListMap[item["Gene Symbol"]] {
		return false
	}
	for _, af := range avdAfList {
		if gt(item[af], 0.05) {
			return false
		}
	}
	if isClinVar[item["ClinVar Significance"]] {
		return true
	}
	if item["ACMG"] == "B" || item["ACMG"] == "LB" {
		return false
	}
	if isHGMD[item["HGMD Pred"]] {
		return true
	}
	for _, af := range avdAfList {
		if gt(item[af], 0.01) {
			return false
		}
	}
	if item["Function"] == "intron" && item["SpliceAI Pred"] == "D" {
		return true
	}
	if functionExcludeMap[item["Function"]] {
		return false
	}
	return true
}

// LOF : Lost Of Function
var LOF = map[string]bool{
	"nonsense":   true,
	"frameshift": true,
	"splice-3":   true,
	"splice-5":   true,
}

func updateLOF(item map[string]string) {
	if !LOF[item["Function"]] || gt(item["GnomAD AF"], 0.01) || gt(item["1000G AF"], 0.01) {
		item["LOF"] = "NO"
	} else {
		item["LOF"] = "YES"
	}
}

func updateDisease(item map[string]string) {
	var gene = item["Gene Symbol"]
	var disease, ok = diseaseDb[gene]
	if ok {
		item["疾病中文名"] = disease["疾病"]
		item["遗传模式"] = disease["遗传模式"]
		item["疾病简介"] = disease["疾病简介"]
	}
}

func addDiseases2Cnv(item map[string]string, sep string, genes ...string) {
	var diseaseCN, inherit []string
	for _, gene := range genes {
		var info = diseaseDb[gene]
		diseaseCN = append(diseaseCN, info["疾病"])
		inherit = append(inherit, info["遗传模式"])
	}
	item["疾病中文名"] = strings.Join(diseaseCN, sep)
	item["遗传模式"] = strings.Join(inherit, sep)
}

var afList = []string{
	"1000Gp3 AF",
	"1000Gp3 EAS AF",
	"GnomAD EAS HomoAlt Count",
	"GnomAD EAS AF",
	"GnomAD HomoAlt Count",
	"GnomAD AF",
	"ExAC EAS AF",
	"ExAC HomoAlt Count",
	"ExAC AF",
	"1000G AF",
	"1000G EAS AF",
	"ExAC EAS HomoAlt Count",
}

func updateAf(item map[string]string) {
	for _, af := range afList {
		if item[af] == "-1" || item[af] == "-1.0" {
			item[af] = "."
		}
	}
}

func updateAvd(item map[string]string) {
	updateABC(item)
	item["ClinVar星级"] = item["ClinVar Number of gold stars"]
	item["1000Gp3 AF"] = item["1000G AF"]
	item["1000Gp3 EAS AF"] = item["1000G EAS AF"]
	var gene = item["Gene Symbol"]
	item["gene+cHGVS"] = gene + ":" + item["cHGVS"]
	item["gene+pHGVS3"] = gene + ":" + item["pHGVS3"]
	item["gene+pHGVS1"] = gene + ":" + item["pHGVS1"]
	anno.Score2Pred(item)
	updateLOF(item)
	updateDisease(item)
	var mainKey = item["Transcript"] + "\t" + item["cHGVS"]
	var db, ok = localDb[mainKey]
	if ok {
		if db["是否是包装位点"] == "是" {
			item["Database"] = "NBS-in"
			item["报告类别"] = "正式报告"
		} else {
			item["Database"] = "NBS-out"
		}
		item["参考文献"] = db["Reference"]
		item["位点关联疾病"] = db["Disease"]
		item["位点关联遗传模式"] = db["遗传模式"]
		item["Evidence New + Check"] = db["证据项"]
		item["Definition"] = db["Definition"]
	} else {
		item["Database"] = "."
		if item["LOF"] == "YES" {
			item["报告类别"] = "补充报告"
		}
	}
	anno.UpdateFunction(item)
	acmg2015.AddEvidences(item)
	item["自动化判断"] = acmg2015.PredACMG2015(item, *autoPVS1)
	anno.UpdateAutoRule(item)
	if filterAvd(item) {
		item["filterAvd"] = "Y"
	}
	updateAf(item)
}

func updateFromAvd(item, geneHash map[string]string, geneInfo map[string]*GeneInfo, sampleID string) {
	if item["filterAvd"] != "Y" {
		return
	}
	var info, ok = geneInfo[item["Gene Symbol"]]
	if !ok {
		info = new(GeneInfo).new(item)
		geneInfo[item["Gene Symbol"]] = info
	} else {
		info.count(item)
	}
	if *gender == "M" || genderMap[sampleID] == "M" {
		item["Sex"] = "M"
		info.性别 = "M"
		UpdateGeneHash(geneHash, item, "M")
	} else if *gender == "F" || genderMap[sampleID] == "F" {
		item["Sex"] = "F"
		info.性别 = "F"
		UpdateGeneHash(geneHash, item, "F")
	}
}

func updateGeneHashAR(geneHash, item map[string]string, gene, genePred string) string {
	if item["Zygosity"] == "Hom" {
		return "可能患病"
	} else if item["Zygosity"] == "Het" {
		if genePred == "" {
			return "携带者"
		} else if genePred == "携带者" {
			return "可能患病"
		}
	}
	return ""
}

func updateGeneHashAD(geneHash, item map[string]string, gene string) string {
	if item["Zygosity"] == "Hom" || item["Zygosity"] == "Het" {
		return "可能患病"
	}
	return ""
}

func updateGeneHashXL(geneHash, item map[string]string, gene, gender string) string {
	if gender == "M" {
		if item["Zygosity"] == "Hemi" {
			return "可能患病"
		}
	} else if gender == "F" {
		if item["Zygosity"] == "Hom" || item["Zygosity"] == "Het" {
			return "可能患病"
		}
	}
	return ""
}

func updateGeneHash(geneHash, item map[string]string, gene, genePred, gender string) string {
	switch item["遗传模式"] {
	case "AR":
		return updateGeneHashAR(geneHash, item, gene, genePred)
	case "AR/AR":
		return updateGeneHashAR(geneHash, item, gene, genePred)
	case "AD":
		updateGeneHashAD(geneHash, item, gene)
	case "AD,AR":
		updateGeneHashAD(geneHash, item, gene)
	case "XL":
		updateGeneHashXL(geneHash, item, gene, gender)
	}
	return ""
}

// UpdateGeneHash : update geneHash
func UpdateGeneHash(geneHash, item map[string]string, gender string) {
	if item["Definition"] != "P" && item["Definition"] != "LP" {
		return
	}
	var gene = item["Gene Symbol"]
	var genePred, ok = geneHash[gene]
	if !ok || genePred != "可能患病" {
		geneHash[gene] = updateGeneHash(geneHash, item, gene, genePred, gender)
	}
}

func updateDmd(item map[string]string) {
	item["SampleID"] = item["#Sample"]
	updateABC(item)
	item["#sample"] = item["#Sample"]
	item["OMIM"] = item["Disease"]
	if item["Significant"] != "YES" {
		item["Significant"] = "NO"
	}
	/*
		item["depth_rate"]=item["batch_control"]
		item["others_rate"]=item["all_control"]
		item["G/H"]=item["Mean_Ratio"]
		item["Zscore"]=item["Median_Ratio"]
		var omimWebsite="http://omim.org/search/?index=entry&start=1&limit=10&sort=score+desc%2C+prefix_sort+desc&search="+item["gene"]
		item["omimWebsite"]=omimWebsite
	*/
	// primerDesign
	var exID = item["exon"]
	var cdsID = item["exon"]
	var ratioVal, err = strconv.ParseFloat(item["Mean_Ratio"], 64)
	if err != nil {
		ratioVal = 0
	}
	if ratioVal >= 1.3 && ratioVal < 1.8 {
		item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exID + " DUP; - ;" + exID + "; " + cdsID + "; Het"
	} else if ratioVal >= 1.8 {
		if item["chr"] == "chrX" {
			item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exID + " DUP; - ;" + exID + "; " + cdsID + "; Hemi"
		} else {
			item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exID + " DUP; - ;" + exID + "; " + cdsID + "; Hom"
		}
	} else if ratioVal >= 0.2 && ratioVal <= 0.75 {
		item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exID + " DEL; - ;" + exID + "; " + cdsID + "; Het"
	} else if ratioVal < 0.2 {
		if item["chr"] == "chrX" {
			item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exID + " DEL; - ;" + exID + "; " + cdsID + "; Hemi"
		} else {
			item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exID + " DEL; - ;" + exID + "; " + cdsID + "; Hom"
		}
	} else {
		item["primerDesign"] = "-"
	}
}

func updateDipin(item map[string]string, db map[string]map[string]string) {
	var sampleID = item["sample"]
	var info, ok = db[sampleID]
	if !ok {
		info = item
	}
	var qc, aResult, bResult string
	if item["QC"] != "pass" {
		qc = "_等验证"
	}
	if item["chr11"] == "N" {
		bResult = "阴性"
	} else {
		bResult = item["chr11"]
	}
	if item["chr16"] == "N" {
		aResult = "阴性"
	} else {
		aResult = item["chr16"]
	}
	info["SampleID"] = item["sample"]
	info["地贫_QC"] = item["QC"]
	info["β地贫_chr11"] = item["chr11"]
	info["α地贫_chr16"] = item["chr16"]
	info["β地贫_最终结果"] = bResult + qc
	info["α地贫_最终结果"] = aResult + qc
	db[sampleID] = info
}

func updateSma(item map[string]string, db map[string]map[string]string) {
	var sampleID = item["SampleID"]
	var info, ok = db[sampleID]
	if !ok {
		info = item
	}
	var result, qc, qcResult string
	var Categorization = item["SMN1_ex7_cn"]
	var QC = item["qc"]
	if Categorization == "1.5" || Categorization == "1" || QC != "1" {
		qcResult = "_等验证"
	}
	switch Categorization {
	case "0":
		result = "纯阳性"
	case "0.5":
		result = "纯合灰区"
	case "1":
		result = "杂合阳性"
	case "1.5":
		result = "杂合灰区"
	default:
		result = "阴性"
	}
	if QC == "1" {
		qc = "Pass"
	} else {
		qc = "Fail"
	}
	info["SMN1_检测结果"] = result
	info["SMN1_质控结果"] = qc
	info["SMN1 EX7 del最终结果"] = result + qcResult
	db[sampleID] = info
}

func updateAe(item map[string]string) {
	updateABC(item)
	item["HyperLink"] = filepath.Join(*batch+".result_batCNV-dipin", "chr11_chr16_chrX_cnemap", item["SampleID"]+"_W30S25_cne.jpg")
	item["β地贫_最终结果_HyperLink"] = item["HyperLink"]
	item["α地贫_最终结果_HyperLink"] = item["HyperLink"]
	item["F8int1h-1.5k&2k最终结果"] = "检测范围外"
	item["F8int22h-10.8k&12k最终结果"] = "检测范围外"
}

func writeRow(excel *excelize.File, sheetName string, item map[string]string, title []string, rIdx int) {
	for j, k := range title {
		var axis = simpleUtil.HandleError(excelize.CoordinatesToCellName(j+1, rIdx)).(string)
		if formulaTitle[k] {
			simpleUtil.CheckErr(excel.SetCellFormula(sheetName, axis, item[k]))
		} else if hyperLinkTitle[k] {
			simpleUtil.CheckErr(excel.SetCellValue(sheetName, axis, item[k]))
			simpleUtil.CheckErr(excel.SetCellHyperLink(sheetName, axis, item[k+"_HyperLink"], "External"))
		} else {
			simpleUtil.CheckErr(excel.SetCellValue(sheetName, axis, item[k]))
		}
		var list, ok = dropListMap[k]
		if ok {
			var dvRange = excelize.NewDataValidation(true)
			dvRange.Sqref = axis
			simpleUtil.CheckErr(dvRange.SetDropList(list))
			simpleUtil.CheckErr(excel.AddDataValidation(sheetName, dvRange))
		}
	}
}

func writeTitle(excel *excelize.File, sheetName string, title []string) {
	for j, k := range title {
		var axis = simpleUtil.HandleError(excelize.CoordinatesToCellName(j+1, 1)).(string)
		simpleUtil.CheckErr(excel.SetCellValue(sheetName, axis, k))
	}
}

func loadBatchCNV(cnv string) {
	log.Println("Load BatchCNV Start")
	BatchCnv, BatchCnvTitle = textUtil.File2MapArray(cnv, "\t", nil)
	for _, item := range BatchCnv {
		var sampleID = item["sample"]
		var cn, err = strconv.Atoi(item["copyNumber"])
		simpleUtil.CheckErr(err, item["sample"]+" "+item["chr"]+":"+item["start"]+"-"+item["end"])
		updateSampleGeneInfo(float64(cn), sampleID, strings.Split(item["gene"], ",")...)
	}
	log.Println("Load BatchCNV Done")
}

func updateSampleGeneInfo(cn float64, sampleID string, genes ...string) {
	if cn != 2 {
		var geneInfo, ok = SampleGeneInfo[sampleID]
		if !ok {
			geneInfo = make(map[string]*GeneInfo)
			for _, gene := range genes {
				geneInfo[gene] = &GeneInfo{
					基因:   gene,
					遗传模式: geneInheritance[gene],
					cnv:  true,
					cnv0: cn == 0,
				}
			}
			SampleGeneInfo[sampleID] = geneInfo
		} else {
			for _, gene := range genes {
				var info, ok = geneInfo[gene]
				if !ok {
					geneInfo[gene] = &GeneInfo{
						基因:   gene,
						遗传模式: geneInheritance[gene],
						cnv:  true,
						cnv0: cn == 0,
					}
				} else {
					info.cnv = true
					info.cnv0 = info.cnv0 || cn == 0
				}
			}
		}
	}
}

func updateCnvTags(item map[string]string, sampleID string, genes ...string) {
	var tag3, tag4 bool
	for _, gene := range genes {
		var info, ok = SampleGeneInfo[sampleID][gene]
		if ok {
			info.标签4()
			if info.tag3 {
				tag3 = true
			}
			if info.tag4 {
				tag4 = true
			}
		}
	}
	if tag3 {
		item["Database"] += "3"
	}
	if tag4 {
		item["Database"] += "4"
	}
}

func loadQC(qc string) (qcDb []map[string]string) {
	var excel = simpleUtil.HandleError(excelize.OpenFile(qc)).(*excelize.File)
	var rows, err = excel.GetRows("Sheet1")
	simpleUtil.CheckErr(err)
	var title = rows[0]
	for i := range rows {
		if i > 0 {
			var item = make(map[string]string)
			for j := range title {
				item[title[j]] = rows[i][j]
			}
			qcDb = append(qcDb, item)
		}
	}
	return
}
