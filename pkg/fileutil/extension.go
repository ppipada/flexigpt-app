package fileutil

type ExtensionMode string

const (
	ExtensionModeText     ExtensionMode = "text"
	ExtensionModeImage    ExtensionMode = "image"
	ExtensionModeDocument ExtensionMode = "document"
	ExtensionModeDefault  ExtensionMode = "default"
)

const ExtensionPDF = ".pdf"

var ExtensionToMode = map[string]ExtensionMode{
	".txt":      ExtensionModeText,
	".md":       ExtensionModeText,
	".markdown": ExtensionModeText,
	".log":      ExtensionModeText,
	".json":     ExtensionModeText,
	".yaml":     ExtensionModeText,
	".yml":      ExtensionModeText,
	".toml":     ExtensionModeText,
	".js":       ExtensionModeText,
	".ts":       ExtensionModeText,
	".tsx":      ExtensionModeText,
	".jsx":      ExtensionModeText,
	".py":       ExtensionModeText,
	".go":       ExtensionModeText,
	".rs":       ExtensionModeText,
	".java":     ExtensionModeText,
	".c":        ExtensionModeText,
	".cpp":      ExtensionModeText,
	".h":        ExtensionModeText,
	".hpp":      ExtensionModeText,
	".cs":       ExtensionModeText,
	".rb":       ExtensionModeText,
	".php":      ExtensionModeText,
	".html":     ExtensionModeText,
	".css":      ExtensionModeText,
	".scss":     ExtensionModeText,
	".less":     ExtensionModeText,
	".sql":      ExtensionModeText,

	".jpg":  ExtensionModeImage,
	".jpeg": ExtensionModeImage,
	".png":  ExtensionModeImage,
	".gif":  ExtensionModeImage,
	".webp": ExtensionModeImage,
	".bmp":  ExtensionModeImage,
	".svg":  ExtensionModeImage,

	ExtensionPDF: ExtensionModeDocument,
	".doc":       ExtensionModeDocument,
	".docx":      ExtensionModeDocument,
	".ppt":       ExtensionModeDocument,
	".pptx":      ExtensionModeDocument,
	".xls":       ExtensionModeDocument,
	".xlsx":      ExtensionModeDocument,
	".odt":       ExtensionModeDocument,
	".ods":       ExtensionModeDocument,
}
