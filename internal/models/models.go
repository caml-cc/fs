package models

type Config struct {
	PORT        string
	API_KEY     string
	MAX_STORAGE int
	MAX_UPLOAD  int64 // bytes
	KEEP        int
	UPLOAD_DIR  string
}
