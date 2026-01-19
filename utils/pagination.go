// Package utils 提供通用工具函数
package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PageRequest 分页请求参数
type PageRequest struct {
	Page     int    `form:"page" json:"page"`           // 页码，从1开始
	PageSize int    `form:"page_size" json:"page_size"` // 每页数量
	Sort     string `form:"sort" json:"sort"`           // 排序字段
	Order    string `form:"order" json:"order"`         // 排序方向: asc/desc
}

// PageResponse 分页响应
type PageResponse struct {
	List     interface{} `json:"list"`      // 数据列表
	Total    int64       `json:"total"`     // 总记录数
	Page     int         `json:"page"`      // 当前页码
	PageSize int         `json:"page_size"` // 每页数量
	Pages    int         `json:"pages"`     // 总页数
}

// 分页默认值
const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// NewPageRequest 从Gin上下文创建分页请求
func NewPageRequest(c *gin.Context) *PageRequest {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	sort := c.DefaultQuery("sort", "id")
	order := c.DefaultQuery("order", "desc")

	return &PageRequest{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Order:    order,
	}
}

// Normalize 规范化分页参数
func (p *PageRequest) Normalize() {
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.PageSize < 1 {
		p.PageSize = DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		p.PageSize = MaxPageSize
	}
	if p.Order != "asc" && p.Order != "desc" {
		p.Order = "desc"
	}
}

// Offset 计算偏移量
func (p *PageRequest) Offset() int {
	p.Normalize()
	return (p.Page - 1) * p.PageSize
}

// Limit 获取限制数量
func (p *PageRequest) Limit() int {
	p.Normalize()
	return p.PageSize
}

// OrderClause 获取排序子句
func (p *PageRequest) OrderClause() string {
	p.Normalize()
	if p.Sort == "" {
		p.Sort = "id"
	}
	return p.Sort + " " + p.Order
}

// Paginate GORM分页作用域
func Paginate(req *PageRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		req.Normalize()
		return db.Offset(req.Offset()).Limit(req.Limit())
	}
}

// PaginateWithOrder GORM分页+排序作用域
func PaginateWithOrder(req *PageRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		req.Normalize()
		return db.Offset(req.Offset()).Limit(req.Limit()).Order(req.OrderClause())
	}
}

// NewPageResponse 创建分页响应
func NewPageResponse(list interface{}, total int64, req *PageRequest) *PageResponse {
	req.Normalize()
	pages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		pages++
	}

	return &PageResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Pages:    pages,
	}
}

// PaginateQuery 执行分页查询
func PaginateQuery(db *gorm.DB, req *PageRequest, dest interface{}) (*PageResponse, error) {
	req.Normalize()

	var total int64

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 获取分页数据
	if err := db.Scopes(PaginateWithOrder(req)).Find(dest).Error; err != nil {
		return nil, err
	}

	return NewPageResponse(dest, total, req), nil
}

// SimplePaginate 简化分页（不返回总数，性能更好）
func SimplePaginate(db *gorm.DB, page, pageSize int, dest interface{}) error {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	offset := (page - 1) * pageSize
	return db.Offset(offset).Limit(pageSize).Find(dest).Error
}
