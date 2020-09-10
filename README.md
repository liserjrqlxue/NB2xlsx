# NB2xlsx
安馨可生信注释格式

## All variants data

### 过滤
```
现在输出到Sheet1的逻辑（也就是下面邮件的第一个要求）更改如下：
满足以下任一一个条件就输出到sheet1
1.	164基因上Clinvar的标签是Pathogenic或者Likely_pathogenic或者Pathogenic/Likely_pathogenic
2.	164基因上HGMD的标签是DM或者DM？或者DM/DM?
3.	164基因上Clinvar/HGMD数据库外GnomAD频率≤0.01，并且变异类型不包括intron、promoter、no-change、UTR区变异
```
1. 过滤 etc/gene.list.txt 之外的基因
2. "ClinVar Significance" 是 Pathogenic 或者 Likely_pathogenic 或者 Pathogenic/Likely_pathogenic 的保留
3. "HGMD Pred" 是 DM 或者 DM? 或者 DM/DM? 的保留
2. "Function" 不在 etc/function.exclude.txt 中，且 "GnomAD AF" <= 0.01 的保留

### 疾病数据库
```
临床新生儿注释表shee1的Q列“Gene Symbol”与疾病库的C列“基因”相匹配，
匹配上的在sheet1表的BQ列“疾病中文名”输出疾病库的D列“疾病”，sheet1表的BR列“遗传模式”输出疾病库的E列“遗传模式”
```

key1|key2|note
-|-|-
Gene Symbol|基因|main key
疾病中文名|疾病|
遗传模式|遗传模式|

### 已解读数据库
```
第三附件1中的BS列Database，在已解读数据库内并且已解读数据库的DU列是否是包装位点记录为“是”：标记NBS-in
                         在已解读数据库内并且已解读数据库的DU列是否是包装位点无记录：标记NBS-out
                         不在已解读数据库：标记.
第四附件1中的BU列Definition，填写的是已解读数据库中CW列Definition的致病等级
第五附件1中的BW列报告类别，同孕前，数据库内包装的变异（BS列Database标记NBS-in）标记正式报告；数据库外的烈性（LOF列为YES）且低频(GnomAD≤1%，且千人≤1%)：标记补充报告
```
```
sheet1里面的CC列“参考文献”，提取的是已解读数据库中的DM列“Reference”的内容
```

key1|key2|note
-|-|-
Transcript|Transcript|main key 1
cHGVS|cHGVS|main key 2
Definition|Definition|
参考文献|参考文献|

### other
```
第二附件1中的BL列LOF同孕前：nonsense、frameshift、splice-3、splice-5类型且低频(GnomAD≤1%，且千人≤1%)，标记YES，否则标记NO。
```


## excel 格式
### DataValidation
etc/drop.list.txt 包含对应列的下拉表

