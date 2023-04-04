package main

import (
	"github.com/liserjrqlxue/goUtil/stringsUtil"
	"log"
	"path/filepath"
	"regexp"
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

func buildDiseaseDb(diseaseMapArray []map[string]string, diseaseTitle []string, key string) {
	for _, item := range diseaseMapArray {
		if item["报告逻辑"] != "" {
			item["报告逻辑"] = item["报告逻辑"] + "（" + item["疾病"] + "）"
		}
		var mainKey = item[key]
		var mainItem, ok = diseaseDb[mainKey]
		if ok {
			for _, k := range diseaseTitle {
				if k == "报告逻辑" {
					if item[k] != "" {
						if mainItem[k] == "" {
							mainItem[k] = item[k]
						} else {
							mainItem[k] += diseaseSep + item[k]
						}
					}
				} else {
					mainItem[k] += diseaseSep + item[k]
				}
			}
		} else {
			mainItem = item
		}
		diseaseDb[mainKey] = mainItem
	}
}

func loadDiseaseDb(i18n string) {
	// load disease database
	log.Println("Load Disease Start")
	if i18n == "EN" {
		var diseaseMapArray, diseaseTitle = textUtil.File2MapArray(
			filepath.Join(etcPath, "新生儿疾病库.EN.xlsx.新生儿疾病库V2-英文版.txt"),
			"\t", nil,
		)
		buildDiseaseDb(diseaseMapArray, diseaseTitle, "Gene")
		for gene, m := range diseaseDb {
			m["疾病"] = m["Condition Name"]
			m["遗传模式"] = m["Inherited Mode"]
			m["遗传模式merge"] = mergeSep(m["遗传模式"], diseaseSep)
			m["疾病简介"] = m["Disease Generalization"]
			m["包装疾病分类"] = m["Condition Category"]
			geneInheritance[gene] = m["遗传模式"]
		}
	} else {
		var diseaseMapArray, diseaseTitle = simpleUtil.Slice2MapArray(
			simpleUtil.HandleError(
				simpleUtil.HandleError(
					excelize.OpenFile(*diseaseExcel),
				).(*excelize.File).
					GetRows(*diseaseSheetName),
			).([][]string),
		)

		buildDiseaseDb(diseaseMapArray, diseaseTitle, "基因")

		for gene, info := range diseaseDb {
			info["遗传模式merge"] = mergeSep(info["遗传模式"], diseaseSep)
			geneInheritance[gene] = info["遗传模式"]
		}
	}
	log.Println("Load Database Done")
}

func loadDb() {
	log.Println("Load Database Start")
	// load gene info list
	geneInfoMap, _ = textUtil.File2MapMap(*geneInfoList, "Gene Symbol", "\t", nil)
	if *im {
		for s, m := range geneInfoMap {
			if m["一体机过滤基因"] == "TRUE" {
				geneIMListMap[s] = true
				if m["备注"] != "定点" {
					geneListMap[s] = true
				}
			}
		}
	} else {
		// load gene list
		for _, key := range textUtil.File2Array(*geneList) {
			geneListMap[key] = true
		}
	}

	// load gene sub list
	var geneSubs, _ = textUtil.File2MapArray(filepath.Join(etcPath, "gene.sub.list.txt"), "\t", nil)
	for _, item := range geneSubs {
		geneSubListMap[item["基因"]] = true
	}
	// load deafness gene list
	var deafnessGenes, _ = textUtil.File2MapArray(filepath.Join(etcPath, "耳聋24基因.xlsx.Sheet1.txt"), "\t", nil)
	for _, gene := range deafnessGenes {
		deafnessGeneList[gene["基因"]] = true
	}

	// load gene exclude list
	for _, key := range textUtil.File2Array(filepath.Join(etcPath, "gene.exclude.list.txt")) {
		geneExcludeListMap[key] = true
	}
	// load function exclude list
	for _, key := range textUtil.File2Array(*functionExclude) {
		functionExcludeMap[key] = true
	}

	// load CNV database
	log.Println("Load CNV database Start")
	var cnvDbArray, _ = simpleUtil.Slice2MapArray(
		simpleUtil.HandleError(
			simpleUtil.HandleError(
				excelize.OpenFile(filepath.Join(etcPath, "CNV配置文件.xlsx")),
			).(*excelize.File).
				GetRows("CNV库配置文件"),
		).([][]string),
	)
	for _, m := range cnvDbArray {
		var key = m["Gene Symbol"] + " " + m["Function"]
		cnvDb[key] = m
	}

	// load drop list
	log.Println("Load DropList Start")
	for k, v := range simpleUtil.HandleError(textUtil.File2Map(*dropList, "\t", false)).(map[string]string) {
		dropListMap[k] = strings.Split(v, ",")
	}

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

	if *cs {
		var region *Region
		var repeatRegionArray, _ = textUtil.File2MapArray(filepath.Join(etcPath, "repeat.txt"), "\t", nil)
		for _, m := range repeatRegionArray {
			region = &Region{
				chr:   "",
				start: stringsUtil.Atoi(m["Start"]),
				end:   stringsUtil.Atoi(m["Stop"]),
				gene:  "",
			}
			repeatRegion = append(repeatRegion, region)
		}
		var homologousRegionArray, _ = textUtil.File2MapArray(filepath.Join(etcPath, "homologous.regions.txt"), "\t", nil)
		for _, m := range homologousRegionArray {
			region = newRegion(m["目标区域（疑似有同源区域）"])
			if region != nil {
				homologousRegion = append(homologousRegion, region)
			}
			region = newRegion(m["相似区域"])
			if region != nil {
				homologousRegion = append(homologousRegion, region)
			}
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

var (
	cHGVSalt = regexp.MustCompile(`alt: (\S+) \)`)
	cHGVSstd = regexp.MustCompile(`std: (\S+) `)
)

func cHgvsAlt(cHgvs string) string {
	if m := cHGVSalt.FindStringSubmatch(cHgvs); m != nil {
		return m[1]
	}
	return cHgvs
}

func cHgvsStd(cHgvs string) string {
	if m := cHGVSstd.FindStringSubmatch(cHgvs); m != nil {
		return m[1]
	}
	return cHgvs
}

func filterAvd(item map[string]string) bool {
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
	if item["自动化判断"] == "B" || item["自动化判断"] == "LB" {
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
		item["报告逻辑"] = disease["报告逻辑"]
	}
	if *cs {
		item["遗传模式"] = item["Inheritance"]
		item["疾病中文名"] = item["Chinese disease name"]
		item["中文-疾病背景"] = item["Chinese disease introduction"]
		item["中文-突变详情"] = item["Chinese mutation information"]
		item["Disease*"] = item["English disease name"]
		item["英文-疾病背景"] = item["English disease introduction"]
		item["英文-突变详情"] = item["English mutation information"]
		item["参考文献"] = item["Reference-final-Info"]
	}
}

func addDatabase2Cnv(item map[string]string) {
	var gene = item["gene"]
	var mut = item["核苷酸变化"]
	var key = gene + " " + mut
	var db, ok = cnvDb[key]
	if ok {
		item["新生儿一体机包装变异"] = db["新生儿一体机包装变异"]
		item["中文-突变判定"] = db["中文-突变判定"]
	}
	if item["新生儿一体机包装变异"] == "" {
		item["新生儿一体机包装变异"] = "否"
	}
	item["报告类别"] = item["新生儿一体机包装变异"]
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

func annoLocaDb(item map[string]string, varDb map[string]map[string]string, subFlag bool) {
	var (
		transcript = item["Transcript"]
		c          = item["cHGVS"]
		cAlt       = cHgvsAlt(c)
		cStd       = cHgvsStd(c)
		mainKey    = transcript + "\t" + c
		mainKey1   = transcript + "\t" + cAlt
		mainKey2   = transcript + "\t" + cStd
	)

	var db, ok = varDb[mainKey]
	if !ok {
		db, ok = varDb[mainKey1]
	}
	if !ok {
		db, ok = varDb[mainKey2]
	}
	if ok {
		if db["是否是包装位点"] == "是" {
			if *im {
				item["报告类别"] = "是"
				item["In BGI database"] = "是"
			}
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
			} else if subFlag && mainKey == "NM_000277.1\tc.158G>A" {
				item["isReport"] = "Y"
				item["报告类别-原始"] = "补充报告"
			}
		}
		item["参考文献"] = db["Reference"]
		item["位点关联疾病"] = db["Disease"]
		item["位点关联遗传模式"] = db["遗传模式"]
		//item["Evidence New + Check"] = db["证据项"]
		item["Definition"] = db["Definition"]
		item["filterAvd"] = "Y"
	} else {
		item["Database"] = "."
		if item["LOF"] == "YES" && !geneExcludeListMap[item["Gene Symbol"]] {
			item["报告类别-原始"] = "补充报告"
			item["isReport"] = "Y"
		}
		if filterAvd(item) {
			item["filterAvd"] = "Y"
		}
	}
}

func updateAvd(item map[string]string, subFlag bool) {
	updateABC(item, item["SampleID"])
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

	if *im {
		item["报告类别"] = "否"
		item["In BGI database"] = "否"
	}

	// reads_picture
	// reads_picture_HyperLink
	readsPicture(item)
	// Function
	anno.UpdateFunction(item)
	// AF -1 -> .
	updateAf(item)

	if *acmg {
		acmg2015.AddEvidences(item)
		item["自动化判断"] = acmg2015.PredACMG2015(item, *autoPVS1)
		anno.UpdateAutoRule(item)
	}
	if *cs {
		floatFormat(item, afFloatFormatArray, 6)
		// remove trailing zeros
		floatFormat(item, afFloatFormatArray, -1)
		var (
			repeatHit     bool
			homologousHit bool
			start         = stringsUtil.Atoi(item["Start"])
			end           = stringsUtil.Atoi(item["Stop"])
			hits          []string
		)
		for _, region := range repeatRegion {
			if region.gene == item["Gene Symbol"] && start > region.start && end < region.end {
				repeatHit = true
			}
		}
		for _, region := range homologousRegion {
			if region.chr == item["#Chr"] && start > region.start && end < region.end {
				homologousHit = true
			}
		}
		if repeatHit {
			hits = append(hits, "重复区域变异")
		}
		if homologousHit {
			hits = append(hits, "同源区域变异")
		}
		item["需验证的变异"] = strings.Join(hits, ";")
		item["#Chr"] = addChr(item["#Chr"])
	} else {
		annoLocaDb(item, localDb, subFlag)
	}
	item["引物设计"] = anno.PrimerDesign(item)
	item["验证"] = ifCheck(item)
}

func ifCheck(item map[string]string) string {
	var (
		depth int
		ratio float64
		err   error
	)
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
		info.gender = "M"
		UpdateGeneHash(geneHash, item, "M")
	} else if *gender == "F" || genderMap[sampleID] == "F" {
		item["Sex"] = "F"
		info.gender = "F"
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
	item["sampleID"] = sampleID
	item["SampleID"] = sampleID
	updateABC(item, sampleID)
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
	updateDMDHyperlLink(item)
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
	var tag, aResult, bResult string
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

	if *im {
		if item["QC"] == "pass" {
			item["QC"] = "通过"
		} else {
			item["QC"] = "不通过"
			aResult = "灰区"
			bResult = "灰区"
		}

		if aResult == "." {
			aResult = "灰区"
		}
		if bResult == "." {
			bResult = "灰区"
		}
	} else {
		if item["QC"] != "pass" {
			tag = "_等验证"
		}
	}
	info["β地贫自动化结果"] = bResult + tag
	info["α地贫自动化结果"] = aResult + tag
	info["β地贫_最终结果"] = bResult + tag
	info["α地贫_最终结果"] = aResult + tag
	info["SampleID"] = item["sample"]
	info["地贫_QC"] = item["QC"]
	info["β地贫_chr11"] = item["chr11"]
	info["α地贫_chr16"] = item["chr16"]
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
	if Categorization == "1.5" || Categorization == "1" || Categorization == "1.0" || QC != "1" {
		qcResult = "_等验证"
	}
	if *im {
		info["Official Report"] = "否"
		switch Categorization {
		case "0", "0.0":
			result = "纯合阳性"
			info["Official Report"] = "是"
		case "1", "1.0":
			result = "杂合阳性"
		case "0.5", "1.5":
			result = "灰区"
		default:
			result = "阴性"
		}
		if QC == "1" {
			info["SMN1_质控结果"] = "通过"
			info["SMN1 EX7 del最终结果"] = result
		} else {
			info["SMN1_质控结果"] = "不通过"
			info["SMN1 EX7 del最终结果"] = "灰区"
		}
	} else {
		switch Categorization {
		case "0", "0.0":
			result = "纯合阳性"
			if *im {
				info["Official Report"] = "是"
			}
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
			switch Categorization {
			case "0", "0.0", "1", "1.0":
				info["SMN1 EX7 del最终结果"] = result
			default:
				info["SMN1 EX7 del最终结果"] = result + qcResult
			}
		} else {
			info["SMN1_质控结果"] = "Fail"
			info["SMN1 EX7 del最终结果"] = result + qcResult
		}
	}
	info["SMN1_检测结果"] = result
	info["SMN2_ex7_cn"] = item["SMN2_ex7_cn"]
	updateABC(item, sampleID)
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
	if *cs {
		info["SMN1_质控结果"] = "fail"
		info["SMN1_检测结果"] = ""
		info["SMN1 EX7 del最终结果"] = "_等验证"
		if info["SMN1_CN"] == "None" {
			info["SMN1_检测结果"] = "."
		} else {
			var cn, err = strconv.ParseFloat(info["SMN1_CN"], 64)
			if err == nil {
				if cn == 0 {
					info["SMN1_质控结果"] = "pass"
					info["SMN1_检测结果"] = "纯合阳性"
					info["SMN1 EX7 del最终结果"] = "纯合阳性_等验证"
				} else if cn == 0.5 {
					info["SMN1_质控结果"] = "pass"
					info["SMN1_检测结果"] = "纯合灰区"
					info["SMN1 EX7 del最终结果"] = "纯合灰区_等验证"

				} else if cn == 1 {
					info["SMN1_质控结果"] = "pass"
					info["SMN1_检测结果"] = "杂合阳性"
					info["SMN1 EX7 del最终结果"] = "杂合阳性_等验证"

				} else if cn == 1.5 {
					info["SMN1_质控结果"] = "pass"
					info["SMN1_检测结果"] = "杂合灰区"
					info["SMN1 EX7 del最终结果"] = "杂合灰区_等验证"

				} else if cn >= 2 {
					info["SMN1_质控结果"] = "pass"
					info["SMN1_检测结果"] = "阴性"
					info["SMN1 EX7 del最终结果"] = "阴性"
				}
			}
		}
	}
	db[sampleID] = info
}

func updateAe(item map[string]string) {
	if *wgs {
		item["HyperLink"] = filepath.Join(*batch+".result_batCNV-dipin", "chr11_chr16_chrX_cnemap", item["SampleID"]+"_W60S50_cne.jpg")
	} else if *cs {
		item["HyperLink"] = filepath.Join("batCNV", item["SampleID"]+"_W60S50_cne.jpg")
	} else {
		item["HyperLink"] = filepath.Join(*batch+".result_batCNV-dipin", "chr11_chr16_chrX_cnemap", item["SampleID"]+"_W30S25_cne.jpg")
	}
	item["β地贫_chr11_HyperLink"] = item["HyperLink"]
	item["α地贫_chr16_HyperLink"] = item["HyperLink"]
	item["β地贫_最终结果_HyperLink"] = item["HyperLink"]
	item["α地贫_最终结果_HyperLink"] = item["HyperLink"]
	item["F8int1h-1.5k&2k最终结果"] = "检测范围外"
	item["F8int22h-10.8k&12k最终结果"] = "检测范围外"
}

func writeRow(excel *excelize.File, sheetName string, item map[string]string, title []string, rIdx int) {
	var axis0 = simpleUtil.HandleError(excelize.CoordinatesToCellName(1, rIdx)).(string)
	var axis1 = simpleUtil.HandleError(excelize.CoordinatesToCellName(len(title), rIdx)).(string)
	for j, k := range title {
		if *im {
			item[k] = getI18n(item[k], k)
		}
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
	BatchCnv, _ = textUtil.File2MapArray(cnv, "\t", nil)
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
					gene:        gene,
					inheritance: geneInheritance[gene],
					cnv:         true,
					cnv0:        cn == 0,
				}
			}
			SampleGeneInfo[sampleID] = geneInfo
		} else {
			for _, gene := range genes {
				var info, ok = geneInfo[gene]
				if !ok {
					geneInfo[gene] = &GeneInfo{
						gene:        gene,
						inheritance: geneInheritance[gene],
						cnv:         true,
						cnv0:        cn == 0,
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
			info.Tag4()
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

func floatFormat(item map[string]string, keys []string, prec int) {
	for _, key := range keys {
		value := item[key]
		if value == "" || value == "." {
			item[key] = ""
			return
		}
		floatValue, e := strconv.ParseFloat(value, 64)
		if e != nil {
			log.Printf("can not ParseFloat:%s[%s]\n", key, value)
		} else {
			item[key] = strconv.FormatFloat(floatValue, 'f', prec, 64)
		}
	}
}

// QC
func writeQC(excel *excelize.File, db []map[string]string) {
	var rows = simpleUtil.HandleError(excel.GetRows(*qcSheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)
	for i, item := range db {
		rIdx++
		if *cs {
			item["Q20(%)"] = item["Q20"]
			item["Q30(%)"] = item["Q30"]
			item["sampleID"] = item["Sample"]
			floatFormat(item, qcFloatFormatArray, 2)
			updateInfo(item, item["Sample"])
		} else {
			updateQC(item, i)
		}
		updateINDEX(item, "B", rIdx)
		writeRow(excel, *qcSheetName, item, title, rIdx)
	}
}

func updateQC(item map[string]string, i int) {
	var sampleID = item["sampleID"]
	item["Order"] = strconv.Itoa(i + 1)
	item["Gender"] = item["gender_analysed"]
	if *im {
		updateInfo(item, sampleID)
		if item["gender_analysed"] != item["gender"] {
			item["Gender"] = item["gender"] + "!!!Sequenced" + item["gender_analysed"]
		}
	} else {
		if *wgs {
			updateInfo(item, item["sampleID"])
			updateGender(item, item["sampleID"])
			item["Gender"] = item["Sex"]
			var inputGender = imInfo[sampleID]["gender"]
			if inputGender != genderMap[sampleID] {
				item["Gender"] = inputGender + "!!!Sequenced" + genderMap[sampleID]
			}
		} else {
			var inputGender = "null"
			if limsInfo[item["Sample"]]["SEX"] == "1" {
				inputGender = "M"
			} else if limsInfo[item["Sample"]]["SEX"] == "2" {
				inputGender = "F"
			} else {
				inputGender = "null"
			}
			if inputGender != genderMap[sampleID] {
				item["Gender"] = inputGender + "!!!Sequenced" + genderMap[sampleID]
			}
			//item["RESULT"]=item[""]
			item["产品编号"] = limsInfo[item["Sample"]]["PRODUCT_CODE"]
		}
	}
	updateColumns(item, sheetTitleMap["QC"])
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
		if *im {
			updateInfo(item, item["sampleID"])
			updateColumns(item, sheetTitleMap[sheetName])
		}
		if *wgs {
			updateInfo(item, item["sampleID"])
			updateGender(item, item["sampleID"])
		}
		writeRow(excel, sheetName, item, title, rIdx)
	}
}

var infoTitle = []string{
	"sampleID",
	"SampleType",
	"Date of Birth",
	"Received Date",
	"ProductID_ProductName",
	"Clinical information",
	"ProductID",
	"TaskID",
	"flow ID",
	"Lane ID",
	"Barcode ID",
	"pipeline",
}

func updateGender(item map[string]string, sampleID string) {
	if *gender == "M" || genderMap[sampleID] == "M" {
		item["Sex"] = "M"
	} else if *gender == "F" || genderMap[sampleID] == "F" {
		item["Sex"] = "F"
	}
}

func updateInfo(item map[string]string, sampleID string) {
	for _, s := range infoTitle {
		item[s] = imInfo[sampleID][s]
	}
	if *cs || *wgs {
		item["期数"] = item["TaskID"]
		item["flow ID"] = item["flow ID"]
		item["产品编号"] = item["ProductID"]
		item["产品编码_产品名称"] = item["ProductID_ProductName"]
	}
}

func updateColumns(item, titleMap map[string]string) {
	for k, v := range titleMap {
		item[v] = item[k]
	}
}

// File2MapArray
func updateDataFile2Sheet(excel *excelize.File, sheetName, path string, fn handleItem) {
	log.Printf("update [%s]", sheetName)
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)

	var db, _ = textUtil.File2MapArray(path, "\t", nil)
	for _, item := range db {
		rIdx++
		fn(item)
		updateINDEX(item, "D", rIdx)
		writeRow(excel, sheetName, item, title, rIdx)
	}
}

// List of File2MapArray
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

func updateDMDCNV(item map[string]string) {
	var sampleID = item["#Sample"]
	item["sampleID"] = sampleID
	updateCnvTags(item, sampleID, item["gene"])
}

func updateSample(item map[string]string) {
	updateColumns(item, sheetTitleMap["Sample"])
}

func updateNator(item map[string]string) {
	var sampleID = item["Sample"]
	item["#sample"] = sampleID
	item["sampleID"] = sampleID
	item["SampleID"] = sampleID
	item["Source"] = "Nator"
	updateABC(item, sampleID)
	updateInfo(item, sampleID)
	item["gender"] = item["Sex"]

	switch item["CNV_type"] {
	case "deletion":
		item["CNV_type"] = "DEL"
	case "duplication":
		item["CNV_type"] = "DUP"
	}
	if item["normalized_RD"] != "" && item["CopyNum"] == "" {
		var ratio, err = strconv.ParseFloat(item["normalized_RD"], 64)
		if err == nil {
			if item["gender"] == "M" && item["Chr"] == "chrX" {
				if ratio <= 0.75 {
					item["CopyNum"] = "0"
				} else if ratio <= 1.25 {
					item["CopyNum"] = "1"
				} else if ratio <= 1.75 {
					item["CopyNum"] = "2"
				} else {
					item["CopyNum"] = "3"
				}
			} else {
				if ratio <= 0.2 {
					item["CopyNum"] = "0"
				} else if ratio <= 0.75 {
					item["CopyNum"] = "1"
				} else if ratio <= 1.25 {
					item["CopyNum"] = "2"
				} else if ratio <= 1.75 {
					item["CopyNum"] = "3"
				} else {
					item["CopyNum"] = "4"
				}
			}
		}
	}
	if item["NM"] == "" && item["Gene"] == "DMD" {
		item["NM"] = "NM_004006.2"
	}
	item["OMIM_EX"] = strings.TrimSuffix(item["OMIM_EX"], ",")
	item["primerDesign"] = strings.Join(
		[]string{
			item["Gene"],
			item["NM"],
			item["OMIM_EX"] + " " + item["CNV_type"],
			"-",
			item["OMIM_EX"],
			item["OMIM_EX"],
			item["杂合性"],
		},
		"; ",
	)

	if *cs {
		item["Chr"] = addChr(item["Chr"])
		item["报告类别"] = "正式报告"
		item["P0_HyperLink"] = filepath.Join("PP100_exon_graph", item["SampleID"]+".DMD.NM_004006.2.png")
	}
}

func updateDMDHyperlLink(item map[string]string) {
	var (
		sampleID  = item["SampleID"]
		pngSuffix = "." + item["gene"] + "." + item["NM"] + ".png"
	)
	item["P0_HyperLink"] = filepath.Join("DMD_exon_graph", item["SampleID"]+pngSuffix)
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

func updateLumpy(item map[string]string) {

	item["Chr"] = item["CHROM"]
	item["Start"] = item["POS"]
	item["End"] = item["END"]
	item["CNV_type"] = item["SVTYPE"]
	item["Gene"] = item["OMIM_Gene"]
	item["OMIM_EX"] = item["OMIM_exon"]

	updateNator(item)
	item["Source"] = "Lumpy"
}

func updateFeature(item map[string]string) {
	item["参考文献"] = strings.ReplaceAll(item["参考文献"], "<br/>", "\n")
	if *wgs {
		updateInfo(item, item["SampleID"])
		updateGender(item, item["SampleID"])
	} else {
		updateABC(item, item["SampleID"])
	}
}
func updateGeneID(item map[string]string) {
	if *wgs {
		updateInfo(item, item["SampleID"])
		updateGender(item, item["SampleID"])
	} else {
		updateABC(item, item["SampleID"])
	}
}

func updateDrug(item map[string]string) {
	if *wgs {
		updateInfo(item, item["样本编号"])
		updateGender(item, item["样本编号"])
	} else {
		updateABC(item, item["样本编号"])
	}
}

func updateBatchCNV(item map[string]string) {
	var sampleID = item["sample"]
	item["sampleID"] = sampleID
	var genes = strings.Split(item["gene"], ",")
	updateCnvTags(item, sampleID, genes...)
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
	updateABC(item, sampleID)
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
	var (
		i    int
		path string
	)
	for i, path = range textUtil.File2Array(list) {
		var axis = simpleUtil.HandleError(excelize.CoordinatesToCellName(1, i+1)).(string)
		simpleUtil.CheckErr(
			excel.SetCellStr(
				*bamPathSheetName,
				axis,
				path,
			),
		)
	}
	if *cs {
		for i2, line := range textUtil.File2Slice(filepath.Join(templatePath, "bam文件路径.txt"), "\t") {
			for j, s := range line {
				var axis = simpleUtil.HandleError(excelize.CoordinatesToCellName(j+1, i+i2+2)).(string)
				simpleUtil.CheckErr(
					excel.SetCellStr(
						*bamPathSheetName,
						axis,
						s,
					),
				)
			}
		}
	}
}
