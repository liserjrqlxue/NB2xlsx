package main

import (
	"path/filepath"
	"strings"

	"github.com/liserjrqlxue/goUtil/textUtil"
)

func isPLP(item map[string]string) bool {
	if item["Definition"] == "P" || item["Definition"] == "LP" {
		return true
	}
	if LOFofPLP[item["Function"]] {
		return true
	}
	if isClinVar[item["ClinVar Significance"]] {
		return true
	}
	if isHGMD[item["HGMD Pred"]] && !notClinVar[item["ClinVar Significance"]] {
		return true
	}
	return false
}

func isPLP2(item map[string]string) bool {
	if item["Definition"] == "P" || item["Definition"] == "LP" {
		return true
	}
	if LOFofPLP[item["Function"]] {
		return true
	}
	if isClinVar[item["ClinVar Significance"]] {
		return true
	}
	if isHGMD[item["HGMD Pred"]] && !notClinVar2[item["ClinVar Significance"]] {
		return true
	}
	return false
}

func isVUS(item map[string]string) bool {
	if notClinVar[item["ClinVar Significance"]] {
		return false
	}
	if item["Definition"] == "VUS" || isPLPVUS.MatchString(item["自动化判断"]) {
		return true
	}
	return false
}

func spliceP(item map[string]string) (count int) {
	for _, pred := range []string{
		item["dbscSNV_RF_pred"],
		item["dbscSNV_ADA_pred"],
		item["SpliceAI Pred"],
	} {
		if isP.MatchString(pred) || isI.MatchString(pred) {
			return 0
		} else if isD.MatchString(pred) {
			count++
		}
	}
	if isD.MatchString(item["SpliceAI Pred"]) {
		return 2
	}
	return
}

func noSpliceP(item map[string]string) (count int) {
	if isNeutral.MatchString(item["Ens Condel Pred"]) {
		return 0
	} else if isDeleterious.MatchString(item["Ens Condel Pred"]) {
		count++
	}
	for _, pred := range []string{
		item["SIFT Pred"],
		item["MutationTaster Pred"],
		item["Polyphen2 HVAR Pred"],
	} {
		if isP.MatchString(pred) || isI.MatchString(pred) {
			return 0
		} else if isD.MatchString(pred) {
			count++
		}
	}
	return
}

func compositeP(item map[string]string) bool {
	if cdsList[item["Function"]] && item["RepeatTag"] == "" {
		return true
	}
	var count int
	if spliceList[item["Function"]] {
		count = spliceP(item)
	} else {
		count = noSpliceP(item)
	}
	if count > 1 {
		return true
	}
	return false
}

func compositePCS(item map[string]string) bool {
	if cdsList[item["Function"]] && item["RepeatTag"] == "" {
		return true
	}
	var count int
	if spliceCSList[item["Function"]] {
		count = spliceP(item)
	} else {
		count = noSpliceP(item)
	}
	if count > 1 {
		return true
	}
	return false
}

func (info *GeneInfo) new(item map[string]string) *GeneInfo {
	info.gene = item["Gene Symbol"]
	info.inheritance = geneInheritance[info.gene]
	return info
}

func (info *GeneInfo) count(item map[string]string) {

	if item["自动化判断"] == "VUS" {
		if item["Zygosity"] == "Het" && compositeP(item) {
			info.HpVUS++
		}
	}
	if isPLP(item) {
		item["P/LP*"] = "1"
		info.PLP++
	} else if isVUS(item) {
		item["VUS*"] = "1"
		info.VUS++
	}
}

func (info *GeneInfo) getTag(item map[string]string) (tag string) {
	var tags []string
	var tag1 = 标签1(item, info)
	if tag1 != "" {
		tags = append(tags, tag1)
	}
	var tag2 = 标签2(item, info)
	if tag2 != "" {
		tags = append(tags, tag2)
	}
	var tag3 = 标签3(item, info)
	if tag3 != "" {
		tags = append(tags, tag3)
	}
	var tag5 = 标签5(item)
	if tag5 != "" {
		tags = append(tags, tag5)
	}
	var tag6 = 标签6(item)
	if tag6 != "" {
		tags = append(tags, tag6)
	}
	var tag7 = 标签7(item, info)
	if tag7 != "" {
		tags = append(tags, tag7)
	}
	return strings.Join(tags, ";")
}

func 标签1(item map[string]string, info *GeneInfo) string {
	if item["P/LP*"] == "1" {
		if info.isAD() && info.lowADAF(item) {
			return "1-P/LP"
		}
		if info.isAR() {
			if item["Zygosity"] == "Hom" {
				return "1-P/LP"
			}
			if item["Zygosity"] == "Het" {
				if info.PLP > 1 {
					return "1-P/LP"
				}
				if info.VUS > 0 {
					return "1-VUS"
				}
			}
		}
	} else if item["VUS*"] == "1" && info.PLP > 0 && info.isAR() {
		return "1-VUS"
	}
	return ""
}

func 标签2(item map[string]string, info *GeneInfo) (tag string) {
	if item["VUS*"] != "1" {
		return
	}
	if !compositePCS(item) {
		return
	}
	if info.isAD() && info.lowADAF(item) {
		return "2-AD/XL-VUS"
	}
	if info.isAR() && (item["Zygosity"] == "Hom" || info.VUS > 1) {
		return "2-AR/XLR-VUS"
	}
	return
}

func 标签3(item map[string]string, info *GeneInfo) string {
	if info.cnv && info.isAR() {
		if item["P/LP*"] == "1" {
			info.tag3 = "3-P/LP/CNV"
			return "3-P/LP/CNV"
		} else if item["VUS*"] == "1" && compositeP(item) {
			if info.tag3 == "" {
				info.tag3 = "3-VUS/CNV"
			}
			return "3-VUS/CNV"
		}
	}
	return ""
}

func (info *GeneInfo) 标签4() {
	if info.cnv && info.isAD() {
		info.tag4 = true
	} else if info.cnv0 && info.isAR() {
		info.tag4 = true
	}
}

func 标签5(item map[string]string) string {
	if item["isReport"] == "Y" {
		return "5"
	}
	return ""
}

func 标签6(item map[string]string) string {
	if isPLP2(item) {
		return "6"
	}
	return ""
}

func init() {
	for _, gene := range textUtil.File2Array(filepath.Join(etcPath, "tag7.gene.txt")) {
		tag7gene[gene] = true
	}
}

func 标签7(item map[string]string, info *GeneInfo) string {
	if item["P/LP*"] == "1" && tag7gene[item["Gene Symbol"]] && info.性别 == "F" && 标签1(item, info) != "1-P/LP" {
		return "F-XLR"
	}
	return ""
}

func (info *GeneInfo) isAD() bool {
	switch info.inheritance {
	case "AD", "AD,AR", "AD,SMu", "Mi", "XLD", "XL":
		return true
	case "XLR":
		if info.性别 == "M" {
			return true
		} else {
			return false
		}
	default:
		return false
	}
}

func (info *GeneInfo) isAR() bool {
	if info.inheritance == "AR" || info.inheritance == "AR;AR" || (info.inheritance == "XLR" && info.性别 == "F") {
		return true
	}
	return false

}

func (info *GeneInfo) lowADAF(item map[string]string) bool {
	if info.inheritance == "AD" || info.inheritance == "AD,AR" || info.inheritance == "AD,SMu" {
		for af := range af0List {
			if gt(item[af], afThreshold) {
				return false
			}
		}
	}
	return true
}
