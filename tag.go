package main

import "regexp"

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
	if notClinVar[item["ClinVar Significance"]] {
		return false
	}
	if isHGMD[item["HGMD Pred"]] {
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

var (
	isP           = regexp.MustCompile(`P`)
	isI           = regexp.MustCompile(`I`)
	isD           = regexp.MustCompile(`D`)
	isNeutral     = regexp.MustCompile(`neutral`)
	isDeleterious = regexp.MustCompile(`deleterious`)
)

func compositeP(item map[string]string) bool {
	if cdsList[item["Function"]] && item["RepeatTag"] == "" {
		return true
	}
	var count int
	if spliceList[item["Function"]] {
		for _, pred := range []string{
			item["dbscSNV_RF_pred"],
			item["dbscSNV_ADA_pred"],
			item["SpliceAI Pred"],
		} {
			if isP.MatchString(pred) || isI.MatchString(pred) {
				return false
			} else if isD.MatchString(pred) {
				count++
			}
		}
		if isD.MatchString(item["SpliceAI Pred"]) {
			return true
		}
	} else {
		if isNeutral.MatchString(item["Ens Condel Pred"]) {
			return false
		} else if isDeleterious.MatchString(item["Ens Condel Pred"]) {
			count++
		}
		for _, pred := range []string{
			item["SIFT Pred"],
			item["MutationTaster Pred"],
			item["Polyphen2 HVAR Pred"],
		} {
			if isP.MatchString(pred) || isI.MatchString(pred) {
				return false
			} else if isD.MatchString(pred) {
				count++
			}
		}
	}
	if count > 1 {
		return true
	}
	return false
}

type GeneInfo struct {
	基因               string
	遗传模式             string
	性别               string
	PLP, hetPLP, VUS int
	cnv, cnv0        bool
	tag3, tag4       bool
}

func (info *GeneInfo) new(item map[string]string) *GeneInfo {
	info.基因 = item["Gene Symbol"]
	info.遗传模式 = geneInheritance[info.基因]
	info.count(item)
	return info
}

func (info *GeneInfo) count(item map[string]string) {
	if item["自动化判断"] == "VUS" {
		info.VUS++
	}
	if isPLP(item) {
		item["P/LP*"] = "1"
		info.PLP++
		if item["Zygosity"] == "Het" {
			info.hetPLP++
		}
	}
}

func (info *GeneInfo) getTag(item map[string]string) (tag string) {
	tag += 标签1(item, info)
	tag += 标签2(item, info)
	tag += 标签3(item, info)
	tag += 标签5(item)
	return
}

func 标签1(item map[string]string, info *GeneInfo) string {
	if item["P/LP*"] == "1" {
		if info.isAD() && info.lowADAF(item) {
			return "1"
		}
		if info.isAR() {
			if item["Zygosity"] == "Hom" {
				return "1"
			}
			if item["Zygosity"] == "Het" {
				if info.PLP > 1 || info.VUS > 1 || (info.VUS == 1 && item["自动化判断"] != "VUS") {
					return "1"
				}
			}
		}
	} else if item["自动化判断"] == "VUS" && info.hetPLP == 1 && info.isAR() {
		return "1"
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
	tag2Pred = map[string]bool{
		"P":   true,
		"LP":  true,
		"VUS": true,
	}
)

func 标签2(item map[string]string, info *GeneInfo) (tag string) {
	if !compositeP(item) {
		return
	}
	if info.isAD() && item["P/LP*"] != "1" && tag2Pred[item["自动化判断"]] && info.lowADAF(item) {
		return "2"
	}
	if info.isAR() && item["自动化判断"] == "VUS" && (item["Zygosity"] == "Hom" || info.VUS > 1) {
		return "2"
	}
	return
}

func 标签3(item map[string]string, info *GeneInfo) string {
	if info.cnv && info.isAR() {
		if (info.VUS == 0 && item["P/LP*"] == "1") || (item["自动化判断"] == "VUS" && compositeP(item)) {
			info.tag3 = true
			return "3"
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
	if item["Definition"] == "P" || item["Definition"] == "LP" || LOFofPLP[item["Function"]] || isClinVar[item["ClinVar Significance"]] || (isHGMD[item["HGMD Pred"]] && !notClinVar2[item["ClinVar Significance"]]) {
		return "5"
	}
	return ""
}

func (info *GeneInfo) isAD() bool {
	if info.遗传模式 == "AD" || info.遗传模式 == "AD,AR" || info.遗传模式 == "AD,SMu" || info.遗传模式 == "Mi" || ((info.遗传模式 == "XL" || info.遗传模式 == "YL") && info.性别 == "M") {
		return true
	}
	return false
}

func (info *GeneInfo) isAR() bool {
	if info.遗传模式 == "AR" || info.遗传模式 == "AR/AR" || (info.遗传模式 == "XL" && info.性别 == "F") {
		return true
	}
	return false

}

func (info *GeneInfo) lowADAF(item map[string]string) bool {
	if info.遗传模式 == "AD" || info.遗传模式 == "AD,AR" || info.遗传模式 == "AD,SMu" || info.遗传模式 == "YL" {
		for af := range af0List {
			if gt(item[item[af]], 2e-5) {
				return false
			}
		}
	}
	return true
}
