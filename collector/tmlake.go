package collector

import (
	"fmt"
	"github.com/gosnmp/gosnmp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/snmp_exporter/config"
	"github.com/prometheus/snmp_exporter/utils"
	"strconv"
	"strings"
)

const moduleNameTmLake = "tmlake"
const (
	TmLakeMetricAdcCpuInfoUsr            = "cpu_usage"
	TmLakeMetricAdcDiskInfoDesc          = "filesystem_usage"
	TmLakeMetricAdcMemInfoUsedPercent    = "memory_usage"
	TmLakeMetricAdcLLBAppStatus          = "llb_vserver_alive"
	TmLakeMetricAdcMonInfoCpuTemp        = "cpu_temperature"
	TmLakeMetricAdcLLBAppName            = "adcLLBAppName"
	TmLakeMetricAdcInterfaceStatThrputRx = "adcInterfaceStatThrputRx" //接收
	TmLakeMetricAdcInterfaceStatThrputTx = "adcInterfaceStatThrputTx" //发送
	TmLakeMetricAdcInterfaceStatName     = "adcInterfaceStatName"
)

const (
	InterfaceStatNameOid = "1.3.6.1.4.1.99999.1.2.2.2.1.2"
	AdcLLBAppNameOid     = "1.3.6.1.4.1.99999.1.6.1.2.1.2"
)

type TmLakeRe struct {
	Metric      *config.Metric
	ValueType   prometheus.ValueType
	IndexOids   []int
	LabelNames  []string
	LabelValues []string
	Value       float64
	C           collector
}

type LabelValues map[string][]string

func (t *TmLakeRe) TmLakeReset() prometheus.Metric {

	var err error
	switch t.Metric.Name {
	case TmLakeMetricAdcCpuInfoUsr:
		err = t.reTmLakeMetricAdcCpuInfoUsr()

	case TmLakeMetricAdcMemInfoUsedPercent:
		err = t.reTmLakeMetricAdcMemInfoUsedPercent()

	case TmLakeMetricAdcDiskInfoDesc:
		err = t.reTmLakeMetricAdcDiskInfoDesc()
	case TmLakeMetricAdcMonInfoCpuTemp:
		err = t.reTmLakeMetricAdcMonInfoCpuTemp()

	case TmLakeMetricAdcInterfaceStatThrputRx:
		err = t.reTmLakeMetricAdcInterfaceStatThrput(t.C)
	case TmLakeMetricAdcInterfaceStatThrputTx:
		err = t.reTmLakeMetricAdcInterfaceStatThrput(t.C)

	case TmLakeMetricAdcLLBAppStatus:
		err = t.reTmLakeMetricAdcLLBAppStatus(t.C)

	}
	if err != nil {
		return nil
	}

	sample, err := prometheus.NewConstMetric(prometheus.NewDesc(t.Metric.Name, t.Metric.Help, t.LabelNames, nil),
		t.ValueType, t.Value, t.LabelValues...)
	if err != nil {
		sample = prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "Error calling NewConstMetric", nil, nil),
			fmt.Errorf("error for metric %s with labels %v from indexOids %v: %v", t.Metric.Name, t.LabelValues, t.IndexOids, err))
	}

	return sample
}

func (t *TmLakeRe) reTmLakeMetricAdcDiskInfoDesc() error {
	if len(t.LabelValues) < 2 {
		return fmt.Errorf("AdcDiskInfoDesc TmLakeReset labelValues len err")
	}

	// 磁盘取678 磁盘分区
	distMap := map[string]int{"6": 1, "7": 1, "8": 1}
	if _, ok := distMap[t.LabelValues[0]]; !ok {
		return fmt.Errorf("Not target data ")
	}

	// 更改值 将标签中得使用率
	s := utils.Hex2String(t.LabelValues[1])
	t.LabelValues[1] = s

	spList := strings.Split(s, " ")
	splist2 := make([]string, 6)
	i := 0
	for _, sp := range spList {
		if sp != "" {
			splist2[i] = strings.TrimRight(sp, "\n")
			i++
		}
	}

	// 重新赋值
	mountpoint := ""
	size := ""
	used := ""
	avail := ""
	filesystem := ""
	if len(spList) >= 6 {
		usageStr := strings.TrimRight(splist2[4], "%")
		t.Value, _ = strconv.ParseFloat(usageStr, 64)
		mountpoint = splist2[5]
		size = splist2[1]
		used = splist2[2]
		avail = splist2[3]
		filesystem = splist2[0]
	}
	// 处理标签增加挂载点标签
	t.LabelNames = t.LabelNames[:1]
	t.LabelValues = t.LabelValues[:1]

	t.LabelNames = append(t.LabelNames, "mountpoint")
	t.LabelValues = append(t.LabelValues, mountpoint)

	t.LabelNames = append(t.LabelNames, "size")
	t.LabelValues = append(t.LabelValues, size)

	t.LabelNames = append(t.LabelNames, "used")
	t.LabelValues = append(t.LabelValues, used)

	t.LabelNames = append(t.LabelNames, "avail")
	t.LabelValues = append(t.LabelValues, avail)

	t.LabelNames = append(t.LabelNames, "filesystem")
	t.LabelValues = append(t.LabelValues, filesystem)

	return nil
}

func (t *TmLakeRe) reTmLakeMetricAdcMemInfoUsedPercent() error {
	if len(t.LabelValues) < 1 {
		return fmt.Errorf("memory_usage TmLakeReset labelValues len err")
	}
	// memory adcMemInfoIndex==1  其他的舍弃
	if t.LabelValues[0] != "1" {
		return fmt.Errorf("Not target data")
	}
	return nil

}

func (t *TmLakeRe) reTmLakeMetricAdcCpuInfoUsr() error {
	if len(t.LabelValues) < 2 {
		return fmt.Errorf("CpuInfoUsr TmLakeReset labelValues len err")
	}
	// cpu取平均值 adcCpuInfoindex==0  其他的舍弃
	if t.LabelValues[0] != "0" {
		return fmt.Errorf("Not target data ")
	}
	// 更改值 取标签中的adcCpuInfoUsr 作为值 并转化成十进制
	s := utils.Hex2String(t.LabelValues[1])
	t.Value, _ = strconv.ParseFloat(s, 64)
	t.LabelValues[1] = s
	return nil
}

func (t *TmLakeRe) reTmLakeMetricAdcMonInfoCpuTemp() error {
	if len(t.LabelValues) < 1 {
		return fmt.Errorf("CpuTemp TmLakeReset labelValues len err")
	}

	// 更改值 并转化成十进制
	s := utils.Hex2String(t.LabelValues[0])
	t.Value, _ = strconv.ParseFloat(s, 64)
	t.LabelValues[0] = s
	return nil
}

// 吞吐量逻辑相同
func (t *TmLakeRe) reTmLakeMetricAdcInterfaceStatThrput(c collector) error {
	if len(t.LabelValues) < 1 {
		return fmt.Errorf("reTmLakeMetricAdcInterfaceStatThrput labelValues len err")
	}
	// 请求 接口名称
	module := NewTmLakeModule(c, InterfaceStatNameOid, TmLakeMetricAdcInterfaceStatName)
	results, err := ScrapeTarget(c.ctx, c.target, module, c.logger)
	if err != nil {
		return fmt.Errorf("reTmLakeMetricAdcInterfaceStatThrput get  TmLakeMetricAdcInterfaceStatName err")
	}
	indexMap := indexMap(results.pdus)

	// 添加interface name标签
	adcInterfaceStatIndex := t.LabelValues[0]
	t.LabelNames = append(t.LabelNames, "interface_name")
	t.LabelValues = append(t.LabelValues, indexMap[adcInterfaceStatIndex])

	return nil
}

func (t *TmLakeRe) reTmLakeMetricAdcLLBAppStatus(c collector) error {
	if len(t.LabelValues) < 1 {
		return fmt.Errorf("reTmLakeMetricAdcLLBAppStatus labelValues len err")
	}
	// 请求 接口名称
	module := NewTmLakeModule(c, AdcLLBAppNameOid, TmLakeMetricAdcLLBAppName)
	results, err := ScrapeTarget(c.ctx, c.target, module, c.logger)
	if err != nil {
		return fmt.Errorf("reTmLakeMetricAdcLLBAppStatus get  TmLakeMetricAdcInterfaceStatName err")
	}

	indexMap := indexMap(results.pdus)

	// 添加interface name标签
	adcInterfaceStatIndex := t.LabelValues[0]
	t.LabelNames = append(t.LabelNames, "vserver_name")
	t.LabelValues = append(t.LabelValues, indexMap[adcInterfaceStatIndex])

	return nil
}

func indexMap(pdus []gosnmp.SnmpPDU) map[string]string {
	indexMap := make(map[string]string, len(pdus))
	for _, pdu := range pdus {
		index := 0
		if len(pdu.Name) > 0 {
			index = len(pdu.Name) - 1
		}
		value := string(pdu.Value.([]uint8))
		indexMap[pdu.Name[index:]] = value
	}

	return indexMap
}

func NewTmLakeModule(c collector, oid, name string) *config.Module {
	return &config.Module{
		Name: moduleNameTmLake,
		Walk: []string{oid},
		Get:  nil,
		Metrics: []*config.Metric{&config.Metric{
			Name: name,
			Oid:  oid,
			Type: "OctetString",
			Help: "",
			Indexes: []*config.Index{&config.Index{
				Labelname:  name,
				Type:       "gauge",
				FixedSize:  0,
				Implied:    false,
				EnumValues: nil,
			}},
			Lookups:        nil,
			RegexpExtracts: nil,
			EnumValues:     nil,
		}},
		WalkParams: c.module.WalkParams,
	}
}
