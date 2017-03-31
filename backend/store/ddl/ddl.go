package ddl

// to install go-bindata
// go get -u github.com/jteeuwen/go-bindata/...

// the following comment instructs go to use go-godata to embed
// binary files into the final executable

//go:generate go-bindata -pkg ddl -o ddl_gen.go sqlite3/
