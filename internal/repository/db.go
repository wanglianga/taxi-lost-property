package repository

import (
	"os"
	"path/filepath"
	"taxi-lost-property/internal/model"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}
	dbPath := filepath.Join(dataDir, "taxi_lost_property.db")

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	err = DB.AutoMigrate(
		&model.LostReport{},
		&model.TaxiOrder{},
		&model.DriverSubmission{},
		&model.ItemInventory{},
		&model.Station{},
		&model.ClaimRecord{},
	)
	if err != nil {
		return err
	}

	return seedData()
}

func seedData() error {
	var count int64
	DB.Model(&model.Station{}).Count(&count)
	if count > 0 {
		return nil
	}

	stations := []model.Station{
		{Name: "总部站点", Address: "北京市朝阳区建国路88号", Phone: "010-12345678", WorkTime: "09:00-21:00"},
		{Name: "海淀站点", Address: "北京市海淀区中关村大街1号", Phone: "010-87654321", WorkTime: "09:00-18:00"},
		{Name: "西城站点", Address: "北京市西城区金融街15号", Phone: "010-11112222", WorkTime: "09:00-18:00"},
	}
	for i := range stations {
		if err := DB.Create(&stations[i]).Error; err != nil {
			return err
		}
	}

	orders := []model.TaxiOrder{
		{
			OrderNo: "ORD20260610001", PlateNo: "京B·12345", FleetCompany: "首汽集团", DriverID: 1001,
			DriverName: "张师傅", DriverPhone: "13800138001", PassengerPhone: "13900139001",
			BoardingPoint: "北京首都机场T3", AlightingPoint: "朝阳区国贸中心",
			Distance: 28.5, Amount: 98.50, PaymentNo: "PAY20260610001",
		},
		{
			OrderNo: "ORD20260610002", PlateNo: "京B·67890", FleetCompany: "北汽集团", DriverID: 1002,
			DriverName: "李师傅", DriverPhone: "13800138002", PassengerPhone: "13900139002",
			BoardingPoint: "北京南站", AlightingPoint: "海淀区中关村",
			Distance: 18.2, Amount: 62.00, PaymentNo: "PAY20260610002",
		},
		{
			OrderNo: "ORD20260610003", PlateNo: "京B·54321", FleetCompany: "首汽集团", DriverID: 1003,
			DriverName: "王师傅", DriverPhone: "13800138003", PassengerPhone: "13900139003",
			BoardingPoint: "大兴国际机场", AlightingPoint: "西城区金融街",
			Distance: 45.0, Amount: 156.00, PaymentNo: "PAY20260610003",
		},
		{
			OrderNo: "ORD20260610004", PlateNo: "京B·99999", FleetCompany: "银建集团", DriverID: 1004,
			DriverName: "赵师傅", DriverPhone: "13800138004", PassengerPhone: "13900139004",
			BoardingPoint: "北京西站", AlightingPoint: "朝阳区三里屯",
			Distance: 15.3, Amount: 52.00, PaymentNo: "PAY20260610004",
		},
		{
			OrderNo: "ORD20260610005", PlateNo: "京B·88888", FleetCompany: "北汽集团", DriverID: 1005,
			DriverName: "刘师傅", DriverPhone: "13800138005", PassengerPhone: "13900139005",
			BoardingPoint: "朝阳区国贸中心", AlightingPoint: "海淀区五道口",
			Distance: 22.1, Amount: 75.50, PaymentNo: "PAY20260610005",
		},
	}
	now := parseTime("2026-06-10 14:30:00")
	for i := range orders {
		orders[i].RideStartTime = now.Add(-1 * time.Duration(2*(i+1)) * time.Hour)
		orders[i].RideEndTime = orders[i].RideStartTime.Add(40 * time.Minute)
		orders[i].PaymentTime = orders[i].RideEndTime
		if err := DB.Create(&orders[i]).Error; err != nil {
			return err
		}
	}

	return nil
}

func parseTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return t
}
