package crack

import (
	"fmt"

	"gopkg.in/olivere/elastic.v3"
)

type ElasticsearchCracker struct {
	CrackBase
}

func (s *ElasticsearchCracker) Ping() (succ bool, err error) {
	client, err := elastic.NewClient(elastic.SetURL(fmt.Sprintf("http://%v:%v", s.Ip, s.Port)),
		elastic.SetMaxRetries(1),
	)
	if err == nil {
		defer client.Stop()
		_, _, err = client.Ping(fmt.Sprintf("http://%v:%v", s.Ip, s.Port)).Do()
		if err == nil {
			return true, nil
		}
	}
	return false, err
}

func (s *ElasticsearchCracker) Crack() (succ bool, err error) {
	client, err := elastic.NewClient(elastic.SetURL(fmt.Sprintf("http://%v:%v", s.Ip, s.Port)),
		elastic.SetMaxRetries(1),
		elastic.SetBasicAuth(s.User, s.Pass),
	)
	if err == nil {
		defer client.Stop()
		_, _, err = client.Ping(fmt.Sprintf("http://%v:%v", s.Ip, s.Port)).Do()
		if err == nil {
			return true, nil
		}
	}
	return false, err
}
