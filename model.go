package proxy

type SchemaType string

const (
	SchemaHTTP  SchemaType = "http://"
	SchemaHTTPS SchemaType = "https://"

	HeaderAuthKey = "Proxy-Auth-Key"
)

// 配置文件
type Config struct {
	ProxySchema  SchemaType `json:"proxy_schema" yaml:"proxy_schema" toml:"proxy_schema"`       // SchemaHTTP or SchemaHTTPS
	ProxyHost    string     `json:"proxy_host" yaml:"proxy_host" toml:"proxy_host"`             // 转发到的接口 Host
	ProxyPort    string     `json:"proxy_port" yaml:"proxy_port" toml:"proxy_port"`             // 转发到的接口 Port
	ServerPort   string     `json:"server_port" yaml:"server_port" toml:"server_port"`          // 代理转发服务启动的端口
	ProxyAuthKey string     `json:"proxy_auth_key" yaml:"proxy_auth_key" toml:"proxy_auth_key"` // 代理请求的校验Key
	ShowLog      bool       `json:"show_log" yaml:"show_log" toml:"show_log"`                   // 是否展示转发记录
}
