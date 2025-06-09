package crack

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

type PostgresCracker struct {
	CrackBase
	DBName string
}

func (f *PostgresCracker) Ping() (succ bool, err error) {
	return false, errors.ErrUnsupported
}

func (f *PostgresCracker) Crack() (succ bool, err error) {
	var timeout = 3
	if f.Timeout > 0 {
		timeout = f.Timeout
	}
	targetSplit := strings.Split(f.Target, ":")
	if len(targetSplit) != 2 {
		return
	}
	ip := targetSplit[0]
	port := targetSplit[1]
	dataSourceName := strings.Join([]string{
		fmt.Sprintf("connect_timeout=%d", timeout),
		fmt.Sprintf("dbname=%s", "test"),
		fmt.Sprintf("host=%v", ip),
		fmt.Sprintf("password=%v", f.Pass),
		fmt.Sprintf("port=%v", port),
		"sslmode=disable",
		fmt.Sprintf("user=%v", f.User),
	}, " ")

	conn, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	err = conn.Ping()
	if err != nil {
		if strings.Contains(err.Error(), "not exist") {
			return true, nil
		}
		return false, err
	}
	return true, nil
}
