package userdata

import (
	"context"
	"encoding/json"
	"fa-middleware/config"
	"fmt"

	"github.com/jackc/pgx/v4"
)

func SetUserData(conf config.Config, id string, userData interface{}) error {
	connStr := fmt.Sprintf(
		"postgres://%v:%v@%v:%v/%v?%v",
		conf.PostgresUser,
		conf.PostgresPass,
		conf.PostgresHost,
		conf.PostgresPort,
		conf.PostgresDBName,
		conf.PostgresOptions,
	)

	userDataBytes, err := json.Marshal(userData)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %v", err.Error())
	}

	// https://github.com/jackc/pgx#example-usage
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err.Error())
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(
		context.Background(),
		"insert into userdata(id, user_data_json) values($1, $2) on conflict (id) do update set user_data_json = EXCLUDED.user_data_json",
		id,
		string(userDataBytes),
	)

	if err != nil {
		return fmt.Errorf("failed to upsert id %v: %v", id, err.Error())
	}

	return nil
}
