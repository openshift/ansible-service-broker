package dao

type DaoConfig struct {
	EtcdHost string
	EtcdPort string
}

type Dao struct {
	config DaoConfig
}
