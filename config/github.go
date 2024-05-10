package config

import (
	"fmt"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Github struct {
	Folder     string `yaml:"-"`
	Username   string `yaml:"username"   default:"Drelf2018"`
	Repository string `yaml:"repository" default:"gin.nana7mi.link"`
	Branche    string `yaml:"branche"    default:"gh-pages"`

	done chan int
}

func (g *Github) Api() string {
	return fmt.Sprintf("https://api.github.com/repos/%v/%v/branches/%v", g.Username, g.Repository, g.Branche)
}

func (g *Github) Git() string {
	return fmt.Sprintf("https://github.com/%s/%s.git", g.Username, g.Repository)
}

func (g *Github) ReferenceName() plumbing.ReferenceName {
	return plumbing.ReferenceName("refs/heads/" + g.Branche)
}

func (g *Github) RemoteName() plumbing.ReferenceName {
	return plumbing.ReferenceName("refs/remotes/origin/" + g.Branche)
}

func (g Github) String() string {
	return fmt.Sprintf("https://github.com/%s/%s/tree/%s", g.Username, g.Repository, g.Branche)
}

// 克隆到文件夹
func (g *Github) Clone() error {
	_, err := git.PlainClone(g.Folder, false, &git.CloneOptions{
		URL:           g.Git(),
		ReferenceName: g.ReferenceName(),
	})
	if err == git.ErrRepositoryAlreadyExists {
		return nil
	}
	return err
}

func (g *Github) ForcePull() error {
	repo, _ := git.PlainOpen(g.Folder)
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}
	return worktree.Pull(&git.PullOptions{ReferenceName: g.ReferenceName()})
}

func (g *Github) SyncRepository() error {
	repo, err := git.PlainOpen(g.Folder)
	if err != nil {
		return err
	}

	err = repo.Fetch(&git.FetchOptions{})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = worktree.Pull(&git.PullOptions{ReferenceName: g.ReferenceName()})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (g *Github) AfterConfigGithub(c *Config) {
	g.Folder = c.Path.FullPath.Views

	go func() {
		i := 0
		for g.Clone() != nil {
			i++
			time.Sleep(3 * time.Second)
		}

		if g.done != nil {
			g.done <- i
			close(g.done)
		}
	}()
}

func (g *Github) Done() int {
	g.done = make(chan int)
	return <-g.done
}
