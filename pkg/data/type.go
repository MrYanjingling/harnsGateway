package data

type MqConfig struct {
	Exchange    string
	VirtualHost string
	Queue       string
}

type TimeSeries struct {
	TagName string `gorm:"column:tagName"`
	Time    string `gorm:"column:time"`
	Value   string `gorm:"column:value"`
}

type Spc struct {
	ToolId       string  `json:"tool_id"`
	LineId       string  `json:"line_id"`
	RptTimestamp string  `json:"rpt_timestamp"`
	RptUsr       string  `json:"rpt_usr"`
	FabIdFk      string  `json:"fab_id_fk"`
	SampSize     string  `json:"samp_size"`
	ItemNodes    []*Item `json:"item_nodes"`
}

func NewSpc(item []*Item, ts string) *Spc {
	return &Spc{
		ToolId:       "IT-FMCS-01",
		LineId:       "ITBL",
		RptTimestamp: ts,
		RptUsr:       "FMCS",
		FabIdFk:      "D2",
		SampSize:     "",
		ItemNodes:    item,
	}
}

type Item struct {
	DataGroup    string `json:"data_group"`
	DataType     string `json:"data_type"`
	DataSeq      string `json:"data_seq"`
	DataGroupSeq string `json:"data_group_seq"`
	DataValue    string `json:"data_value"`
	CountingTye  int    `json:"counting_tye"`
}

func NewItem(dataGroup string, index string, dataValue string) *Item {
	return &Item{
		DataGroup:    dataGroup,
		DataType:     "Number",
		DataSeq:      index,
		DataGroupSeq: index,
		DataValue:    dataValue,
		CountingTye:  0,
	}
}

type Tag struct {
	TagNames []string `json:"tagNames"`
}

type ResponseModel struct {
	TagNames interface{} `json:"tagNames,omitempty"`
}
