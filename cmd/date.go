/*
Package cmd
Copyright © 2025 Chuang Libs <chan.toddd@gmail.com>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/abema/go-mp4"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/spf13/cobra"
)

var dir string
var verbose bool

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
	dateCmd.Flags().StringVarP(&dir, "dir", "d", ".", "指定要处理的目录，默认是当前目录")
	dateCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "启用详细输出，如果不启用详细输出则默认会启用文件日志，除非手动关闭了文件日志输出")
}

// processPhotos 处理目录中的照片
func processPhotos() error {
	// 获取绝对路径
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("获取目录绝对路径失败: %v", err)
	}

	// 输出开始信息
	fmt.Printf("开始处理目录: %s\n", absDir)
	if verbose {
		fmt.Printf("处理时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Printf("=====================================\n")
	}

	// 创建处理日志文件
	var logFile *os.File
	var logWriter *os.File
	logFileName := time.Now().Format("2006-01-02_15_04_05") + ".log"
	logFilePath := filepath.Join(absDir, logFileName)
	logFile, err = os.Create(logFilePath)
	if err != nil {
		log.Fatalf("无法创建日志文件 %s: %v", logFilePath, err)
	} else {
		defer logFile.Close()
		logWriter = logFile
		if verbose {
			fmt.Printf("日志文件已创建: %s\n", logFilePath)
		}
	}

	fmt.Fprintf(logWriter, "开始处理目录: %s\n", absDir)
	fmt.Fprintf(logWriter, "处理时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(logWriter, "=====================================\n")

	// 支持的图片格式
	supportedFormats := []string{".jpg", ".jpeg", ".png", ".tiff", ".tif", ".heic", ".mp4", ".mov", ".avi"}

	// 计数器
	totalFiles := 0
	movedFiles := 0
	skippedFiles := 0

	outputText := ""

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

		// 跳过已经在年月目录中的文件
		if isAlreadyInYearMonthDir(path, absDir) {
			outputText = fmt.Sprintf("跳过: %s 已在年月目录中\n", path)
			if verbose {
				fmt.Println(outputText)
			}
			fmt.Fprintln(logWriter, outputText)
			return nil
		}

		totalFiles++

		// 获取照片时间信息
		photoTime, timeSource, err := getPhotoTimeInfo(path)
		if err != nil {
			outputText = fmt.Sprintf("警告: 无法获取 %s 的时间信息: %v\n", path, err)
			if verbose {
				fmt.Println(outputText)
			}
			fmt.Fprintln(logWriter, outputText)
			return nil
		}

		// 创建年月目录 (格式: 2025-01)
		yearMonth := photoTime.Format("2006-01")
		targetDir := filepath.Join(absDir, yearMonth)

		// 创建目录（如果不存在）
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			outputText = fmt.Sprintf("错误: 无法创建目录 %s: %v\n", targetDir, err)
			if verbose {
				fmt.Println(outputText)
			}
			fmt.Fprintln(logWriter, outputText)
			return nil
		}

		// 生成新的文件名（自动处理同一秒的文件）
		newFileName := generateNewFileName(targetDir, info.Name(), photoTime, ext)

		// 如果文件名无需改动，直接移动到目标目录
		if newFileName == info.Name() {
			targetPath := filepath.Join(targetDir, newFileName)
			if err := moveFile(path, targetPath); err != nil {
				outputText = fmt.Sprintf("错误: 无法移动文件 %s 到 %s: %v\n", path, targetPath, err)
				if verbose {
					fmt.Println(outputText)
				}

				fmt.Fprintln(logWriter, outputText)
				return nil
			}
			movedFiles++

			// 显示处理信息
			var message string
			switch timeSource {
			case "EXIF":
				message = fmt.Sprintf("[EXIF] %s -> %s: %s (无需重命名)", path, targetPath, photoTime.Format("2006-01-02 15:04:05"))
			case "VIDEO_META":
				message = fmt.Sprintf("[视频元数据] %s -> %s: %s (无需重命名)", path, targetPath, photoTime.Format("2006-01-02 15:04:05"))
			default:
				message = fmt.Sprintf("[文件时间] %s -> %s: %s (无需重命名)", path, targetPath, photoTime.Format("2006-01-02 15:04:05"))
			}

			if verbose {
				fmt.Println(message)
			}

			fmt.Fprintln(logWriter, message)
			return nil
		}

		// 需要重命名的情况
		targetPath := filepath.Join(targetDir, newFileName)

		// 检查目标文件是否已存在
		if _, err := os.Stat(targetPath); err == nil {
			outputText = fmt.Sprintf("跳过: 目标文件 %s 已存在\n", targetPath)
			if verbose {
				fmt.Println(outputText)
			}

			fmt.Fprintln(logWriter, outputText)
			skippedFiles++
			return nil
		}

		// 移动文件
		if err := moveFile(path, targetPath); err != nil {
			outputText = fmt.Sprintf("错误: 无法移动文件 %s 到 %s: %v\n", path, targetPath, err)
			if verbose {
				fmt.Println(outputText)
			}

			fmt.Fprintln(logWriter, outputText)
			return nil
		}

		movedFiles++

		// 显示处理信息
		var message string
		switch timeSource {
		case "EXIF":
			message = fmt.Sprintf("[EXIF] %s -> %s: %s", path, targetPath, photoTime.Format("2006-01-02 15:04:05"))
		case "VIDEO_META":
			message = fmt.Sprintf("[视频元数据] %s -> %s: %s", path, targetPath, photoTime.Format("2006-01-02 15:04:05"))
		default:
			message = fmt.Sprintf("[文件时间] %s -> %s: %s", path, targetPath, photoTime.Format("2006-01-02 15:04:05"))
		}

		if verbose {
			fmt.Println(message)
		}

		fmt.Fprintln(logWriter, message)

		return nil
	})

	if err != nil {
		return fmt.Errorf("遍历目录失败: %v", err)
	}

	// 输出统计信息
	fmt.Printf("\n处理完成！\n")
	fmt.Printf("总计找到 %d 个照片/视频文件\n", totalFiles)
	fmt.Printf("成功移动: %d 个\n", movedFiles)
	fmt.Printf("跳过: %d 个\n", skippedFiles)
	fmt.Printf("=====================================\n")
	fmt.Printf("处理结束时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	fmt.Fprintf(logWriter, "\n处理完成！\n")
	fmt.Fprintf(logWriter, "总计找到 %d 个照片/视频文件\n", totalFiles)
	fmt.Fprintf(logWriter, "成功移动: %d 个\n", movedFiles)
	fmt.Fprintf(logWriter, "跳过: %d 个\n", skippedFiles)
	fmt.Fprintf(logWriter, "=====================================\n")
	fmt.Fprintf(logWriter, "处理结束时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))

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
			return videoTime, "VIDEO_META", nil
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

	return info.ModTime(), "FILE_TIME", nil
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

// generateNewFileName 生成新的文件名，处理同一秒的文件，并自动重命名已存在的文件
func generateNewFileName(targetDir string, originalName string, photoTime time.Time, ext string) string {
	// 检查是否已经是标准格式
	if isStandardFileName(originalName, ext) {
		return originalName
	}

	// 根据文件类型选择前缀
	prefix := "IMG"
	if isVideoFile(ext) {
		prefix = "VID"
	}

	// 基础文件名: IMG_年月_时分秒
	baseName := fmt.Sprintf("%s_%s", prefix, photoTime.Format("20060102_150405"))

	// 首先检查基础文件名是否存在
	baseFileName := fmt.Sprintf("%s%s", baseName, ext)
	baseTargetPath := filepath.Join(targetDir, baseFileName)

	// 如果基础文件名不存在，使用它
	if _, err := os.Stat(baseTargetPath); os.IsNotExist(err) {
		return baseFileName
	}

	// 基础文件名已存在，需要重命名已存在的文件为_001，然后继续顺序编号
	// 首先重命名已存在的文件
	if err := renameExistingFileToFirstSequential(targetDir, baseFileName, baseName, ext); err != nil {
		fmt.Printf("警告: 无法重命名已存在的文件 %s: %v\n", baseTargetPath, err)
	}

	// 现在从_001开始为当前文件寻找可用的文件名
	for i := 1; i <= 999; i++ {
		newName := fmt.Sprintf("%s_%03d%s", baseName, i, ext)
		targetPath := filepath.Join(targetDir, newName)

		// 检查文件是否存在
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			return newName
		}
	}

	// 如果所有序号都用完了，返回带时间戳的文件名
	return fmt.Sprintf("%s_%d%s", baseName, time.Now().UnixNano()%1000, ext)
}

// isStandardFileName 检查是否已经是标准文件名格式（支持带序号的格式）
func isStandardFileName(name string, ext string) bool {
	// 移除扩展名
	nameWithoutExt := strings.TrimSuffix(name, ext)

	// 检查图片格式: IMG_年月_时分秒 或 IMG_年月_时分秒_序号
	if strings.HasPrefix(nameWithoutExt, "IMG_") {
		// 基础格式: IMG_年月_时分秒 (15字符)
		if len(nameWithoutExt) == 15 {
			return true
		}
		// 带序号格式: IMG_年月_时分秒_001 (19字符)
		if len(nameWithoutExt) == 19 && strings.Count(nameWithoutExt, "_") == 3 {
			// 检查最后一部分是否为数字
			parts := strings.Split(nameWithoutExt, "_")
			if len(parts) == 4 {
				if _, err := strconv.Atoi(parts[3]); err == nil {
					return true
				}
			}
		}
	}

	// 检查视频格式: VID_年月_时分秒 或 VID_年月_时分秒_序号
	if strings.HasPrefix(nameWithoutExt, "VID_") {
		// 基础格式: VID_年月_时分秒 (15字符)
		if len(nameWithoutExt) == 15 {
			return true
		}
		// 带序号格式: VID_年月_时分秒_001 (19字符)
		if len(nameWithoutExt) == 19 && strings.Count(nameWithoutExt, "_") == 3 {
			// 检查最后一部分是否为数字
			parts := strings.Split(nameWithoutExt, "_")
			if len(parts) == 4 {
				if _, err := strconv.Atoi(parts[3]); err == nil {
					return true
				}
			}
		}
	}

	return false
}

// isAlreadyInYearMonthDir 检查文件是否已经在年月目录中
func isAlreadyInYearMonthDir(filePath string, baseDir string) bool {
	relPath, err := filepath.Rel(baseDir, filePath)
	if err != nil {
		return false
	}

	// 分割路径
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) < 2 {
		return false
	}

	// 检查第一级目录是否符合年月格式 (2025-01)
	firstDir := parts[0]
	if len(firstDir) != 7 || firstDir[4] != '-' {
		return false
	}

	// 检查是否为有效的年月
	_, err = time.Parse("2006-01", firstDir)
	return err == nil
}

// moveFile 移动文件并保持时间戳
func moveFile(src, dst string) error {
	// 获取源文件的信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 复制文件内容
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// 写入目标文件
	err = os.WriteFile(dst, input, srcInfo.Mode())
	if err != nil {
		return err
	}

	// 保持原始时间戳
	err = os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
	if err != nil {
		return err
	}

	// 删除源文件
	return os.Remove(src)
}

// renameExistingFileToFirstSequential 将已存在的基础文件重命名为第一个序号文件
func renameExistingFileToFirstSequential(targetDir string, oldFileName string, baseName string, ext string) error {
	oldPath := filepath.Join(targetDir, oldFileName)
	newPath := filepath.Join(targetDir, fmt.Sprintf("%s_001%s", baseName, ext))

	// 检查目标文件是否已存在
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("目标文件 %s 已存在", newPath)
	}

	// 重命名文件
	return os.Rename(oldPath, newPath)
}
