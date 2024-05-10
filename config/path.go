package config

import (
	"os"
	"path/filepath"

	"github.com/Drelf2018/initial/fullpath"
	"github.com/Drelf2018/webhook/database/dao"
	"github.com/Drelf2020/utils"
)

type Path struct {
	Root     string `yaml:"root"   default:"resource"`
	Log      string `yaml:"log"    default:".log"     join:"Root"`
	Views    string `yaml:"views"  default:"views"    join:"Root"`
	UserDB   string `yaml:"userDB" default:"users.db" join:"Root"`
	Public   string `yaml:"public" default:"public"   join:"Root"`
	PostDB   string `yaml:"postDB" default:"posts.db" join:"Public"`
	FullPath *Path  `yaml:"-" initial:"-"`
}

func (p *Path) AfterDefault() (err error) {
	p.FullPath, err = fullpath.New(*p)
	if err != nil {
		return
	}

	err = os.MkdirAll(filepath.Dir(p.FullPath.Log), os.ModePerm)
	if err != nil {
		return err
	}
	utils.SetOutputFile(p.FullPath.Log)

	return dao.Open(p.FullPath.PostDB, p.FullPath.UserDB)
}
