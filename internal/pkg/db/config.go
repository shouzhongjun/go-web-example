package db

type DBConfig struct {
	DSN string
}

func ProvideDBConfig() *DBConfig {
	return &DBConfig{
		// mysql 连接字符串
		DSN: "sonic:n2l5_H81m@tcp(rm-bp139wfyibp65u7522o.mysql.rds.aliyuncs.com:3306)/my_database?charset=utf8mb4&parseTime=True&loc=Local",
	}
}
