package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Market   MarketConfig   `mapstructure:"market"`
	Trading  TradingConfig  `mapstructure:"trading"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

type ServerConfig struct {
	Port    int    `mapstructure:"port"`
	Mode    string `mapstructure:"mode"`
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Name            string `mapstructure:"name"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type MarketConfig struct {
	UpdateInterval string            `mapstructure:"update_interval"`
	DataSource     string            `mapstructure:"data_source"`
	APIURL         string            `mapstructure:"api_url"`
	Symbols        []string          `mapstructure:"symbols"`
	Hyperliquid    HyperliquidConfig `mapstructure:"hyperliquid"`
}

type HyperliquidConfig struct {
	InfoEndpoint string `mapstructure:"info_endpoint"`
	WSEndpoint   string `mapstructure:"ws_endpoint"`
}

type TradingConfig struct {
	DefaultFeeRate float64 `mapstructure:"default_fee_rate"`
	MakerFeeRate   float64 `mapstructure:"maker_fee_rate"`
	TakerFeeRate   float64 `mapstructure:"taker_fee_rate"`
	MinOrderAmount float64 `mapstructure:"min_order_amount"`
}

type AuthConfig struct {
	JWTSecret   string `mapstructure:"jwt_secret"`
	TokenExpire int    `mapstructure:"token_expire"`
}

type LoggingConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// Load 加载配置
func Load() (*Config, error) {
	v := viper.New()

	// 设置配置文件名和路径
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	// 环境变量
	v.SetEnvPrefix("QS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// GetDSN 返回数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}
