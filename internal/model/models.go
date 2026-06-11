package model

import (
	"time"
)

type LostReportStatus string

const (
	StatusPendingMatch   LostReportStatus = "pending_match"
	StatusNoMatch        LostReportStatus = "no_match"
	StatusDriverNotFound LostReportStatus = "driver_not_found"
	StatusSubmitted      LostReportStatus = "submitted"
	StatusMultipleClaims LostReportStatus = "multiple_claims"
	StatusVerified       LostReportStatus = "verified"
	StatusReturned       LostReportStatus = "returned"
)

type ClaimStatus string

const (
	ClaimStatusPending  ClaimStatus = "pending"
	ClaimStatusVerified ClaimStatus = "verified"
	ClaimStatusRejected ClaimStatus = "rejected"
	ClaimStatusReturned ClaimStatus = "returned"
)

type ReturnMethod string

const (
	ReturnMethodPickup  ReturnMethod = "pickup"
	ReturnMethodExpress ReturnMethod = "express"
)

type LostReport struct {
	ID              int64            `json:"id" gorm:"primaryKey"`
	PassengerName   string           `json:"passenger_name" gorm:"size:100;not null"`
	PassengerPhone  string           `json:"passenger_phone" gorm:"size:20;not null"`
	PassengerIDCard string           `json:"passenger_id_card,omitempty" gorm:"size:30"`
	RideTime        time.Time        `json:"ride_time" gorm:"not null"`
	BoardingPoint   string           `json:"boarding_point" gorm:"size:200;not null"`
	AlightingPoint  string           `json:"alighting_point" gorm:"size:200;not null"`
	PaymentNo       string           `json:"payment_no" gorm:"size:50"`
	Amount          float64          `json:"amount"`
	PlateNo         string           `json:"plate_no" gorm:"size:20"`
	DriverName      string           `json:"driver_name" gorm:"size:100"`
	ItemDescription string           `json:"item_description" gorm:"size:1000;not null"`
	ItemCategory    string           `json:"item_category" gorm:"size:50"`
	ItemValue       float64          `json:"item_value"`
	IsValuable      bool             `json:"is_valuable"`
	Photos          string           `json:"photos" gorm:"size:1000"`
	Status          LostReportStatus `json:"status" gorm:"size:30;default:pending_match"`
	MatchedOrderID  int64            `json:"matched_order_id,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

type TaxiOrder struct {
	ID             int64     `json:"id" gorm:"primaryKey"`
	OrderNo        string    `json:"order_no" gorm:"size:50;uniqueIndex;not null"`
	PlateNo        string    `json:"plate_no" gorm:"size:20;not null"`
	DriverID       int64     `json:"driver_id" gorm:"not null"`
	DriverName     string    `json:"driver_name" gorm:"size:100;not null"`
	DriverPhone    string    `json:"driver_phone" gorm:"size:20;not null"`
	PassengerPhone string    `json:"passenger_phone" gorm:"size:20"`
	RideStartTime  time.Time `json:"ride_start_time" gorm:"not null"`
	RideEndTime    time.Time `json:"ride_end_time" gorm:"not null"`
	BoardingPoint  string    `json:"boarding_point" gorm:"size:200;not null"`
	AlightingPoint string    `json:"alighting_point" gorm:"size:200;not null"`
	Distance       float64   `json:"distance"`
	Amount         float64   `json:"amount"`
	PaymentNo      string    `json:"payment_no" gorm:"size:50"`
	PaymentTime    time.Time `json:"payment_time"`
	CreatedAt      time.Time  `json:"created_at"`
}

type DriverSubmission struct {
	ID              int64     `json:"id" gorm:"primaryKey"`
	ReportID        int64     `json:"report_id" gorm:"index"`
	OrderID         int64     `json:"order_id" gorm:"index;not null"`
	DriverID        int64     `json:"driver_id" gorm:"not null"`
	DriverName      string    `json:"driver_name" gorm:"size:100;not null"`
	PlateNo         string    `json:"plate_no" gorm:"size:20;not null"`
	ItemDescription string    `json:"item_description" gorm:"size:1000;not null"`
	ItemCategory    string    `json:"item_category" gorm:"size:50"`
	ItemValue       float64   `json:"item_value"`
	IsValuable      bool      `json:"is_valuable"`
	Photos          string    `json:"photos" gorm:"size:1000"`
	FoundLocation   string    `json:"found_location" gorm:"size:200"`
	FoundTime       time.Time `json:"found_time" gorm:"not null"`
	SubmitTime      time.Time `json:"submit_time" gorm:"not null"`
	Remark          string    `json:"remark" gorm:"size:500"`
	CreatedAt       time.Time `json:"created_at"`
}

type ItemInventory struct {
	ID              int64     `json:"id" gorm:"primaryKey"`
	SubmissionID    int64     `json:"submission_id" gorm:"uniqueIndex;not null"`
	ReportID        int64     `json:"report_id" gorm:"index"`
	ItemDescription string    `json:"item_description" gorm:"size:1000;not null"`
	ItemCategory    string    `json:"item_category" gorm:"size:50"`
	ItemValue       float64   `json:"item_value"`
	IsValuable      bool      `json:"is_valuable"`
	Photos          string    `json:"photos" gorm:"size:1000"`
	StationID       int64     `json:"station_id" gorm:"not null"`
	StationName     string    `json:"station_name" gorm:"size:100;not null"`
	CabinetNo       string    `json:"cabinet_no" gorm:"size:50;not null"`
	StoredBy        string    `json:"stored_by" gorm:"size:100;not null"`
	StoredAt        time.Time `json:"stored_at" gorm:"not null"`
	Status          string    `json:"status" gorm:"size:30;default:in_stock"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Station struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	Address   string    `json:"address" gorm:"size:300;not null"`
	Phone     string    `json:"phone" gorm:"size:20"`
	WorkTime  string    `json:"work_time" gorm:"size:100"`
	CreatedAt time.Time `json:"created_at"`
}

type ClaimRecord struct {
	ID                int64        `json:"id" gorm:"primaryKey"`
	ReportID          int64        `json:"report_id" gorm:"index;not null"`
	InventoryID       int64        `json:"inventory_id" gorm:"index;not null"`
	PassengerName     string       `json:"passenger_name" gorm:"size:100;not null"`
	PassengerPhone    string       `json:"passenger_phone" gorm:"size:20;not null"`
	PassengerIDCard   string       `json:"passenger_id_card,omitempty" gorm:"size:30"`
	VerifyMaterials   string       `json:"verify_materials" gorm:"size:1000"`
	VerifyDescription string       `json:"verify_description" gorm:"size:1000"`
	VerifiedBy        string       `json:"verified_by" gorm:"size:100"`
	VerifiedAt        *time.Time   `json:"verified_at"`
	Status            ClaimStatus  `json:"status" gorm:"size:30;default:pending"`
	ReturnMethod      ReturnMethod `json:"return_method" gorm:"size:20"`
	ExpressNo         string       `json:"express_no" gorm:"size:50"`
	ExpressCompany    string       `json:"express_company" gorm:"size:50"`
	ReceiverName      string       `json:"receiver_name" gorm:"size:100"`
	ReceiverPhone     string       `json:"receiver_phone" gorm:"size:20"`
	ReceiverAddress   string       `json:"receiver_address" gorm:"size:300"`
	PickupStationID   int64        `json:"pickup_station_id"`
	PickupStationName string       `json:"pickup_station_name" gorm:"size:100"`
	PickupTime        *time.Time   `json:"pickup_time"`
	PickupPerson      string       `json:"pickup_person" gorm:"size:100"`
	ReturnedAt        *time.Time   `json:"returned_at"`
	Remark            string       `json:"remark" gorm:"size:500"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
}

type MatchResult struct {
	Order     TaxiOrder `json:"order"`
	MatchRate float64   `json:"match_rate"`
}
