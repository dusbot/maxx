package crack

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
)

type MssqlCracker struct {
	CrackBase
}

func (s *MssqlCracker) Ping() (succ bool, err error) {
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
	dataSourceName := fmt.Sprintf("server=%s;port=%s;user id=%s;password=%v;database=%v;connection timeout=%v;encrypt=disable", ip,
		port, "sa", "", "master", timeout)

	conn, err := sql.Open("mssql", dataSourceName)
	if err != nil {
		return false, err
	}

	err = conn.Ping()
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unable to open tcp connection") {
			return false, ERR_CONNECTION
		}
		return false, err
	}
	defer conn.Close()
	return true, nil
}

func (s *MssqlCracker) Crack() (succ bool, err error) {
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
	dataSourceName := fmt.Sprintf("server=%s;port=%s;user id=%s;password=%s;database=%s;connection timeout=%d;encrypt=disable", ip,
		port, s.User, s.Pass, "master", timeout)

	conn, err := sql.Open("mssql", dataSourceName)
	if err != nil {
		return false, err
	}

	err = conn.Ping()
	if err != nil {
		return false, err
	}
	defer conn.Close()
	return true, nil
}
