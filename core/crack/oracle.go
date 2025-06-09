package crack

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/sijms/go-ora/v2"
)

type OracleCracker struct {
	CrackBase
	ServiceName string
}

func (s *OracleCracker) Ping() (succ bool, err error) {
	return false, errors.ErrUnsupported
}

func (s *OracleCracker) Crack() (succ bool, err error) {
	var timeout = 3
	if s.Timeout > 0 {
		timeout = s.Timeout
	}
	targetSplit := strings.Split(s.Target, ":")
	if len(targetSplit) != 2 {
		return
	}
	ip := targetSplit[0]
	port := targetSplit[1]
	var conn *sql.DB
	if s.ServiceName != "" {
		conn, err = serviceNameLogin(s.ServiceName, s.User, s.Pass, ip, port, timeout)
	} else {
		conn, err = sidLogin("orcl", s.User, s.Pass, ip, port, timeout)
	}
	if err != nil {
		return false, err
	}
	defer conn.Close()
	err = conn.Ping()
	if err != nil {
		return false, err
	}
	return true, nil
}

func sidLogin(sid, user, pass, ip, port string, timeout int) (*sql.DB, error) {
	if sid == "" {
		sid = "orcl"
	}
	connStr := fmt.Sprintf("oracle://%s:%s@%s:%s/%s?connection_timeout=%d&connection_pool_timeout=%d", user,
		pass, ip, port, sid, timeout, timeout)

	conn, err := sql.Open("oracle", connStr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func serviceNameLogin(serviceName, user, pass, ip, port string, timeout int) (*sql.DB, error) {
	connStr := fmt.Sprintf("oracle://%s:%s@%s:%s/?service_name=%s&connection_timeout=%d&connection_pool_timeout=%d", user,
		pass, ip, port, serviceName, timeout, timeout)

	conn, err := sql.Open("oracle", connStr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
