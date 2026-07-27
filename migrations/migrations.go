package migrations

import "embed"

//go:embed *_table.*.sql
var fs embed.FS

func GetMigration(fileName string) (string, error) {
	fileBytes, err := fs.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	// Convert the raw byte slice to a string to see the SQL code
	sqlContent := string(fileBytes)
	
	return sqlContent, nil
}