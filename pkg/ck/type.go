package ck

// DWS OUT FACTORY EMS_V
type CimData struct {
	Factory string `gorm:"column:FACTORY"`
	Time    string `gorm:"column:DATE_TIMEKEY"`
	Count   int    `gorm:"column:OUT_TOTAL"`
}

// ems_manual_production
type EmsData struct {
	Id     string `gorm:"column:id"`
	PYear  string `gorm:"column:p_year"`
	PMonth string `gorm:"column:p_month"`
	PData  string `gorm:"column:p_data"`
}
