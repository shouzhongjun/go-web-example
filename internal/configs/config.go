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
	Model       string        `yaml:"model"`
	Server      Server        `yaml:"server"`
	Log         Log           `yaml:"log"`
	Cors        *Cors         `yaml:"cors"`
	Trace       *Trace        `yaml:"trace"`
	Database    Database      `yaml:"database"`
	Redis       Redis         `yaml:"redis"`
	Kafka       KafkaConfig   `yaml:"kafka"`
	Etcd        *Etcd         `yaml:"etcd"`
	MongoDB     *MongoDB      `yaml:"mongodb"`
	JWT         JWTConfig     `yaml:"jwt"`
	Swagger     Swagger       `yaml:"swagger"`
	RateLimiter *RateLimiter  `yaml:"rateLimiter"`
	OpenAPI     OpenAPIConfig `yaml:"openapi"`
}

// Trace 链路追踪配置
type Trace struct {
	ServiceName    string        `yaml:"serviceName"`    // 服务名称
	ServiceVersion string        `yaml:"serviceVersion"` // 服务版本
	Environment    string        `yaml:"environment"`    // 环境（dev/prod等）
	Endpoint       string        `yaml:"endpoint"`       // 链路追踪服务器地址
	SamplingRatio  float64       `yaml:"samplingRatio"`  // 采样率
	Enable         bool          `yaml:"enable"`         // 是否启用链路追踪
	BatchTimeout   time.Duration `yaml:"batchTimeout"`   // 批处理超时时间
	MaxBatchSize   int           `yaml:"maxBatchSize"`   // 最大批处理大小
	MaxQueueSize   int           `yaml:"maxQueueSize"`   // 最大队列大小
	ClientTimeout  time.Duration `yaml:"clientTimeout"`  // 客户端超时时间
	RetryInitial   time.Duration `yaml:"retryInitial"`   // 重试初始间隔
	RetryMax       time.Duration `yaml:"retryMax"`       // 最大重试间隔
	RetryElapsed   time.Duration `yaml:"retryElapsed"`   // 重试总时长
}

// Log 日志配置
type Log struct {
	Level         string `yaml:"level"`
	EnableFile    bool   `yaml:"enableFile"`
	EnableConsole bool   `yaml:"enableConsole"`
	Prefix        string `yaml:"prefix"`
	Path          string `yaml:"path"`
	PrintParam    bool   `yaml:"printParam"`
	// 日志文件压缩配置
	MaxSize    int  `yaml:"maxSize"`    // 单个日志文件最大大小，单位MB，默认100MB
	MaxBackups int  `yaml:"maxBackups"` // 保留的旧日志文件最大数量，默认保留所有
	MaxAge     int  `yaml:"maxAge"`     // 保留的旧日志文件最大天数，默认保留所有
	Compress   bool `yaml:"compress"`   // 是否压缩旧日志文件，默认不压缩
}

type Cors struct {
	Enable              bool     `yaml:"enable"`
	AllowedOrigins      []string `yaml:"allowedOrigins"`
	AllowedMethods      []string `yaml:"allowedMethods"`
	AllowedHeaders      []string `yaml:"allowedHeaders"`
	ExposeHeaders       []string `yaml:"exposeHeaders"`
	AllowCredentials    bool     `yaml:"allowCredentials"`
	MaxAge              int      `yaml:"maxAge"`
	AllowPrivateNetwork bool     `yaml:"allowPrivateNetwork"`
}

// Database 数据库配置
type Database struct {
	SSLMode         string `yaml:"ssl_mode"`
	MaxOpenConns    int    `yaml:"maxOpen_conns"`
	ConnMaxLifetime *int64 `yaml:"connMaxLifetime"`
	Host            string `yaml:"host"`
	UserName        string `yaml:"username"`
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
		db.UserName, db.Password, db.Host, db.Port, db.DBName)

	//return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
	//	db.Host, db.UserName, db.Password, db.DBName, db.Port, db.SSLMode)
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
	Host      string   `yaml:"host"`
	Port      int      `yaml:"port"`
	Brokers   []string `yaml:"brokers"`
	Topic     string   `yaml:"topic"`
	GroupID   string   `yaml:"groupId"`
	BatchSize int      `yaml:"batchSize"`
	Async     bool     `yaml:"async"`
	Enable    bool     `yaml:"enable"`
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

type JWTConfig struct {
	SecretKey string        `yaml:"secretKey"`
	Issuer    string        `yaml:"issuer"`
	Duration  time.Duration `yaml:"duration"`
}

// Swagger Swagger 配置
type Swagger struct {
	Enable bool `yaml:"enable"` // 是否启用 Swagger
}

type AutoGenerated struct {
	RateLimiter RateLimiter `yaml:"rateLimiter"`
}
type Local struct {
	Capacity              int `yaml:"capacity"`
	Interval              int `yaml:"interval"`
	MaxWaitCount          int `yaml:"maxWaitCount"`
	MaxWaitCountPerDay    int `yaml:"maxWaitCountPerDay"`
	MaxWaitCountPerDecade int `yaml:"maxWaitCountPerDecade"`
	MaxWaitCountPerHour   int `yaml:"maxWaitCountPerHour"`
	MaxWaitCountPerMinute int `yaml:"maxWaitCountPerMinute"`
	MaxWaitCountPerMonth  int `yaml:"maxWaitCountPerMonth"`
	MaxWaitCountPerSecond int `yaml:"maxWaitCountPerSecond"`
	MaxWaitCountPerYear   int `yaml:"maxWaitCountPerYear"`
	MaxWaitTime           int `yaml:"maxWaitTime"`
}
type RateLimiter struct {
	Burst    int    `yaml:"burst"`
	Duration int    `yaml:"duration"`
	Enable   bool   `yaml:"enable"`
	Local    Local  `yaml:"local"`
	Rate     int    `yaml:"rate"`
	Strategy string `yaml:"strategy"`
}

// OpenAPIConfig OpenAPI配置
type OpenAPIConfig struct {
	Enable bool `yaml:"enable"` // 是否启用OpenAPI
}

// GetBatchTimeout 获取批处理超时时间，如果未配置则返回默认值
func (t *Trace) GetBatchTimeout() time.Duration {
	if t.BatchTimeout <= 0 {
		return 5 * time.Second
	}
	return t.BatchTimeout
}

// GetMaxBatchSize 获取最大批处理大小，如果未配置则返回默认值
func (t *Trace) GetMaxBatchSize() int {
	if t.MaxBatchSize <= 0 {
		return 100
	}
	return t.MaxBatchSize
}

// GetMaxQueueSize 获取最大队列大小，如果未配置则返回默认值
func (t *Trace) GetMaxQueueSize() int {
	if t.MaxQueueSize <= 0 {
		return 1000
	}
	return t.MaxQueueSize
}

// GetClientTimeout 获取客户端超时时间，如果未配置则返回默认值
func (t *Trace) GetClientTimeout() time.Duration {
	if t.ClientTimeout <= 0 {
		return 5 * time.Second
	}
	return t.ClientTimeout
}

// GetRetryInitial 获取重试初始间隔，如果未配置则返回默认值
func (t *Trace) GetRetryInitial() time.Duration {
	if t.RetryInitial <= 0 {
		return time.Second
	}
	return t.RetryInitial
}

// GetRetryMax 获取最大重试间隔，如果未配置则返回默认值
func (t *Trace) GetRetryMax() time.Duration {
	if t.RetryMax <= 0 {
		return 5 * time.Second
	}
	return t.RetryMax
}

// GetRetryElapsed 获取重试总时长，如果未配置则返回默认值
func (t *Trace) GetRetryElapsed() time.Duration {
	if t.RetryElapsed <= 0 {
		return 30 * time.Second
	}
	return t.RetryElapsed
}
