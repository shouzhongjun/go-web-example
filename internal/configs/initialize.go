package configs

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// ConfigPath 配置文件路径
var ConfigPath string

func init() {
	flag.StringVar(&ConfigPath, "conf", "configs/config.dev.yaml", "配置文件路径, 例: -conf ./configs/config.dev.yaml")
}

// AllConfig 应用全局配置
type AllConfig struct {
	Server   Server      `yaml:"server"`
	Log      Log         `yaml:"log"`
	Database Database    `yaml:"database"`
	Redis    Redis       `yaml:"redis"`
	Etcd     *Etcd       `yaml:"etcd"`
	Kafka    KafkaConfig `yaml:"kafka"`
	MongoDB  *MongoDB    `yaml:"mongodb"`
}

// Log 日志配置
type Log struct {
	Level         string `yaml:"level"`
	EnableFile    bool   `yaml:"enableFile"`
	EnableConsole bool   `yaml:"enableConsole"`
	Prefix        string `yaml:"prefix"`
	Path          string `yaml:"path"`
}

// Database 数据库配置
type Database struct {
	SSLMode         string `yaml:"ssl_mode"`
	MaxOpenConns    int    `yaml:"maxOpen_conns"`
	ConnMaxLifetime *int64 `yaml:"connMaxLifetime"`
	Host            string `yaml:"host"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Port            int    `yaml:"port"`
	DBName          string `yaml:"dbname"`
	MaxIdleConns    int    `yaml:"maxIdleConns"`
	LogLevel        string `yaml:"logLevel"`
	Trace           bool   `yaml:"trace"`
	ConnMaxIdleTime *int64 `yaml:"connMaxIdleTime"`
}

// DSN 获取数据库连接字符串
func (db *Database) DSN() string {
	// mysql dsn
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		db.User, db.Password, db.Host, db.Port, db.DBName)

	//return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
	//	db.Host, db.User, db.Password, db.DBName, db.Port, db.SSLMode)
}

// Redis 缓存配置
type Redis struct {
	MaxActiveConns int    `yaml:"maxActiveConns"`
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Password       string `yaml:"password"`
	Db             int    `yaml:"db"`
	MaxIdleConns   int    `yaml:"maxIdleConns"`
	Enable         bool   `yaml:"enable"`
}

// RedisAddr 获取Redis地址
func (r *Redis) RedisAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// Etcd 配置
type Etcd struct {
	Host        string `yaml:"host"`
	Port        int64  `yaml:"port"`
	DialTimeOut int64  `yaml:"dialTimeOut"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	LeaseTTL    int64  `yaml:"leaseTTL"`
	Enable      bool   `yaml:"enable"`
}

// EtcdAddr 获取Etcd地址
func (e *Etcd) EtcdAddr() string {
	return fmt.Sprintf("%s:%d", e.Host, e.Port)
}

func (e *Etcd) DialTimeout() time.Duration {
	return time.Duration(e.DialTimeOut) * time.Second
}

// GetLeaseTTL 获取租约TTL（秒）
func (e *Etcd) GetLeaseTTL() int64 {
	if e.LeaseTTL <= 0 {
		return 30 // 默认30秒
	}
	return e.LeaseTTL
}

// Server 服务器配置
type Server struct {
	ServerName string `yaml:"serverName"`
	Port       int    `yaml:"port"`
	Host       string `yaml:"host"`
	Version    string `yaml:"version"`
}

// IsDev 判断是否为开发环境
func (config *AllConfig) IsDev() bool {
	return !strings.Contains(ConfigPath, "prod")
}

// ReadConfig 读取配置文件
func ReadConfig(configPath string) *AllConfig {
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("警告: 无法读取配置文件: %s", err)

		// 尝试创建默认配置文件
		if os.IsNotExist(err) {
			log.Fatalf("配置文件不存在，服务停止")
			return nil
		}

		log.Fatalf("配置文件格式不正确: %s", err)
	}

	var allConfig AllConfig
	if err := viper.Unmarshal(&allConfig); err != nil {
		log.Fatalf("解析配置文件失败: %s", err)
	}

	return &allConfig
}

type KafkaConfig struct {
	Brokers   []string `yaml:"brokers"`
	Topic     string   `yaml:"topic"`
	GroupID   string   `yaml:"groupId"`
	BatchSize int      `yaml:"batchSize"`
	Async     bool     `yaml:"async"`
}

// KafkaBrokers 获取Kafka broker地址列表
func (k *KafkaConfig) KafkaBrokers() []string {
	if len(k.Brokers) == 0 {
		return []string{"localhost:9092"}
	}
	return k.Brokers
}

// KafkaDSN 获取Kafka连接字符串
func (k *KafkaConfig) KafkaDSN() string {
	return strings.Join(k.KafkaBrokers(), ",")
}

// MongoDB MongoDB配置
type MongoDB struct {
	URI             string `yaml:"uri"`
	MaxPoolSize     int    `yaml:"maxPoolSize"`
	MinPoolSize     int    `yaml:"minPoolSize"`
	MaxConnIdleTime int    `yaml:"maxConnIdleTime"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
}
