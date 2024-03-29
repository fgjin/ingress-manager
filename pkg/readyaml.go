package pkg

import (
	"errors"

	"github.com/spf13/viper"
)

// // 定义结构体以映射 YAML 内容
// type IngressYAML struct {
// 	APIVersion string `mapstructure:"apiVersion"`
// 	Kind       string `mapstructure:"kind"`
// 	Metadata   struct {
// 		Name        string            `mapstructure:"name"`
// 		Namespace   string            `mapstructure:"namespace"`
// 		Annotations map[string]string `mapstructure:"annotations"`
// 	} `mapstructure:"metadata"`
// 	Spec struct {
// 		IngressClassName string `mapstructure:"ingressClassName"`
// 		Rules            []struct {
// 			Host string `mapstructure:"host"`
// 			HTTP struct {
// 				Paths []struct {
// 					Path     string `mapstructure:"path"`
// 					PathType string `mapstructure:"pathType"`
// 					Backend  struct {
// 						Service struct {
// 							Name string `mapstructure:"name"`
// 							Port struct {
// 								Number int32 `mapstructure:"number"`
// 							} `mapstructure:"port"`
// 						} `mapstructure:"service"`
// 					} `mapstructure:"backend"`
// 				} `mapstructure:"paths"`
// 			} `mapstructure:"http"`
// 		} `mapstructure:"rules"`
// 	} `mapstructure:"spec"`
// }

// 定义 Metadata 结构体
type Metadata struct {
	Name        string            `mapstructure:"name"`
	Namespace   string            `mapstructure:"namespace"`
	Annotations map[string]string `mapstructure:"annotations"`
}

// 定义 Backend 结构体
type Backend struct {
	Service BackendService `mapstructure:"service"`
}

type BackendService struct {
	Name string             `mapstructure:"name"`
	Port BackendServicePort `mapstructure:"port"`
}

type BackendServicePort struct {
	Number int32 `mapstructure:"number"`
}

// 定义 Path 结构体
type Path struct {
	Path     string  `mapstructure:"path"`
	PathType string  `mapstructure:"pathType"`
	Backend  Backend `mapstructure:"backend"`
}

// 定义 HTTP 结构体
type HTTP struct {
	Paths []Path `mapstructure:"paths"`
}

// 定义 Rule 结构体
type Rule struct {
	Host string `mapstructure:"host"`
	HTTP HTTP   `mapstructure:"http"`
}

// 定义 Spec 结构体
type Spec struct {
	IngressClassName string `mapstructure:"ingressClassName"`
	Rules            []Rule `mapstructure:"rules"`
}

// 定义 IngressYAML 结构体
type IngressYAML struct {
	APIVersion string   `mapstructure:"apiVersion"`
	Kind       string   `mapstructure:"kind"`
	Metadata   Metadata `mapstructure:"metadata"`
	Spec       Spec     `mapstructure:"spec"`
}

func (i *IngressYAML) GetIngressClassName() string {
	return i.Spec.IngressClassName
}

func (i *IngressYAML) GetHost() string {
	// 检查是否存在规则
	if len(i.Spec.Rules) > 0 {
		// 如果有规则，则返回第一个规则的 Host 字段值
		return i.Spec.Rules[0].Host
	}
	// 如果没有规则，则返回空字符串
	return ""
}

func (i *IngressYAML) GetPath() string {
	if len(i.Spec.Rules) > 0 {
		return i.Spec.Rules[0].HTTP.Paths[0].Path
	}
	return "/"
}

func (i *IngressYAML) GetNumber() int32 {
	if len(i.Spec.Rules) > 0 {
		return i.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number
	}
	return 80
}

func ReadYaml(path string, value interface{}) error {
	// 初始化 viper
	viper.SetConfigFile(path)   // 设置配置文件名
	err := viper.ReadInConfig() // 读取配置文件
	if err != nil {
		return errors.New("Failed to read config file: " + err.Error())
	}

	// 解析 YAML 文件内容到结构体
	err = viper.Unmarshal(&value)
	if err != nil {
		return errors.New("Failed to unmarshal YAML: " + err.Error())
	}
	// 输出解析结果
	// log.Printf("Ingress YAML: %+v", ingressYAML)
	return nil
}
