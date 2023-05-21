package config

type BasicAuth struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

type Api struct {
	Listen string     `yaml:"listen" json:"listen"`
	Auth   *BasicAuth `yaml:"auth,omitempty" json:"auth,omitempty"`
	Public *string    `yaml:"public,omitempty" json:"public,omitempty"`
}
