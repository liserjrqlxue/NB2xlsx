package main

import (
	"regexp"
	"strings"
)

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
)

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

func (info *GeneInfo) new(item map[string]string) *GeneInfo {
	info.基因 = item["Gene Symbol"]
	info.遗传模式 = geneInheritance[info.基因]
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

var (
	af0List = map[string]bool{
		"ESP6500 AF":    true,
		"1000G AF":      true,
		"ExAC AF":       true,
		"ExAC EAS AF":   true,
		"GnomAD AF":     true,
		"GnomAD EAS AF": true,
	}
)

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

func (info *GeneInfo) isAD() bool {
	if info.遗传模式 == "AD" || info.遗传模式 == "AD,AR" || info.遗传模式 == "AD,SMu" || info.遗传模式 == "Mi" || info.遗传模式 == "XLD" || (info.遗传模式 == "XLR" && info.性别 == "M") {
		return true
	}
	return false
}

func (info *GeneInfo) isAR() bool {
	if info.遗传模式 == "AR" || info.遗传模式 == "AR;AR" || (info.遗传模式 == "XLR" && info.性别 == "F") {
		return true
	}
	return false

}

var afThreshold = 1e-4

func (info *GeneInfo) lowADAF(item map[string]string) bool {
	if info.遗传模式 == "AD" || info.遗传模式 == "AD,AR" || info.遗传模式 == "AD,SMu" {
		for af := range af0List {
			if gt(item[af], afThreshold) {
				return false
			}
		}
	}
	return true
}
