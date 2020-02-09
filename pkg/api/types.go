package api

type Log struct {
	Pod       string
	Container string
	Content   []byte
}
