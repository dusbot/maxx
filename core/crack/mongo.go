package crack

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoCracker struct {
	CrackBase
}

func (r *MongoCracker) Crack() (succ bool, err error) {
	var timeout = 3
	if r.Timeout > 0 {
		timeout = r.Timeout
	}
	var url string
	if r.Pass == "" {
		url = fmt.Sprintf("mongodb://%s", r.Target)
	} else {
		url = fmt.Sprintf("mongodb://%v:%v@%s", r.User, r.Pass, r.Target)
	}
	clientOptions := options.Client().ApplyURI(url).SetConnectTimeout(time.Duration(timeout) * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return false, err
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, nil)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (*MongoCracker) Class() string {
	return CLASS_DB
}
