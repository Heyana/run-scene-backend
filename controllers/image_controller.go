package controllers

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"net/http"

	"github.com/gin-gonic/gin"
	"go_wails_project_manager/response"
)

// ImageController 图片处理控制器
type ImageController struct{}

// NewImageController 创建图片控制器
func NewImageController() *ImageController {
	return &ImageController{}
}

// FlipYAndToWebp 图片翻转并转换为WebP
// @Summary 图片翻转并转换为WebP
// @Tags 图片处理
// @Accept multipart/form-data
// @Produce json
// @Param flip query bool false "是否翻转Y轴" default(false)
// @Param files formData file true "图片文件（可多个）"
// @Success 200 {object} response.Response
// @Router /api/image/flipy-webp [post]
func (c *ImageController) FlipYAndToWebp(ctx *gin.Context) {
	// 获取翻转参数
	flip := ctx.DefaultQuery("flip", "false") == "true"

	// 获取上传的文件
	form, err := ctx.MultipartForm()
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "获取文件失败")
		return
	}

	files := form.File
	if len(files) == 0 {
		response.Error(ctx, http.StatusBadRequest, "未找到上传文件")
		return
	}

	result := make(map[string]interface{})

	// 处理每个文件
	for fieldName, fileHeaders := range files {
		if len(fileHeaders) == 0 {
			continue
		}

		fileHeader := fileHeaders[0]

		// 打开文件
		file, err := fileHeader.Open()
		if err != nil {
			result[fieldName] = map[string]interface{}{
				"error": fmt.Sprintf("打开文件失败: %v", err),
			}
			continue
		}
		defer file.Close()

		// 解码图片
		img, format, err := image.Decode(file)
		if err != nil {
			result[fieldName] = map[string]interface{}{
				"error": fmt.Sprintf("解码图片失败: %v", err),
			}
			continue
		}

		// 如果需要翻转Y轴
		if flip {
			img = flipImageY(img)
		}

		// 转换为PNG（暂时不支持WebP，因为需要CGO）
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			result[fieldName] = map[string]interface{}{
				"error": fmt.Sprintf("编码图片失败: %v", err),
			}
			continue
		}

		// 返回字节数组
		result[fieldName] = map[string]interface{}{
			"data":   buf.Bytes(),
			"format": format,
			"width":  img.Bounds().Dx(),
			"height": img.Bounds().Dy(),
		}
	}

	response.Success(ctx, result)
}

// flipImageY 翻转图片Y轴
func flipImageY(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	flipped := image.NewRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			flipped.Set(x, height-1-y, img.At(x, y))
		}
	}

	return flipped
}
