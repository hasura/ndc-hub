package cmd

func getFilePathRelativeToRepository(filePath string) string {
	return "../../" + filePath
}
