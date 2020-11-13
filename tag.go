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
	isP = regexp.MustCompile(`P`)
	isI = regexp.MustCompile(`I`)
	isD = regexp.MustCompile(`D`)
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

func 标签1(item map[string]string, geneInfo *GeneInfo) (tag string) {
	var (
		遗传模式     = geneInfo.遗传模式
		性别       = geneInfo.性别
		PLP      = geneInfo.PLP
		hetPLP   = geneInfo.hetPLP
		VUS      = geneInfo.VUS
		自动化判断    = item["自动化判断"]
		Zygosity = item["Zygosity"]
	)
	if item["P/LP*"] == "1" {
		if 遗传模式 == "AD" || 遗传模式 == "AD,AR" || 遗传模式 == "AD,SMu" || 遗传模式 == "Mi" || ((遗传模式 == "XL" || 遗传模式 == "YL") && 性别 == "M") {
			tag = "1"
		}
		if 遗传模式 == "AR" || 遗传模式 == "AR/AR" || (遗传模式 == "XL" && 性别 == "F") {
			switch Zygosity {
			case "Hom":
				tag = "1"
			case "Het":
				if PLP > 1 || VUS > 1 || (VUS == 1 && 自动化判断 != "VUS") {
					tag = "1"
				}
			}
		}
	} else if 自动化判断 == "VUS" && hetPLP == 1 && (遗传模式 == "AR" || 遗传模式 == "AR/AR" || (遗传模式 == "XL" && 性别 == "F")) {
		tag = "1"
	}
	return
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

func 标签2(item map[string]string, geneInfo *GeneInfo) (tag string) {
	var (
		遗传模式 = geneInfo.遗传模式
		性别   = geneInfo.性别
		VUS  = geneInfo.VUS
	)
	if 遗传模式 == "AD" || 遗传模式 == "AD,AR" || 遗传模式 == "AD,SMu" || 遗传模式 == "Mi" || ((遗传模式 == "XL" || 遗传模式 == "YL") && 性别 == "M") {
		if item["P/LP*"] == "1" || !tag2Pred[item["自动化判断"]] {
			return
		}
		if 遗传模式 == "AD" || 遗传模式 == "AD,AR" || 遗传模式 == "AD,SMu" || 遗传模式 == "YL" {
			for af := range af0List {
				if gt(item[item[af]], 2e-5) {
					return
				}
			}
		}
		if compositeP(item) {
			tag = "2"
		}
	} else if 遗传模式 == "AR" || 遗传模式 == "AR/AR" || (遗传模式 == "XL" && 性别 == "F") {
		if item["自动化判断"] == "VUS" && (item["Zygosity"] == "Hom" || VUS > 1) && compositeP(item) {
			tag = "2"
		}
	}
	return
}

func 标签3(item map[string]string, geneInfo *GeneInfo) (tag string) {
	var (
		遗传模式  = geneInfo.遗传模式
		性别    = geneInfo.性别
		cnv   = geneInfo.cnv
		VUS   = geneInfo.VUS
		自动化判断 = item["自动化判断"]
	)
	if 遗传模式 == "AR" || 遗传模式 == "AR/AR" || (遗传模式 == "XL" && 性别 == "F") {
		if !cnv {
			return
		}
		if VUS == 0 && item["P/LP*"] == "1" {
			geneInfo.tag3 = true
			tag = "3"
		} else if 自动化判断 == "VUS" && compositeP(item) {
			geneInfo.tag3 = true
			tag = "3"
		}
	}
	return
}

func (info *GeneInfo) 标签4() {
	var (
		遗传模式 = info.遗传模式
		性别   = info.性别
		cnv  = info.cnv
		cnv0 = info.cnv0
	)
	if 遗传模式 == "AD" || 遗传模式 == "AD,AR" || 遗传模式 == "AD,SMu" || 遗传模式 == "Mi" || ((遗传模式 == "XL" || 遗传模式 == "YL") && 性别 == "M") {
		if cnv {
			info.tag4 = true
		}
	}
	if 遗传模式 == "AR" || 遗传模式 == "AR/AR" || (遗传模式 == "XL" && 性别 == "F") {
		if cnv0 {
			info.tag4 = true
		}
	}
}

func 标签5(item map[string]string) (tag string) {
	if item["Definition"] == "P" || item["Definition"] == "LP" || LOFofPLP[item["Function"]] || isClinVar[item["ClinVar Significance"]] || (isHGMD[item["HGMD Pred"]] && !notClinVar2[item["ClinVar Significance"]]) {
		tag = "5"
	}
	return
}
