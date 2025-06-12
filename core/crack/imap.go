package crack

import (
	"net"
	"strings"

	"github.com/dusbot/maxx/libs/utcp"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-sasl"
)

type ImapCracker struct {
	CrackBase
}

func (f *ImapCracker) Ping() (succ bool, err error) {
	var conn net.Conn
	conn, err = utcp.NewDialer().Dial(f.Target, f.Timeout)
	if err != nil {
		return false, ERR_CONNECTION
	}
	c := imapclient.New(conn, &imapclient.Options{})
	defer c.Close()
	if err = c.List("", "%", nil).Wait(); err != nil {
		// The handling here is critical. For the 'Ping then Crack' approach.
		// we must first verify that the target uses IMAP protocol.
		// Therefore, all returned errors must strictly conform to IMAP format.
		if _, ok := err.(*imap.Error); ok {
			return false, nil
		}
		return false, ERR_CONNECTION
	}
	return true, nil
}

func (f *ImapCracker) Crack() (succ bool, err error) {
	var conn net.Conn
	conn, err = utcp.NewDialer().Dial(f.Target, f.Timeout)
	if err != nil {
		return
	}
	c := imapclient.New(conn, &imapclient.Options{})
	defer c.Close()
	authMechanisms := c.Caps().AuthMechanisms()
	if len(authMechanisms) > 0 {
		var authClient sasl.Client
		for _, ext := range authMechanisms {
			switch ext {
			case "CRAM-MD5":
				authClient = newCramMD5Client(f.User, f.Pass)
			case "LOGIN":
				authClient = sasl.NewLoginClient(f.User, f.Pass)
			case "PLAIN":
				authClient = sasl.NewPlainClient("", f.User, f.Pass)
			}
			if strings.Contains(ext, "SCRAM") {
				authClient, err = newScramClient(ext, f.User, f.Pass)
				if err != nil {
					return false, err
				}
			}
			if authClient != nil {
				break
			}
		}
		if authClient != nil {
			if err := c.Authenticate(authClient); err != nil {
				return false, err
			}
		}
	} else {
		if err := c.Login(f.User, f.Pass).Wait(); err != nil {
			return false, err
		}
		defer c.Logout().Wait()
	}
	if err := c.List("", "%", nil).Wait(); err != nil {
		if strings.Contains(err.Error(), "unexpected EOF") {
			return true, nil
		}
		return false, err
	}
	return true, nil
}

func (*ImapCracker) Class() string {
	return CLASS_EMAIL
}
