package main

import (
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
	AES "github.com/liserjrqlxue/crypto/aes"
	"github.com/liserjrqlxue/goUtil/jsonUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/xuri/excelize/v2"
)

func mergeSep(str, sep string) string {
	var strMap = make(map[string]string)
	for _, s := range strings.Split(str, sep) {
		strMap[s] = s
	}
	var strs []string
	for s := range strMap {
		strs = append(strs, s)
	}
	return strings.Join(strs, sep)

}

func loadDb() {
	log.Println("Load Database Start")
	// load gene list
	for _, key := range textUtil.File2Array(*geneList) {
		geneListMap[key] = true
	}
	// load gene info list
	geneInfoMap, _ = textUtil.File2MapMap(*geneInfoList, "Gene Symbol", "\t", nil)
	// load gene sub list
	var geneSubs, _ = textUtil.File2MapArray(filepath.Join(etcPath, "gene.sub.list.txt"), "\t", nil)
	for _, item := range geneSubs {
		geneSubListMap[item["基因"]] = true
	}
	// load gene exclude list
	for _, key := range textUtil.File2Array(filepath.Join(etcPath, "gene.exclude.list.txt")) {
		geneExcludeListMap[key] = true
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
		diseaseSep,
	)
	for gene, info := range diseaseDb {
		info["遗传模式merge"] = mergeSep(info["遗传模式"], diseaseSep)
		geneInheritance[gene] = info["遗传模式"]
	}

	// load drop list
	log.Println("Load DropList Start")
	for k, v := range simpleUtil.HandleError(textUtil.File2Map(*dropList, "\t", false)).(map[string]string) {
		dropListMap[k] = strings.Split(v, ",")
	}
	log.Println("Load Database Done")

	// load sample detail
	if *detail != "" {
		var details = textUtil.File2Slice(*detail, "\t")
		for _, line := range details {
			var info = make(map[string]string)
			var sampleID = line[0]
			info["productCode"] = line[1]
			info["hospital"] = line[2]
			sampleDetail[sampleID] = info
		}
	}
}

func loadLocalDb(throttle chan bool) {
	// load 已解读数据库
	log.Println("Load LocalDb Start")
	localDb = jsonUtil.Json2MapMap(AES.DecodeFile(*mutDb, []byte(codeKey)))
	log.Println("Load LocalDb Done")
	<-throttle
}

var formulaTitle = map[string]bool{
	"解读人": true,
	"审核人": true,
}

var hyperLinkTitle = map[string]bool{
	"β地贫_最终结果":      true,
	"α地贫_最终结果":      true,
	"reads_picture": true,
	"P0":            true,
	"P1":            true,
	"P2":            true,
	"P3":            true,
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
	if err == nil && af >= tv {
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
		item["遗传模式merge"] = disease["遗传模式merge"]
		item["ModeInheritance"] = item["遗传模式"]
		item["疾病简介"] = disease["疾病简介"]
		item["包装疾病分类"] = disease["包装疾病分类"]
	}
}

func addDiseases2Cnv(item map[string]string, sep string, genes ...string) {
	var diseaseCN, inherit, diseaseInfo []string
	for _, gene := range genes {
		var info = diseaseDb[gene]
		diseaseCN = append(diseaseCN, info["疾病"])
		inherit = append(inherit, info["遗传模式"])
		diseaseInfo = append(diseaseInfo, info["疾病简介"])
	}
	item["疾病中文名"] = strings.Join(diseaseCN, sep)
	item["遗传模式"] = strings.Join(inherit, sep)
	item["中文-疾病背景"] = strings.Join(diseaseInfo, sep)
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

func updateAvd(item map[string]string, subFlag bool) {
	updateABC(item)
	item["HGMDorClinvar"] = "否"
	if isHGMD[item["HGMD Pred"]] || isClinVar[item["ClinVar Significance"]] {
		item["HGMDorClinvar"] = "是"
	}
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
			item["isReport"] = "Y"
			if subFlag {
				if geneSubListMap[item["Gene Symbol"]] {
					item["报告类别-原始"] = "正式报告"
				} else {
					item["报告类别-原始"] = "补充报告"
				}

			} else {
				item["报告类别-原始"] = "正式报告"
			}
		} else {
			item["Database"] = "NBS-out"
			if item["LOF"] == "YES" && !geneExcludeListMap[item["Gene Symbol"]] {
				item["isReport"] = "Y"
				item["报告类别-原始"] = "补充报告"
			}
		}
		item["参考文献"] = db["Reference"]
		item["位点关联疾病"] = db["Disease"]
		item["位点关联遗传模式"] = db["遗传模式"]
		//item["Evidence New + Check"] = db["证据项"]
		item["Definition"] = db["Definition"]
	} else {
		item["Database"] = "."
		if item["LOF"] == "YES" && !geneExcludeListMap[item["Gene Symbol"]] {
			item["报告类别-原始"] = "补充报告"
			item["isReport"] = "Y"
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
	item["引物设计"] = anno.PrimerDesign(item)
	item["验证"] = ifCheck(item)
	readsPicture(item)
}

func ifCheck(item map[string]string) string {
	var depth int
	var ratio float64
	depth, err = strconv.Atoi(item["Depth"])
	if err != nil {
		return "Y"
	}
	ratio, err = strconv.ParseFloat(item["A.Ratio"], 64)
	if err != nil {
		return "Y"
	}
	if depth < 40 || ratio < 0.4 {
		return "Y"
	}
	if len(item["Ref"]) != 1 || len(item["Call"]) != 1 || item["Ref"] == "." || item["Call"] == "." {
		if depth < 60 || ratio < 0.45 {
			return "Y"
		}
	}
	return ""
}

func readsPicture(item map[string]string) {
	if ifPlotReads(item) {
		var sampleID = item["SampleID"]
		var chr = item["#Chr"]
		if chr == "MT" {
			chr = "chrM_NC_012920.1"
		} else {
			chr = "chr" + chr
		}
		var stop = item["Stop"]
		var png = strings.Join([]string{sampleID, chr, stop}, "_") + ".png"
		item["reads_picture"] = png
		item["reads_picture_HyperLink"] = filepath.Join("reads_picture", png)
	}
}

func ifPlotReads(item map[string]string) bool {
	if item["isReport"] == "Y" {
		return true
	}
	if isClinVar[item["ClinVar Significance"]] || isHGMD[item["HGMD Pred"]] {
		return true
	}
	if item["Database"] != "" {
		return true
	}
	return false
}

func updateFromAvd(item, geneHash map[string]string, geneInfo map[string]*GeneInfo, sampleID string) {
	if item["filterAvd"] != "Y" {
		return
	}
	var info, ok = geneInfo[item["Gene Symbol"]]
	if !ok {
		info = new(GeneInfo).new(item)
		geneInfo[item["Gene Symbol"]] = info
	}
	info.count(item)
	if *gender == "M" || genderMap[sampleID] == "M" {
		item["Sex"] = "M"
		info.性别 = "M"
		UpdateGeneHash(geneHash, item, "M")
	} else if *gender == "F" || genderMap[sampleID] == "F" {
		item["Sex"] = "F"
		info.性别 = "F"
		UpdateGeneHash(geneHash, item, "F")
	}
	geneInfo[item["Gene Symbol"]] = info
}

func updateGeneHashAD(item map[string]string) string {
	switch item["Zygosity"] {
	case "Hom", "Het":
		return "可能患病"
	default:
		return ""
	}
}

func updateGeneHashXLD(item map[string]string) string {
	switch item["Zygosity"] {
	case "Hom", "Het", "Hemi":
		return "可能患病"
	default:
		return ""
	}
}

func updateGeneHashXL(item map[string]string) string {
	if item["Gene Symbol"] == "OTC" || item["Gene Symbol"] == "GLA" || item["Gene Symbol"] == "PCDH19" {
		return updateGeneHashXLD(item)
	}
	return ""
}

func updateGeneHashAR(item map[string]string, genePred string) string {
	switch item["Zygosity"] {
	case "Hom":
		return "可能患病"
	case "Het":
		if genePred == "" {
			return "携带者"
		}
		return "可能患病"
	default:
		return ""
	}
}

func updateGeneHashXLR(item map[string]string, genePred, gender string) string {
	switch gender {
	case "M":
		return updateGeneHashXLD(item)
	case "F":
		return updateGeneHashAR(item, genePred)
	default:
		return ""
	}
}

func updateGeneHash(item map[string]string, genePred, gender string) string {
	if isAD.MatchString(item["遗传模式merge"]) {
		return updateGeneHashAD(item)
	}
	if isXLD.MatchString(item["遗传模式merge"]) {
		return updateGeneHashXLD(item)
	}
	switch item["遗传模式merge"] {
	case "AR":
		return updateGeneHashAR(item, genePred)
	case "MI":
		return updateGeneHashAD(item)
	case "XLR":
		return updateGeneHashXLR(item, genePred, gender)
	case "XL":
		return updateGeneHashXL(item)
	default:
		return ""
	}
}

// UpdateGeneHash : update geneHash
func UpdateGeneHash(geneHash, item map[string]string, gender string) {
	if item["isReport"] != "Y" {
		return
	}
	var gene = item["Gene Symbol"]
	var genePred, ok = geneHash[gene]
	if !ok || genePred != "可能患病" {
		geneHash[gene] = updateGeneHash(item, genePred, gender)
	}
}

func updateDmd(item map[string]string) {
	var sampleID = item["#Sample"]
	item["SampleID"] = sampleID
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
	var pngSuffix = "." + item["gene"] + "." + item["NM"] + ".png"
	item["P0_HyperLink"] = filepath.Join("DMD_exon_graph", item["SampleID"]+"."+pngSuffix)
	var info, ok = sampleInfos[item["SampleID"]]
	if ok {
		item["P0"] = info.p0
		updateP(item, "P1", info.p1, pngSuffix)
		updateP(item, "P2", info.p2, pngSuffix)
		updateP(item, "P3", info.p3, pngSuffix)
	} else {
		log.Printf("can not find info of [%s] from %s", sampleID, *qc)
	}
}

func updateP(item map[string]string, k, v, suffix string) {
	var ps = strings.Split(v, ",")
	var sampleID = ps[0]
	item[k] = strings.Join(ps[1:], ",")
	item[k+"_HyperLink"] = filepath.Join("DMD_exon_graph", sampleID+suffix)
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
	var result, qcResult string
	var Categorization = item["SMN1_ex7_cn"]
	var QC = item["qc"]
	if Categorization == "1.5" || Categorization == "1" || QC != "1" {
		qcResult = "_等验证"
	}
	switch Categorization {
	case "0", "0.0":
		result = "纯合阳性"
	case "0.5":
		result = "纯合灰区"
	case "1", "1.0":
		result = "杂合阳性"
	case "1.5":
		result = "杂合灰区"
	default:
		result = "阴性"
	}
	if QC == "1" {
		info["SMN1_质控结果"] = "Pass"
		if Categorization == "0" || Categorization == "1" {
			info["SMN1 EX7 del最终结果"] = result
		} else {
			info["SMN1 EX7 del最终结果"] = result + qcResult
		}
	} else {
		info["SMN1_质控结果"] = "Fail"
		info["SMN1 EX7 del最终结果"] = result + qcResult
	}
	info["SMN1_检测结果"] = result
	updateABC(item)
	db[sampleID] = info
}
func updateSma2(item map[string]string, db map[string]map[string]string) {
	var sampleID = item["Sample"]
	var info, ok = db[sampleID]
	if !ok {
		info = item
	}
	info["SMN1_CN"] = item["SMN1_CN"]
	info["SMN1_CN_raw"] = item["SMN1_CN_raw"]
	db[sampleID] = info
}

func updateAe(item map[string]string) {
	updateABC(item)
	if *wgs {
		item["HyperLink"] = filepath.Join(*batch+".result_batCNV-dipin", "chr11_chr16_chrX_cnemap", item["SampleID"]+"_W60S50_cne.jpg")
	} else {
		item["HyperLink"] = filepath.Join(*batch+".result_batCNV-dipin", "chr11_chr16_chrX_cnemap", item["SampleID"]+"_W30S25_cne.jpg")
	}
	item["β地贫_最终结果_HyperLink"] = item["HyperLink"]
	item["α地贫_最终结果_HyperLink"] = item["HyperLink"]
	item["F8int1h-1.5k&2k最终结果"] = "检测范围外"
	item["F8int22h-10.8k&12k最终结果"] = "检测范围外"
}

func writeRow(excel *excelize.File, sheetName string, item map[string]string, title []string, rIdx int) {
	var axis0 = simpleUtil.HandleError(excelize.CoordinatesToCellName(1, rIdx)).(string)
	var axis1 = simpleUtil.HandleError(excelize.CoordinatesToCellName(len(title), rIdx)).(string)
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
	var formalID, supplementaryID int
	if item["验证"] == "Y" {
		formalID = formalCheckStyleID
		supplementaryID = supplementaryCheckStyleID
	} else {
		formalID = formalStyleID
		supplementaryID = supplementaryStyleID
	}
	switch item["报告类别-原始"] {
	case "正式报告":
		simpleUtil.CheckErr(excel.SetCellStyle(sheetName, axis0, axis1, formalID), sheetName, axis0, axis1)
	case "补充报告":
		simpleUtil.CheckErr(excel.SetCellStyle(sheetName, axis0, axis1, supplementaryID), sheetName, axis0, axis1)
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
	var tagMap = make(map[string]bool)
	for _, gene := range genes {
		var info, ok = SampleGeneInfo[sampleID][gene]
		if ok {
			info.标签4()
			if info.tag3 != "" {
				tagMap[info.tag3] = true
			}
			if info.tag4 {
				tagMap["4"] = true
			}
		}
	}
	var tags []string
	for k := range tagMap {
		tags = append(tags, k)
	}
	sort.Strings(tags)
	item["Database"] = strings.Join(tags, ";")
}

// QC
func writeQC(excel *excelize.File, db []map[string]string) {
	var rows = simpleUtil.HandleError(excel.GetRows(*qcSheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)
	qcMap, err = textUtil.File2Map(*qcTitle, "\t", true)
	simpleUtil.CheckErr(err, "load qcTitle fail")
	for i, item := range db {
		rIdx++
		updateQC(item, qcMap, i)
		updateINDEX(item, "B", rIdx)
		writeRow(excel, *qcSheetName, item, title, rIdx)
	}
}

func updateQC(item, qcMap map[string]string, i int) {
	item["Order"] = strconv.Itoa(i + 1)
	for k, v := range qcMap {
		item[k] = item[v]
	}
	var inputGender = "null"
	if limsInfo[item["Sample"]]["SEX"] == "1" {
		inputGender = "M"
	} else if limsInfo[item["Sample"]]["SEX"] == "2" {
		inputGender = "F"
	} else {
		inputGender = "null"
	}
	if inputGender != genderMap[limsInfo[item["Sample"]]["MAIN_SAMPLE_NUM"]] {
		item["Gender"] = inputGender + "!!!Sequenced" + genderMap[limsInfo[item["Sample"]]["MAIN_SAMPLE_NUM"]]
	}
	//item["RESULT"]=item[""]
	item["产品编号"] = limsInfo[item["Sample"]]["PRODUCT_CODE"]
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
			var info = newSampleInfo(item)
			sampleInfos[info.sampleID] = info
			qcDb = append(qcDb, item)
		}
	}
	return
}

type handleItem func(map[string]string)

func updateData2Sheet(excel *excelize.File, sheetName string, db []map[string]string, fn handleItem) {
	log.Printf("update [%s]", sheetName)
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)

	for _, item := range db {
		rIdx++
		fn(item)
		writeRow(excel, sheetName, item, title, rIdx)
	}
}

func updateDataFile2Sheet(excel *excelize.File, sheetName, path string, fn handleItem) {
	log.Printf("update [%s]", sheetName)
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)

	var db, _ = textUtil.File2MapArray(path, "\t", nil)
	for _, item := range db {
		rIdx++
		fn(item)
		writeRow(excel, sheetName, item, title, rIdx)
	}
}

func updateDataList2Sheet(excel *excelize.File, sheetName, list string, fn handleItem) {
	log.Printf("update [%s]", sheetName)
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)

	for _, path := range textUtil.File2Array(list) {
		var db, _ = textUtil.File2MapArray(path, "\t", nil)
		for _, item := range db {
			rIdx++
			fn(item)
			writeRow(excel, sheetName, item, title, rIdx)
		}
	}
}

func updateCNV(item map[string]string) {
	updateCnvTags(item, item["#Sample"], item["gene"])
}

func updateDMD(item map[string]string) {
	item["#sample"] = item["Sample"]
	item["sampleID"] = item["Sample"]
	updateABC(item)
}

func updateFeature(item map[string]string) {
	item["参考文献"] = strings.ReplaceAll(item["参考文献"], "<br/>", "\n")
	updateABC(item)
}

func updateBatchCNV(item map[string]string) {
	var genes = strings.Split(item["gene"], ",")
	updateCnvTags(item, item["sample"], genes...)
	addDiseases2Cnv(item, multiDiseaseSep, genes...)
	item["疾病名称"] = item["疾病中文名"]
	item["疾病简介"] = item["中文-疾病背景"]
	item["SampleID"] = item["sample"]
	var (
		targetGenes       []string
		targetTranscripts []string
	)
	for _, gene := range genes {
		if geneListMap[gene] {
			targetGenes = append(targetGenes, gene)
			targetTranscripts = append(targetTranscripts, geneInfoMap[gene]["Transcript"])
		}
	}
	item["新生儿目标基因"] = strings.Join(targetGenes, ",")
	item["转录本"] = strings.Join(targetTranscripts, ",")
	updateABC(item)
	item["CNVType"] = getCNVtype(item["Sex"], item)
	item["引物设计"] = strings.Join(
		[]string{
			item["gene"],
			item["转录本"],
			item["exons"] + " " + item["CNVType"],
			"-",
			item["exons"],
			item["exons"],
			item["杂合性"],
		},
		"; ",
	)
}

func getCNVtype(gender string, item map[string]string) string {
	switch item["copyNumber"] {
	case "", "-":
		return ""
	case "0":
		return "DEL"
	case "1":
		if item["chr"] == "chrX" || item["chr"] == "chrY" {
			if item["chr"] == "chrX" && gender == "F" {
				return "DEL"
			}
		} else {
			return "DEL"
		}
	case "2":
		if (item["chr"] == "chrX" || item["chr"] == "chrY") && gender == "M" {
			return "DUP"
		}
	default:
		return "DUP"
	}
	return ""
}

func updateBamPath(excel *excelize.File, list string) {
	for i, path := range textUtil.File2Array(list) {
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
