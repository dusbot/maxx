package crack

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

type MysqlCracker struct {
	CrackBase
}

func (s *MysqlCracker) Ping() (succ bool, err error) {
	return false, errors.ErrUnsupported
}

func (s *MysqlCracker) Crack() (succ bool, err error) {
	var timeout = 3
	if s.Timeout > 0 {
		timeout = s.Timeout
	}
	mysql.SetLogger(nilLog{})
	dataSourceName := fmt.Sprintf("%v:%v@tcp(%s)/?timeout=%ds&readTimeout=%ds&writeTimeout=%ds&charset=utf8", s.User,
		s.Pass, s.Target, timeout, timeout, timeout)
	conn, err := sql.Open("mysql", dataSourceName)
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

type nilLog struct {
}

func (l nilLog) Print(v ...interface{}) {

}
