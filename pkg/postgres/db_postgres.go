package db_postgres

import (
	"context"
	"encoding/json"
	"fmt"
	localSetup "ganesh.provengo.io/internal/setup"
	localStructs "ganesh.provengo.io/internal/structs"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"sync"
)

type Postgres struct {
	db *pgxpool.Pool
}

type UserLink struct {
	Username string
	Link     string
}

type ProvengoValues struct {
	identity string
	value    string
}

type BeaconValues struct {
	BeaconId  string
	Timestamp pgtype.Timestamptz
}

type BeaconValuesList struct {
	BeaconId string
}

type BeaconListValues struct {
	BeaconId string
}

type DefineUser struct {
	UserName  string
	Name      string
	Telephone string
	Secret    string
	UserLevel string
}

type DefineUpdateUser struct {
	UserName  string
	Name      string
	Telephone string
	UserLevel string
}

type DefineUpdatePassword struct {
	UserName        string
	Password        string
	CurrentPassword string
}

type DefineBeacon struct {
	Beacon  string
	Nanoid  string
	Ipaddr  string
	Headers string
}

type PasswordUserSecret struct {
	secret string
}

type StringLog struct {
	Type   string
	From   string
	Values string
}

type SessionInsertValues struct {
	Username  string `json:"username"`
	Name      string `json:"name"`
	Telephone string `json:"telephone"`
	BeaconId  string `json:"beacon-id"`
	UserLevel string `json:"user-level"`
}

type SessionInsertValuesString struct {
	Username  string
	Name      string
	Telephone string
	BeaconId  string
	UserLevel string
}

type SessionData struct {
	ApplicationID string
	Values        string
}

type SessionInsert struct {
	LinkId        string
	ApplicationId string
	Values        SessionInsertValuesString
}

type DefineDomain struct {
	DomainId    int
	Domain      string
	Description string
	Public      int
	UserID      int
}

type DefineProbe struct {
	ProbeId       int
	DomainID      int
	Probe         string
	Description   string
	Configuration string
	Status        int
	Identity      string
}

type ProbeConfig struct {
	Address         []string
	Icmp            int
	Tcp             int
	TcpPort         []int
	Snmp            int
	SnmpV2Community string
}

type DefineCollector struct {
	HostName string
	HostID   string
	IpAddr   string
	ProbeId  int64
	Values   interface{}
	RemoteTS int64
	LocalTS  int64
}

var (
	pgInstance *Postgres
	pgOnce     sync.Once
)

func PostgresConnection(ctx context.Context, connString string) (*Postgres, error) {
	var err error
	pgOnce.Do(func() {
		config, _err := pgxpool.ParseConfig(connString)
		if _err != nil {
			err = fmt.Errorf("error parsing connection string: %w", _err)
			return
		}
		config.MinConns = localSetup.PostgresMin
		config.MaxConns = localSetup.PostgresMax

		db, _err := pgxpool.NewWithConfig(ctx, config)
		if _err != nil {
			err = fmt.Errorf("unable to create connection pool: %w", _err)
			return
		}
		pgInstance = &Postgres{db}
	})
	if err != nil {
		return nil, err
	}
	return pgInstance, nil
}

func (pg *Postgres) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}

func (pg *Postgres) Close() {
	pg.db.Close()
}

func (pg *Postgres) InsertUserPassword(ctx context.Context, data localStructs.DataLogin) error {
	queryInsert := `INSERT INTO users (uuid, username, password, hash) VALUES ($1, $2, $3, $4)`
	_, err := pg.db.Exec(ctx, queryInsert, data.UUID, data.Username, data.Password, "")
	if err != nil {
		return fmt.Errorf("unable to insert row: %w", err)
	}
	return nil
}

func (pg *Postgres) UpdateUserPassword(ctx context.Context, data localStructs.DataLogin) error {
	queryUpdate := `UPDATE users SET hash=$1 WHERE uuid=$2`
	_, err := pg.db.Exec(ctx, queryUpdate, data.Password, data.UUID)
	if err != nil {
		return fmt.Errorf("unable to insert row: %w", err)
	}
	return nil
}

func (pg *Postgres) InsertUser(ctx context.Context, data DefineUser, link string) error {
	queryLogin := `INSERT INTO provengo_login VALUES (default, $1, $2, $3, $4, '##', default,'##', default, 'DBCREATE', default)`
	_, err := pg.db.Exec(ctx, queryLogin, data.Name, data.UserName, data.Secret, data.Telephone)
	if err != nil {
		return fmt.Errorf("unable to insert row: %w", err)
	}

	queryLink := `INSERT INTO provengo_userlink VALUES ($1, $2, now(), now() + interval '15 minute', 202)`
	_, err = pg.db.Exec(ctx, queryLink, link, data.UserName)
	if err != nil {
		return fmt.Errorf("unable to insert row: %w", err)
	}

	return nil
}

func (pg *Postgres) UpdateUser(ctx context.Context, data DefineUpdateUser) error {
	queryUpdate := `UPDATE provengo_login SET name = $1, telephone = $2, userLevel = $3, last_action = 'USER_UPDATE' WHERE username = $4`
	res, err := pg.db.Exec(ctx, queryUpdate, data.Name, data.Telephone, data.UserLevel, data.UserName)
	if err != nil {
		return fmt.Errorf("unable to update user: %w", err)
	}

	affectedRows := res.RowsAffected()
	if affectedRows == 0 {
		return fmt.Errorf("unable to update user: %w", err)
	}

	return nil
}

func (pg *Postgres) UpdatePassword(ctx context.Context, data DefineUpdatePassword) error {
	queryUpdate := `UPDATE provengo_login SET secret = $1, last_action = 'PW_UPDATE' WHERE username = $2 AND secret = $3`
	res, err := pg.db.Exec(ctx, queryUpdate, data.Password, data.UserName, data.CurrentPassword)
	if err != nil {
		return fmt.Errorf("unable to update user secret: %w", err)
	}
	affectedRows := res.RowsAffected()
	if affectedRows == 0 {
		return fmt.Errorf("unable to update password: %w", err)
	}

	return nil
}

func (pg *Postgres) InsertBeacon(ctx context.Context, data DefineBeacon) {
	query := `INSERT INTO provengo_beacon VALUES ($1, $2, $3, $4, now())`

	_, err := pg.db.Exec(ctx, query, data.Beacon, data.Nanoid, data.Ipaddr, data.Headers)
	if err != nil {
		log.Printf("unable to insert row: %v", err)
	}
}

func (pg *Postgres) InsertLog(ctx context.Context, data StringLog) {
	query := `INSERT INTO provengo_logs VALUES (default, $1, $2, $3, default)`

	_, err := pg.db.Exec(ctx, query, data.Type, data.From, data.Values)
	if err != nil {
		log.Printf("unable to insert row: %v", err)
	}
}

func (pg *Postgres) InsertAccessLog(ctx context.Context, transactionId string, data interface{}, method string) {
	query := `INSERT INTO provengo_access_log VALUES (default, $1, $2, $3, now())`
	_, err := pg.db.Exec(ctx, query, transactionId, data, method)
	if err != nil {
		log.Printf("unable to insert row: %v", err)
	}
}

func (pg *Postgres) InsertSession(ctx context.Context, data SessionInsert) {
	query := `SELECT  sp_insert_session ($1, $2, $3)`

	dataInsert, _err := json.Marshal(data.Values)
	if _err != nil {
		log.Printf("unable to convert payload: %v", _err)
		dataInsert = []byte("{}")
	}

	_, err := pg.db.Exec(ctx, query, data.LinkId, data.ApplicationId, dataInsert)
	if err != nil {
		log.Printf("unable to insert row: %v", err)
	}
}

func (pg *Postgres) UpdateSessionValues(ctx context.Context, session string, data SessionInsertValuesString) error {
	query := `UPDATE provengo_sessions SET values = $1 WHERE application_id = $2`
	dataUpdate, _err := json.Marshal(data)
	if _err != nil {
		log.Printf("unable to convert payload: %v", _err)
		return fmt.Errorf("error converting payload: %w", _err)
	}
	_, err := pg.db.Exec(ctx, query, dataUpdate, session)
	if err != nil {
		return fmt.Errorf("update session values: %w", _err)
	}
	return nil
}

func (pg *Postgres) LogoutSession(ctx context.Context, session string) {
	query := `UPDATE provengo_sessions SET state = 99, valid_at = now() WHERE application_id = $1 and  state=200`

	_, err := pg.db.Exec(ctx, query, session)
	if err != nil {
		log.Printf("error logoff session %s", err)
	}

}

func (pg *Postgres) SelectSession(ctx context.Context, sessionID string) ([]SessionData, error) {
	query := `SELECT application_id, values from provengo_sessions where session_link = $1 and state=202`

	rows, err := pg.db.Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get data: %w", err)
	}
	defer rows.Close()

	var SessionSetup []SessionData

	for rows.Next() {
		var dataLine SessionData
		err := rows.Scan(&dataLine.ApplicationID, &dataLine.Values)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %w", err)
		}
		SessionSetup = append(SessionSetup, dataLine)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("row iteration error: %w", rows.Err())
	}

	return SessionSetup, nil
}

func (pg *Postgres) SelectSessionValues(ctx context.Context, applicationID string) (string, error) {
	query := `UPDATE provengo_sessions SET valid_at = NOW() + INTERVAL '2 minutes' WHERE application_id = $1 and state = 200 RETURNING values`

	rows, err := pg.db.Query(ctx, query, applicationID)
	if err != nil {
		return "", fmt.Errorf("get data: %w", err)
	}
	defer rows.Close()

	var dataLine string
	for rows.Next() {

		err := rows.Scan(&dataLine)
		if err != nil {
			return "", fmt.Errorf("unable to scan row: %w", err)
		}
	}

	if rows.Err() != nil {
		return "", fmt.Errorf("row iteration error: %w", rows.Err())
	}
	return dataLine, nil
}

func (pg *Postgres) UpdateSessionState(ctx context.Context, sessionID string) {
	query := `UPDATE provengo_sessions SET state=200, valid_at= now() + interval '5 minutes' WHERE session_link = $1 and state=202`
	_, err := pg.db.Exec(ctx, query, sessionID)
	if err != nil {
		log.Printf("unable to insert row: %v", err)
	}
}

func (pg *Postgres) SelectSetup(ctx context.Context) ([]ProvengoValues, error) {
	query := `SELECT identity, value from provengo_setup`

	rows, err := pg.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get data: %w", err)
	}
	defer rows.Close()

	var ProvengoSetup []ProvengoValues

	for rows.Next() {
		var dataLine ProvengoValues
		err := rows.Scan(&dataLine.identity, &dataLine.value)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %w", err)
		}
		ProvengoSetup = append(ProvengoSetup, dataLine)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("row iteration error: %w", rows.Err())
	}

	return ProvengoSetup, nil
}

func (pg *Postgres) SelectBeacons(ctx context.Context, headerHash string) ([]BeaconValues, error) {
	// SelectBeacons TODO: this function must be improved by sql function
	query := `SELECT beacon, created_at from provengo_beacon WHERE md5(concat(ipaddr, values)) = $1 AND created_at >= now() - interval '5 minutes' ORDER BY created_at`

	rows, err := pg.db.Query(ctx, query, headerHash)
	if err != nil {
		return nil, fmt.Errorf("get data: %w", err)
	}
	defer rows.Close()

	var beaconValues []BeaconValues

	for rows.Next() {
		var dataLine BeaconValues
		err := rows.Scan(&dataLine.BeaconId, &dataLine.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %w", err)
		}
		beaconValues = append(beaconValues, dataLine)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("row iteration error: %w", rows.Err())
	}

	return beaconValues, nil
}

func (pg *Postgres) SelectBeaconsToContest(ctx context.Context, strLogin string, strSha256 string) ([]BeaconValuesList, error) {
	query := `SELECT _beacon FROM sp_get_beacons($1, $2);`
	rows, err := pg.db.Query(ctx, query, strLogin, strSha256)
	if err != nil {
		return nil, fmt.Errorf("get data: %w", err)
	}
	defer rows.Close()
	var beaconValues []BeaconValuesList
	for rows.Next() {
		var dataLine BeaconValuesList
		err := rows.Scan(&dataLine.BeaconId)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %w", err)
		}
		beaconValues = append(beaconValues, dataLine)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("row iteration error: %w", rows.Err())
	}
	return beaconValues, nil
}

func (pg *Postgres) SelectUserFromUserName(ctx context.Context, UserHashPassword string) (*DefineUser, error) {
	var dataLine DefineUser
	query := `SELECT name, username, telephone, secret, userlevel from provengo_login where username = $1`
	rows, err := pg.db.Query(ctx, query, UserHashPassword)
	if err != nil {
		return &dataLine, fmt.Errorf("get data: %w", err)
	}
	defer rows.Close()
	isUser := false
	for rows.Next() {
		err := rows.Scan(&dataLine.Name, &dataLine.UserName, &dataLine.Telephone, &dataLine.Secret, &dataLine.UserLevel)
		if err != nil {
			return &dataLine, fmt.Errorf("unable to scan row: %w", err)
		}
		isUser = true
	}
	if rows.Err() != nil {
		return &dataLine, fmt.Errorf("row iteration error: %w", rows.Err())
	}
	if isUser != true {
		return &dataLine, fmt.Errorf("unable to find user")
	}
	return &dataLine, nil
}

func (pg *Postgres) SelectUserIdFromUserName(ctx context.Context, UserName string) (int, error) {
	query := `SELECT id from provengo_login where username = $1`
	rows, err := pg.db.Query(ctx, query, UserName)
	if err != nil {
		return -1, fmt.Errorf("get data: %w", err)
	}
	UserID := -1
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&UserID)
		if err != nil {
			return -1, fmt.Errorf("unable to scan row: %w", err)
		}
	}
	if rows.Err() != nil {
		return UserID, fmt.Errorf("row iteration error: %w", rows.Err())
	}

	return UserID, nil
}

func (pg *Postgres) SelectUserLink(ctx context.Context, link string) ([]UserLink, error) {
	query := `SELECT username, link from provengo_userlink where link = $1 and state=202`

	rows, err := pg.db.Query(ctx, query, link)
	if err != nil {
		return nil, fmt.Errorf("get data: %w", err)
	}
	defer rows.Close()

	var DataReturn []UserLink
	for rows.Next() {
		var dataLine UserLink
		err := rows.Scan(&dataLine.Username, &dataLine.Link)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %w", err)
		}
		DataReturn = append(DataReturn, dataLine)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("row iteration error: %w", rows.Err())
	}
	queryUpdate := `UPDATE provengo_userlink SET state=300, valid_at = now() + interval '1 day' WHERE link = $1`
	_, err = pg.db.Exec(ctx, queryUpdate, link)
	if err != nil {
		log.Printf("unable update  row: %v", err)
	}
	return DataReturn, nil
}

func (pg *Postgres) UpdateInsertDomain(ctx context.Context, data DefineDomain) error {
	if data.DomainId == 0 {
		queryInsert := `INSERT INTO provengo_domain values (default, $1, $2, $3, $4, now(),now());`
		res, err := pg.db.Exec(ctx, queryInsert, data.Domain, data.Description, data.UserID, data.Public)
		if err != nil {
			return fmt.Errorf("unable to insert domain: %w", err)
		}
		affectedRows := res.RowsAffected()
		if affectedRows == 0 {
			return fmt.Errorf("unable to insert domain: %w", err)
		}
	} else {
		queryUpdate := `UPDATE provengo_domain SET domain= $3, description = $4, public = $5, updated_at = now() WHERE user_id = $2 and id = $1 and public <= 100;`
		res, err := pg.db.Exec(ctx, queryUpdate, data.DomainId, data.UserID, data.Domain, data.Description, data.Public)
		if err != nil {
			return fmt.Errorf("unable to update domain: %w", err)
		}
		affectedRows := res.RowsAffected()
		if affectedRows == 0 {
			return fmt.Errorf("this update has return no lines: %w", err)
		}
	}
	return nil
}

func (pg *Postgres) UpdateInsertUser(ctx context.Context, users DefineUser, domains []uint64) (string, error) {
	//query := `SELECT values from provengo_sessions where application_id = $1 and state=200`

	userId, err := pg.SelectUserIdFromUserName(ctx, users.UserName)

	if err != nil {
		return "", err
	}

	if userId == -1 {

	}

	applicationID := 10

	query := `INSERT provengo_users SET valid_at = NOW() + INTERVAL '2 minutes' WHERE application_id = $1 and state = 200 RETURNING values`

	rows, err := pg.db.Query(ctx, query, applicationID)
	if err != nil {
		return "", fmt.Errorf("get data: %w", err)
	}
	defer rows.Close()

	var dataLine string
	for rows.Next() {

		err := rows.Scan(&dataLine)
		if err != nil {
			return "", fmt.Errorf("unable to scan row: %w", err)
		}
	}

	if rows.Err() != nil {
		return "", fmt.Errorf("row iteration error: %w", rows.Err())
	}
	return dataLine, nil
}

func (pg *Postgres) UpdateInsertProbe(ctx context.Context, data DefineProbe) error {

	if data.ProbeId == 0 {
		queryInsert := `INSERT INTO provengo_probe values (default, $1, $2, $3, $4, $5, $6, now(), now());`
		res, err := pg.db.Exec(ctx, queryInsert, data.DomainID, data.Probe, data.Description, data.Configuration, data.Status, data.Identity)
		if err != nil {
			log.Printf("unable to insert probe: %v", err)
			return fmt.Errorf("unable to insert probe: %w", err)
		}
		affectedRows := res.RowsAffected()
		if affectedRows == 0 {
			return fmt.Errorf("unable to insert probe: %w COUNT", err)
		}
	} else {
		queryUpdate := `UPDATE provengo_probe SET domain_id= $2, probe = $3, description = $4, configuration = $5,status = $6, updated_at = now() WHERE id = $1 and status <= 100;`
		res, err := pg.db.Exec(ctx, queryUpdate, data.ProbeId, data.DomainID, data.Probe, data.Description, data.Configuration, data.Status)
		if err != nil {
			return fmt.Errorf("unable to update probe: %w", err)
		}
		affectedRows := res.RowsAffected()
		if affectedRows == 0 {
			return fmt.Errorf("this update has return no lines: %w", err)
		}
	}
	return nil
}

func (pg *Postgres) InsertCollector(ctx context.Context, data DefineCollector) error {

	queryInsert := `INSERT INTO provengo_collector values ($1, $2, $3, $4, $5, $6, $7);`
	res, err := pg.db.Exec(ctx, queryInsert, data.HostName, data.HostID, data.IpAddr, data.ProbeId, data.Values, data.RemoteTS, data.LocalTS)

	if err != nil {
		log.Printf("unable to insert probe: %v", err)
		return fmt.Errorf("unable to insert probe: %w", err)
	}
	affectedRows := res.RowsAffected()
	if affectedRows == 0 {
		return fmt.Errorf("unable to insert probe: %w COUNT", err)
	}
	return nil
}

func (pg *Postgres) SelectDomain(ctx context.Context, UserId int) ([]DefineDomain, error) {
	query := `SELECT id,domain,description,public,user_id from provengo_domain where user_id = $1 and public <= 100`

	rows, err := pg.db.Query(ctx, query, UserId)
	if err != nil {
		return nil, fmt.Errorf("get data: %w", err)
	}
	defer rows.Close()

	var DataReturn []DefineDomain
	for rows.Next() {
		var dataLine DefineDomain
		err := rows.Scan(&dataLine.DomainId, &dataLine.Domain, &dataLine.Description, &dataLine.Public, &dataLine.UserID)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %w", err)
		}
		DataReturn = append(DataReturn, dataLine)
	}
	return DataReturn, nil
}

func (pg *Postgres) SelectProbe(ctx context.Context, UserId int) ([]DefineProbe, error) {
	query := `SELECT id,domain_id,probe,description,configuration,status,identity from provengo_probe where domain_id in (select id from provengo_domain where user_id = $1 and public <=100) and status <= 100 ORDER BY ID`

	rows, err := pg.db.Query(ctx, query, UserId)
	if err != nil {
		return nil, fmt.Errorf("get data: %w", err)
	}
	defer rows.Close()

	var DataReturn []DefineProbe
	for rows.Next() {
		var dataLine DefineProbe
		err := rows.Scan(
			&dataLine.ProbeId,
			&dataLine.DomainID,
			&dataLine.Probe,
			&dataLine.Description,
			&dataLine.Configuration,
			&dataLine.Status,
			&dataLine.Identity,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %w", err)
		}
		DataReturn = append(DataReturn, dataLine)
	}
	return DataReturn, nil
}

func (pg *Postgres) ExpireActivity(ctx context.Context, Activity string) {

	var query string

	switch Activity {
	case "sessions":
		query = `SELECT sp_expire_sessions()`
	case "logs":
		query = `SELECT sp_delete_expired_logs()`
	case "beacon":
		query = `SELECT sp_delete_expired_beacons()`
	case "domain":
		query = `UPDATE provengo_domain SET domain = concat(gen_random_uuid(),'-----', domain), updated_at=now(), public=201 WHERE public = 101`
	case "probe":
		query = `UPDATE provengo_probe SET probe = concat(gen_random_uuid(),'-----', probe), updated_at=now(), status=201 WHERE status = 101`
	default:
		query = `SELECT 1 as total`
	}

	_, err := pg.db.Exec(ctx, query)

	if err != nil {
		log.Printf("unable to get rows: %v", err)
	}
}
