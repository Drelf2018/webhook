package config

type Permission struct {
	// 所有者
	Owner string `yaml:"owner"`
	// 管理员
	Administrators []string `yaml:"administrators"`
	// 信任者
	Trustors []string `yaml:"trustors"`
}
