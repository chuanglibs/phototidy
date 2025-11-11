/*
Package cmd
Copyright © 2025 Chuang Libs <chan.toddd@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/abema/go-mp4"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/spf13/cobra"
)

var dir string

// dateCmd represents the date command
var dateCmd = &cobra.Command{
	Use:   "date",
	Short: "把照片、视频按日期来进行目录分类",
	Long: `读取照片、视频文件的日期来进行识别
以其拍摄日期为准来进行目录分类，按年份-》月份来进行目录创建，并把图片、视频文件移动到对应的目录中
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := processPhotos(); err != nil {
			fmt.Printf("处理照片时出错: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(dateCmd)
	dateCmd.Flags().StringVarP(&dir, "dir", "d", ".", "指定要处理的目录")
}

// processPhotos 处理目录中的照片
func processPhotos() error {
	// 获取绝对路径
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("获取目录绝对路径失败: %v", err)
	}

	fmt.Printf("开始处理目录: %s\n", absDir)

	// 支持的图片格式
	supportedFormats := []string{".jpg", ".jpeg", ".png", ".tiff", ".tif", ".heic", ".mp4", ".mov", ".avi"}

	// 计数器
	totalFiles := 0
	exifFiles := 0
	videoFiles := 0
	modTimeFiles := 0

	// 遍历目录
	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查文件扩展名
		ext := strings.ToLower(filepath.Ext(path))
		isSupported := false
		for _, format := range supportedFormats {
			if ext == format {
				isSupported = true
				break
			}
		}

		if !isSupported {
			return nil
		}

		totalFiles++

		// 获取照片时间信息
		photoTime, timeSource, err := getPhotoTimeInfo(path)
		if err != nil {
			fmt.Printf("警告: 无法获取 %s 的时间信息: %v\n", path, err)
			return nil
		}

		// 根据时间来源进行区分显示
		switch timeSource {
		case "EXIF":
			exifFiles++
			fmt.Printf("[EXIF] %s: %s\n", path, photoTime.Format("2006-01-02 15:04:05"))
		case "视频元数据":
			videoFiles++
			fmt.Printf("[视频元数据] %s: %s\n", path, photoTime.Format("2006-01-02 15:04:05"))
		default:
			modTimeFiles++
			fmt.Printf("[文件时间] %s: %s\n", path, photoTime.Format("2006-01-02 15:04:05"))
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("遍历目录失败: %v", err)
	}

	fmt.Printf("\n处理完成！\n")
	fmt.Printf("总计找到 %d 个照片/视频文件\n", totalFiles)
	fmt.Printf("从EXIF获取时间: %d 个\n", exifFiles)
	fmt.Printf("从视频元数据获取时间: %d 个\n", videoFiles)
	fmt.Printf("从文件时间获取: %d 个\n", modTimeFiles)

	return nil
}

// getPhotoTimeInfo 获取照片时间信息和来源
func getPhotoTimeInfo(filePath string) (time.Time, string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	// 根据文件类型选择不同的处理方式
	if isVideoFile(ext) {
		// 视频文件：尝试从MP4元数据获取创建时间
		videoTime, err := getVideoCreationTime(filePath)
		if err == nil && !videoTime.IsZero() {
			return videoTime, "视频元数据", nil
		}
	} else {
		// 图片文件：首先尝试从EXIF数据获取拍摄日期
		exifTime, err := getExifDateTime(filePath)
		if err == nil && !exifTime.IsZero() {
			return exifTime, "EXIF", nil
		}
	}

	// 如果专用元数据不可用，使用文件修改时间
	info, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("无法获取文件信息: %v", err)
	}

	return info.ModTime(), "文件时间", nil
}

// isVideoFile 判断是否为视频文件
func isVideoFile(ext string) bool {
	videoExts := []string{".mp4", ".mov", ".avi", ".mkv", ".flv", ".wmv"}
	for _, videoExt := range videoExts {
		if ext == videoExt {
			return true
		}
	}
	return false
}

// getExifDateTime 从EXIF数据获取拍摄日期时间
func getExifDateTime(filePath string) (time.Time, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return time.Time{}, err
	}
	defer file.Close()

	// 解码EXIF数据
	x, err := exif.Decode(file)
	if err != nil {
		return time.Time{}, err
	}

	// 获取拍摄日期
	date, err := x.DateTime()
	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}

// getVideoCreationTime 从视频元数据获取创建时间
func getVideoCreationTime(filePath string) (time.Time, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return time.Time{}, err
	}
	defer file.Close()

	// 使用go-mp4库读取视频元数据
	var creationTime time.Time

	// 创建MP4读取器
	_, err = mp4.ReadBoxStructure(file, func(h *mp4.ReadHandle) (interface{}, error) {
		switch h.BoxInfo.Type.String() {
		case "mvhd":
			// Movie header box
			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			if mvhd, ok := box.(*mp4.Mvhd); ok {
				creationTimeInt64 := mvhd.GetCreationTime()
				creationTime = time.Unix(int64(creationTimeInt64), 0)
			}
		case "mdhd":
			// Media header box
			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			if mdhd, ok := box.(*mp4.Mdhd); ok {
				if creationTime.IsZero() {
					creationTimeInt64 := mdhd.GetCreationTime()
					creationTime = time.Unix(int64(creationTimeInt64), 0)
				}
			}
		}
		return nil, nil
	})

	if err != nil {
		return time.Time{}, err
	}

	if creationTime.IsZero() {
		return time.Time{}, fmt.Errorf("未找到视频创建时间")
	}

	return creationTime, nil
}
