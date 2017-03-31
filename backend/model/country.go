package model

type Country struct {
	ID   int64  `meddler:"id,pk"`
	Name string `meddler:"name"`
}
