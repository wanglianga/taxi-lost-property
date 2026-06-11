package service

import (
	"errors"
	"fmt"
	"strings"
	"taxi-lost-property/internal/model"
	"taxi-lost-property/internal/repository"
	"time"
)

type LostReportRequest struct {
	PassengerName   string    `json:"passenger_name" binding:"required"`
	PassengerPhone  string    `json:"passenger_phone" binding:"required"`
	PassengerIDCard string    `json:"passenger_id_card"`
	RideTime        time.Time `json:"ride_time" binding:"required"`
	BoardingPoint   string    `json:"boarding_point" binding:"required"`
	AlightingPoint  string    `json:"alighting_point" binding:"required"`
	PaymentNo       string    `json:"payment_no"`
	Amount          float64   `json:"amount"`
	PlateNo         string    `json:"plate_no"`
	DriverName      string    `json:"driver_name"`
	ItemDescription string    `json:"item_description" binding:"required"`
	ItemCategory    string    `json:"item_category"`
	ItemValue       float64   `json:"item_value"`
	IsValuable      bool      `json:"is_valuable"`
	Photos          string    `json:"photos"`
}

type DriverSubmissionRequest struct {
	ReportID        int64     `json:"report_id"`
	OrderID         int64     `json:"order_id" binding:"required"`
	DriverID        int64     `json:"driver_id" binding:"required"`
	DriverName      string    `json:"driver_name" binding:"required"`
	PlateNo         string    `json:"plate_no" binding:"required"`
	ItemDescription string    `json:"item_description" binding:"required"`
	ItemCategory    string    `json:"item_category"`
	ItemValue       float64   `json:"item_value"`
	IsValuable      bool      `json:"is_valuable"`
	Photos          string    `json:"photos"`
	FoundLocation   string    `json:"found_location"`
	FoundTime       time.Time `json:"found_time" binding:"required"`
	Remark          string    `json:"remark"`
}

type InventoryRequest struct {
	SubmissionID int64  `json:"submission_id" binding:"required"`
	ReportID     int64  `json:"report_id"`
	StationID    int64  `json:"station_id" binding:"required"`
	CabinetNo    string `json:"cabinet_no" binding:"required"`
	StoredBy     string `json:"stored_by" binding:"required"`
}

type ClaimRequest struct {
	ReportID          int64              `json:"report_id" binding:"required"`
	InventoryID       int64              `json:"inventory_id" binding:"required"`
	PassengerName     string             `json:"passenger_name" binding:"required"`
	PassengerPhone    string             `json:"passenger_phone" binding:"required"`
	PassengerIDCard   string             `json:"passenger_id_card"`
	VerifyMaterials   string             `json:"verify_materials"`
	VerifyDescription string             `json:"verify_description" binding:"required"`
	ReturnMethod      model.ReturnMethod `json:"return_method" binding:"required"`
	ExpressNo         string             `json:"express_no"`
	ExpressCompany    string             `json:"express_company"`
	ReceiverName      string             `json:"receiver_name"`
	ReceiverPhone     string             `json:"receiver_phone"`
	ReceiverAddress   string             `json:"receiver_address"`
	PickupStationID   int64              `json:"pickup_station_id"`
	Remark            string             `json:"remark"`
}

type VerifyClaimRequest struct {
	VerifiedBy     string `json:"verified_by" binding:"required"`
	Pass           bool   `json:"pass" binding:"required"`
	RejectReason   string `json:"reject_reason"`
}

type ReturnConfirmRequest struct {
	Operator  string    `json:"operator" binding:"required"`
	ReturnedAt time.Time `json:"returned_at"`
}

func CreateLostReport(req *LostReportRequest) (*model.LostReport, error) {
	if req.IsValuable && req.PassengerIDCard == "" {
		return nil, errors.New("贵重物品遗失必须提供身份证号进行实名登记")
	}

	report := &model.LostReport{
		PassengerName:   req.PassengerName,
		PassengerPhone:  req.PassengerPhone,
		PassengerIDCard: req.PassengerIDCard,
		RideTime:        req.RideTime,
		BoardingPoint:   req.BoardingPoint,
		AlightingPoint:  req.AlightingPoint,
		PaymentNo:       req.PaymentNo,
		Amount:          req.Amount,
		PlateNo:         req.PlateNo,
		DriverName:      req.DriverName,
		ItemDescription: req.ItemDescription,
		ItemCategory:    req.ItemCategory,
		ItemValue:       req.ItemValue,
		IsValuable:      req.IsValuable,
		Photos:          req.Photos,
		Status:          model.StatusPendingMatch,
	}

	if err := repository.DB.Create(report).Error; err != nil {
		return nil, err
	}

	return report, nil
}

func GetLostReport(id int64) (*model.LostReport, error) {
	var report model.LostReport
	if err := repository.DB.First(&report, id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

func ListLostReports(status string, page, size int) ([]model.LostReport, int64, error) {
	var reports []model.LostReport
	var total int64

	query := repository.DB.Model(&model.LostReport{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	if err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

func MatchOrders(reportID int64) ([]model.MatchResult, error) {
	report, err := GetLostReport(reportID)
	if err != nil {
		return nil, err
	}

	var orders []model.TaxiOrder
	query := repository.DB

	if report.PlateNo != "" {
		query = query.Where("plate_no LIKE ?", "%"+report.PlateNo+"%")
	}
	if report.PaymentNo != "" {
		query = query.Or("payment_no LIKE ?", "%"+report.PaymentNo+"%")
	}

	timeWindow := 2 * time.Hour
	startTime := report.RideTime.Add(-timeWindow)
	endTime := report.RideTime.Add(timeWindow)
	query = query.Where("ride_start_time BETWEEN ? AND ? OR ride_end_time BETWEEN ? AND ?",
		startTime, endTime, startTime, endTime)

	if err := query.Find(&orders).Error; err != nil {
		return nil, err
	}

	var results []model.MatchResult
	for _, order := range orders {
		matchRate := calculateMatchRate(report, &order)
		if matchRate > 0.1 {
			results = append(results, model.MatchResult{
				Order:     order,
				MatchRate: matchRate,
			})
		}
	}

	if len(results) == 0 {
		report.Status = model.StatusNoMatch
		repository.DB.Save(report)
	}

	return results, nil
}

func calculateMatchRate(report *model.LostReport, order *model.TaxiOrder) float64 {
	var score float64
	var totalWeight float64

	if report.PlateNo != "" && strings.Contains(strings.ToLower(order.PlateNo), strings.ToLower(report.PlateNo)) {
		score += 40
	}
	totalWeight += 40

	if report.PaymentNo != "" && strings.Contains(order.PaymentNo, report.PaymentNo) {
		score += 30
	}
	totalWeight += 30

	if report.Amount > 0 && order.Amount > 0 {
		diff := abs(report.Amount - order.Amount)
		if diff < 5 {
			score += 10
		} else if diff < 15 {
			score += 5
		}
	}
	totalWeight += 10

	rideDiff := report.RideTime.Sub(order.RideStartTime).Abs()
	if rideDiff < 30*time.Minute {
		score += 10
	} else if rideDiff < time.Hour {
		score += 5
	}
	totalWeight += 10

	if strings.Contains(strings.ToLower(order.BoardingPoint), strings.ToLower(report.BoardingPoint)) ||
		strings.Contains(strings.ToLower(report.BoardingPoint), strings.ToLower(order.BoardingPoint)) {
		score += 5
	}
	totalWeight += 5

	if strings.Contains(strings.ToLower(order.AlightingPoint), strings.ToLower(report.AlightingPoint)) ||
		strings.Contains(strings.ToLower(report.AlightingPoint), strings.ToLower(order.AlightingPoint)) {
		score += 5
	}
	totalWeight += 5

	if totalWeight == 0 {
		return 0
	}
	return score / totalWeight
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func ConfirmMatchedOrder(reportID, orderID int64) error {
	report, err := GetLostReport(reportID)
	if err != nil {
		return err
	}

	var order model.TaxiOrder
	if err := repository.DB.First(&order, orderID).Error; err != nil {
		return errors.New("订单不存在")
	}

	report.MatchedOrderID = orderID
	report.PlateNo = order.PlateNo
	report.DriverName = order.DriverName
	if report.Status == model.StatusNoMatch || report.Status == model.StatusPendingMatch {
		report.Status = model.StatusDriverNotFound
	}

	return repository.DB.Save(report).Error
}

func CreateDriverSubmission(req *DriverSubmissionRequest) (*model.DriverSubmission, error) {
	var order model.TaxiOrder
	if err := repository.DB.First(&order, req.OrderID).Error; err != nil {
		return nil, errors.New("订单不存在")
	}

	submission := &model.DriverSubmission{
		ReportID:        req.ReportID,
		OrderID:         req.OrderID,
		DriverID:        req.DriverID,
		DriverName:      req.DriverName,
		PlateNo:         req.PlateNo,
		ItemDescription: req.ItemDescription,
		ItemCategory:    req.ItemCategory,
		ItemValue:       req.ItemValue,
		IsValuable:      req.IsValuable,
		Photos:          req.Photos,
		FoundLocation:   req.FoundLocation,
		FoundTime:       req.FoundTime,
		SubmitTime:      time.Now(),
		Remark:          req.Remark,
	}

	if err := repository.DB.Create(submission).Error; err != nil {
		return nil, err
	}

	if req.ReportID > 0 {
		var report model.LostReport
		if err := repository.DB.First(&report, req.ReportID).Error; err == nil {
			report.Status = model.StatusSubmitted
			repository.DB.Save(&report)
		}
	}

	return submission, nil
}

func GetDriverSubmission(id int64) (*model.DriverSubmission, error) {
	var submission model.DriverSubmission
	if err := repository.DB.First(&submission, id).Error; err != nil {
		return nil, err
	}
	return &submission, nil
}

func ListDriverSubmissions(page, size int) ([]model.DriverSubmission, int64, error) {
	var submissions []model.DriverSubmission
	var total int64

	if err := repository.DB.Model(&model.DriverSubmission{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	if err := repository.DB.Order("created_at DESC").Offset(offset).Limit(size).Find(&submissions).Error; err != nil {
		return nil, 0, err
	}

	return submissions, total, nil
}

func CreateInventory(req *InventoryRequest) (*model.ItemInventory, error) {
	var submission model.DriverSubmission
	if err := repository.DB.First(&submission, req.SubmissionID).Error; err != nil {
		return nil, errors.New("上交记录不存在")
	}

	var station model.Station
	if err := repository.DB.First(&station, req.StationID).Error; err != nil {
		return nil, errors.New("站点不存在")
	}

	var existingInv model.ItemInventory
	if repository.DB.Where("submission_id = ?", req.SubmissionID).First(&existingInv).Error == nil {
		return nil, errors.New("该上交记录已入库")
	}

	inventory := &model.ItemInventory{
		SubmissionID:    req.SubmissionID,
		ReportID:        req.ReportID,
		ItemDescription: submission.ItemDescription,
		ItemCategory:    submission.ItemCategory,
		ItemValue:       submission.ItemValue,
		IsValuable:      submission.IsValuable,
		Photos:          submission.Photos,
		StationID:       req.StationID,
		StationName:     station.Name,
		CabinetNo:       req.CabinetNo,
		StoredBy:        req.StoredBy,
		StoredAt:        time.Now(),
		Status:          "in_stock",
	}

	if err := repository.DB.Create(inventory).Error; err != nil {
		return nil, err
	}

	return inventory, nil
}

func GetInventory(id int64) (*model.ItemInventory, error) {
	var inventory model.ItemInventory
	if err := repository.DB.First(&inventory, id).Error; err != nil {
		return nil, err
	}
	return &inventory, nil
}

func ListInventory(status string, page, size int) ([]model.ItemInventory, int64, error) {
	var items []model.ItemInventory
	var total int64

	query := repository.DB.Model(&model.ItemInventory{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	if err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func ListStations() ([]model.Station, error) {
	var stations []model.Station
	if err := repository.DB.Find(&stations).Error; err != nil {
		return nil, err
	}
	return stations, nil
}

func GetStation(id int64) (*model.Station, error) {
	var station model.Station
	if err := repository.DB.First(&station, id).Error; err != nil {
		return nil, err
	}
	return &station, nil
}

func CreateClaim(req *ClaimRequest) (*model.ClaimRecord, error) {
	inventory, err := GetInventory(req.InventoryID)
	if err != nil {
		return nil, errors.New("物品库存不存在")
	}
	if inventory.Status != "in_stock" {
		return nil, errors.New("该物品状态不是在库中，无法认领")
	}

	report, err := GetLostReport(req.ReportID)
	if err != nil {
		return nil, errors.New("报失记录不存在")
	}

	if inventory.IsValuable && req.PassengerIDCard == "" {
		return nil, errors.New("贵重物品认领必须提供身份证号")
	}

	if req.ReturnMethod == model.ReturnMethodExpress {
		if req.ReceiverName == "" || req.ReceiverPhone == "" || req.ReceiverAddress == "" {
			return nil, errors.New("快递寄回必须填写收件人姓名、电话和地址")
		}
	}
	if req.ReturnMethod == model.ReturnMethodPickup && req.PickupStationID == 0 {
		return nil, errors.New("到店领取必须选择领取站点")
	}

	var pendingClaims int64
	repository.DB.Model(&model.ClaimRecord{}).
		Where("inventory_id = ? AND status IN (?)", req.InventoryID, []model.ClaimStatus{model.ClaimStatusPending, model.ClaimStatusVerified}).
		Count(&pendingClaims)

	claim := &model.ClaimRecord{
		ReportID:          req.ReportID,
		InventoryID:       req.InventoryID,
		PassengerName:     req.PassengerName,
		PassengerPhone:    req.PassengerPhone,
		PassengerIDCard:   req.PassengerIDCard,
		VerifyMaterials:   req.VerifyMaterials,
		VerifyDescription: req.VerifyDescription,
		Status:            model.ClaimStatusPending,
		ReturnMethod:      req.ReturnMethod,
		ExpressNo:         req.ExpressNo,
		ExpressCompany:    req.ExpressCompany,
		ReceiverName:      req.ReceiverName,
		ReceiverPhone:     req.ReceiverPhone,
		ReceiverAddress:   req.ReceiverAddress,
		PickupStationID:   req.PickupStationID,
		Remark:            req.Remark,
	}

	if req.PickupStationID > 0 {
		station, err := GetStation(req.PickupStationID)
		if err == nil {
			claim.PickupStationName = station.Name
		}
	}

	if err := repository.DB.Create(claim).Error; err != nil {
		return nil, err
	}

	if pendingClaims > 0 {
		report.Status = model.StatusMultipleClaims
		repository.DB.Save(report)
	}

	return claim, nil
}

func GetClaim(id int64) (*model.ClaimRecord, error) {
	var claim model.ClaimRecord
	if err := repository.DB.First(&claim, id).Error; err != nil {
		return nil, err
	}
	return &claim, nil
}

func ListClaims(status string, reportID int64, page, size int) ([]model.ClaimRecord, int64, error) {
	var claims []model.ClaimRecord
	var total int64

	query := repository.DB.Model(&model.ClaimRecord{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if reportID > 0 {
		query = query.Where("report_id = ?", reportID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	if err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&claims).Error; err != nil {
		return nil, 0, err
	}

	return claims, total, nil
}

func VerifyClaim(claimID int64, req *VerifyClaimRequest) (*model.ClaimRecord, error) {
	claim, err := GetClaim(claimID)
	if err != nil {
		return nil, err
	}
	if claim.Status != model.ClaimStatusPending {
		return nil, errors.New("该认领申请状态不是待核验")
	}

	now := time.Now()
	claim.VerifiedBy = req.VerifiedBy
	claim.VerifiedAt = &now

	if req.Pass {
		claim.Status = model.ClaimStatusVerified

		var report model.LostReport
		if repository.DB.First(&report, claim.ReportID).Error == nil {
			report.Status = model.StatusVerified
			repository.DB.Save(&report)
		}

		var otherClaims []model.ClaimRecord
		repository.DB.Where("inventory_id = ? AND id != ? AND status = ?", claim.InventoryID, claimID, model.ClaimStatusPending).
			Find(&otherClaims)
		for i := range otherClaims {
			otherClaims[i].Status = model.ClaimStatusRejected
			otherClaims[i].Remark = fmt.Sprintf("已有其他认领通过核验: %s", req.RejectReason)
			repository.DB.Save(&otherClaims[i])
		}
	} else {
		claim.Status = model.ClaimStatusRejected
		if req.RejectReason != "" {
			if claim.Remark != "" {
				claim.Remark += "; "
			}
			claim.Remark += "拒绝原因: " + req.RejectReason
		}
	}

	if err := repository.DB.Save(claim).Error; err != nil {
		return nil, err
	}

	return claim, nil
}

func ConfirmReturn(claimID int64, req *ReturnConfirmRequest) (*model.ClaimRecord, error) {
	claim, err := GetClaim(claimID)
	if err != nil {
		return nil, err
	}
	if claim.Status != model.ClaimStatusVerified {
		return nil, errors.New("该认领申请状态不是核验通过，无法归还")
	}

	returnedAt := req.ReturnedAt
	if returnedAt.IsZero() {
		returnedAt = time.Now()
	}
	claim.ReturnedAt = &returnedAt
	claim.Status = model.ClaimStatusReturned

	if claim.ReturnMethod == model.ReturnMethodPickup {
		claim.PickupTime = &returnedAt
		if claim.PickupPerson == "" {
			claim.PickupPerson = req.Operator
		}
	}

	if err := repository.DB.Save(claim).Error; err != nil {
		return nil, err
	}

	var inventory model.ItemInventory
	if repository.DB.First(&inventory, claim.InventoryID).Error == nil {
		inventory.Status = "returned"
		repository.DB.Save(&inventory)
	}

	var report model.LostReport
	if repository.DB.First(&report, claim.ReportID).Error == nil {
		report.Status = model.StatusReturned
		repository.DB.Save(&report)
	}

	return claim, nil
}

func GetOrder(id int64) (*model.TaxiOrder, error) {
	var order model.TaxiOrder
	if err := repository.DB.First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func ListOrders(page, size int) ([]model.TaxiOrder, int64, error) {
	var orders []model.TaxiOrder
	var total int64

	if err := repository.DB.Model(&model.TaxiOrder{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	if err := repository.DB.Order("ride_start_time DESC").Offset(offset).Limit(size).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
