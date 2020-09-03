# NB2xlsx
安馨可生信注释格式

## 添加疾病数据库
临床新生儿注释表shee1的Q列“Gene Symbol”与疾病库的C列“基因”相匹配，
匹配上的在sheet1表的BQ列“疾病中文名”输出疾病库的D列“疾病”，sheet1表的BR列“遗传模式”输出疾病库的E列“遗传模式”

## excel 格式
### DataValidation
etc/drop.list.txt 包含对应列的下拉表

## 过滤
1. 保留 etc/gene.list.txt 中的基因
2. 过滤 etc/function.exclude.txt 中的 Function
3. 过滤 GnomAD AF > 0.01