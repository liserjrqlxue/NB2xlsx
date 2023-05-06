package main

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

func 标签7(item map[string]string, info *GeneInfo) string {
	if item["P/LP*"] == "1" && tag7gene[item["Gene Symbol"]] && info.gender == "F" && 标签1(item, info) != "1-P/LP" {
		return "F-XLR"
	}
	return ""
}
