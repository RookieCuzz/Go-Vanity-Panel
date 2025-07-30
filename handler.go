package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type repo struct {
	VCS     string
	Path    string
	Import  string
	Source  string
	Display string
}

type data struct {
	Host  string
	Repos []repo
}

type handler struct {
	conf *Configuration
	data *data
}

func newHandler(conf *Configuration) *handler {
	h := new(handler)
	h.conf = conf
	h.data = &data{
		Host:  h.conf.Host,
		Repos: []repo{},
	}
	for p, r := range h.conf.Paths {
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		h.data.Repos = append(h.data.Repos, repo{
			Path:    p,
			Import:  h.conf.Host + p,
			VCS:     r.VCS.String(),
			Source:  r.Repo,
			Display: r.SCL(),
		})
	}
	return h
}

func (h *handler) getIndex() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := indexTpl.Execute(buf, h.data)
	return buf.Bytes(), err
}

func (h *handler) getRepo(path string) ([]byte, error) {
	for _, r := range h.data.Repos {
		if r.Path == path {
			buf := bytes.NewBuffer(nil)
			err := repoTpl.Execute(buf, r)
			return buf.Bytes(), err
		}
	}
	return nil, errors.New("unknown path")
}

func (h *handler) getConfigPanel() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := configPanelTpl.Execute(buf, map[string]interface{}{
		"Config": h.conf,
	})
	return buf.Bytes(), err
}

func (h *handler) updateConfig(newConfig *Configuration) error {
	// 更新配置
	h.conf = newConfig
	
	// 重新构建数据
	h.data = &data{
		Host:  h.conf.Host,
		Repos: []repo{},
	}
	
	// 调试信息
	fmt.Printf("Updating config with %d paths\n", len(h.conf.Paths))
	
	for p, r := range h.conf.Paths {
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		repo := repo{
			Path:    p,
			Import:  h.conf.Host + p,
			VCS:     r.VCS.String(),
			Source:  r.Repo,
			Display: r.SCL(),
		}
		h.data.Repos = append(h.data.Repos, repo)
		fmt.Printf("Added path: %s -> %s (%s)\n", p, r.Repo, r.VCS.String())
	}
	
	fmt.Printf("Total repos after update: %d\n", len(h.data.Repos))
	return nil
}

func (h *handler) cache() string {
	return fmt.Sprintf("public, max-age=%d", h.conf.CacheMaxAge)
}
