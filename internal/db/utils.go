package db

func GetConnectionString(user, password, host, port, dbName string, dbType string) string {
	return dbType + "://" + user + ":" + password + "@" + host + ":" + port + "/" + dbName + "?sslmode=disable"
}
