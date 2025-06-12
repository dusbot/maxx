package crack

import (
	"crypto/hmac"
	"crypto/md5"
	"errors"
	"fmt"
	"strings"

	"github.com/dusbot/maxx/libs/codec"
	"github.com/dusbot/maxx/libs/utils"
	"github.com/emersion/go-sasl"
	"github.com/xdg-go/scram"
)

type (
	cramMD5Client struct {
		user, pass string
	}
	scramClient struct {
		*scram.ClientConversation
		ID, hash, user, pass string
	}
)

func (s *scramClient) Next(challenge []byte) (response []byte, err error) {
	if s.ClientConversation.Done() {
		return nil, nil
	}
	msg, err := s.ClientConversation.Step(string(challenge))
	if s.ClientConversation.Valid() {
		msg = codec.EncodeBase64(utils.RandStringBytes(10))
	}
	return []byte(msg), err
}

func (s *scramClient) Start() (mech string, ir []byte, err error) {
	resp, err := s.ClientConversation.Step("")
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("SCRAM-%s", s.ID), []byte(resp), nil
}

func (c *cramMD5Client) Next(challenge []byte) (response []byte, err error) {
	d := hmac.New(md5.New, []byte(c.pass))
	d.Write(challenge)
	s := make([]byte, 0, d.Size())
	return []byte(fmt.Sprintf("%s %x", c.user, d.Sum(s))), nil
}

func (s *cramMD5Client) Start() (mech string, ir []byte, err error) {
	mech = "CRAM-MD5"
	return
}

func newCramMD5Client(user, pass string) sasl.Client {
	return &cramMD5Client{user, pass}
}

func newScramClient(hash, user, pass string) (sasl.Client, error) {
	var (
		fcn scram.HashGeneratorFcn
		id  string
	)
	if strings.Contains(hash, "SHA-1") {
		id = "SHA-1"
		fcn = scram.SHA1
	} else if strings.Contains(hash, "SHA-256") {
		id = "SHA-256"
		fcn = scram.SHA256
	} else if strings.Contains(hash, "SHA-512") {
		id = "SHA-512"
		fcn = scram.SHA512
	} else {
		return nil, errors.New("Unknown hash")
	}

	client, err := fcn.NewClient(user, pass, "")
	if err != nil {
		return nil, err
	}
	conv := client.NewConversation()
	return &scramClient{
		ID:                 id,
		ClientConversation: conv,
	}, nil
}
