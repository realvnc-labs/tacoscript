package utils

func GetBackupFilename(filename string, ext string) (backupFilename string) {
	return filename + "." + ext
}
