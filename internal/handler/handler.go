package handler

import (
	"errors"
	"net/http"
	"strconv"
	"taxi-lost-property/internal/service"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: data})
}

func fail(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, Response{Code: 1, Message: err.Error()})
}

func getPageSize(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	return page, size
}

func getID(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func CreateLostReport(c *gin.Context) {
	var req service.LostReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, err)
		return
	}
	report, err := service.CreateLostReport(&req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, report)
}

func GetLostReport(c *gin.Context) {
	id, err := getID(c)
	if err != nil {
		fail(c, err)
		return
	}
	report, err := service.GetLostReport(id)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, report)
}

func ListLostReports(c *gin.Context) {
	status := c.Query("status")
	page, size := getPageSize(c)
	reports, total, err := service.ListLostReports(status, page, size)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, gin.H{"list": reports, "total": total, "page": page, "size": size})
}

func MatchOrders(c *gin.Context) {
	id, err := getID(c)
	if err != nil {
		fail(c, err)
		return
	}
	results, err := service.MatchOrders(id)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, results)
}

func ConfirmMatchedOrder(c *gin.Context) {
	reportID, err := getID(c)
	if err != nil {
		fail(c, err)
		return
	}
	var body struct {
		OrderID int64 `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, err)
		return
	}
	if err := service.ConfirmMatchedOrder(reportID, body.OrderID); err != nil {
		fail(c, err)
		return
	}
	ok(c, nil)
}

func CreateDriverSubmission(c *gin.Context) {
	var req service.DriverSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, err)
		return
	}
	submission, err := service.CreateDriverSubmission(&req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, submission)
}

func GetDriverSubmission(c *gin.Context) {
	id, err := getID(c)
	if err != nil {
		fail(c, err)
		return
	}
	submission, err := service.GetDriverSubmission(id)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, submission)
}

func ListDriverSubmissions(c *gin.Context) {
	page, size := getPageSize(c)
	submissions, total, err := service.ListDriverSubmissions(page, size)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, gin.H{"list": submissions, "total": total, "page": page, "size": size})
}

func CreateInventory(c *gin.Context) {
	var req service.InventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, err)
		return
	}
	inventory, err := service.CreateInventory(&req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, inventory)
}

func GetInventory(c *gin.Context) {
	id, err := getID(c)
	if err != nil {
		fail(c, err)
		return
	}
	inventory, err := service.GetInventory(id)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, inventory)
}

func ListInventory(c *gin.Context) {
	status := c.Query("status")
	page, size := getPageSize(c)
	items, total, err := service.ListInventory(status, page, size)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, gin.H{"list": items, "total": total, "page": page, "size": size})
}

func ListStations(c *gin.Context) {
	stations, err := service.ListStations()
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, stations)
}

func GetStation(c *gin.Context) {
	id, err := getID(c)
	if err != nil {
		fail(c, err)
		return
	}
	station, err := service.GetStation(id)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, station)
}

func CreateClaim(c *gin.Context) {
	var req service.ClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, err)
		return
	}
	claim, err := service.CreateClaim(&req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, claim)
}

func GetClaim(c *gin.Context) {
	id, err := getID(c)
	if err != nil {
		fail(c, err)
		return
	}
	claim, err := service.GetClaim(id)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, claim)
}

func ListClaims(c *gin.Context) {
	status := c.Query("status")
	reportID, _ := strconv.ParseInt(c.Query("report_id"), 10, 64)
	page, size := getPageSize(c)
	claims, total, err := service.ListClaims(status, reportID, page, size)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, gin.H{"list": claims, "total": total, "page": page, "size": size})
}

func VerifyClaim(c *gin.Context) {
	id, err := getID(c)
	if err != nil {
		fail(c, err)
		return
	}
	var req service.VerifyClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, err)
		return
	}
	claim, err := service.VerifyClaim(id, &req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, claim)
}

func ConfirmReturn(c *gin.Context) {
	id, err := getID(c)
	if err != nil {
		fail(c, err)
		return
	}
	var req service.ReturnConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, err)
		return
	}
	claim, err := service.ConfirmReturn(id, &req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, claim)
}

func GetOrder(c *gin.Context) {
	id, err := getID(c)
	if err != nil {
		fail(c, err)
		return
	}
	order, err := service.GetOrder(id)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, order)
}

func ListOrders(c *gin.Context) {
	page, size := getPageSize(c)
	orders, total, err := service.ListOrders(page, size)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, gin.H{"list": orders, "total": total, "page": page, "size": size})
}

func FuzzySearchVehicles(c *gin.Context) {
	var req service.FuzzySearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, err)
		return
	}
	candidates, err := service.FuzzySearchVehicles(&req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, gin.H{"list": candidates, "total": len(candidates)})
}

func GetVerifyRequirements(c *gin.Context) {
	inventoryID, err := strconv.ParseInt(c.Query("inventory_id"), 10, 64)
	if err != nil {
		fail(c, errors.New("请提供有效的 inventory_id"))
		return
	}
	inventory, err := service.GetInventory(inventoryID)
	if err != nil {
		fail(c, errors.New("物品库存不存在"))
		return
	}
	itemType := service.DetectValuableItemType(inventory.ItemDescription, inventory.ItemCategory)
	requirements := service.GetVerifyRequirements(itemType)
	ok(c, gin.H{
		"item_type":     itemType,
		"is_valuable":   inventory.IsValuable,
		"requirements":  requirements,
	})
}
