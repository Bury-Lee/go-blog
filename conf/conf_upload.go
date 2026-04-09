package conf

type UploadConfig struct {
	Size      int64               `yaml:"size"`
	WhiteList map[string]struct{} `yaml:"whiteList"` //小改了一下,使用map实现o1的判断
	UploadDir string              `yaml:"uploadDir"`
}
