package main

type Config struct {
	Dir            string
	DbFileName     string
	AppendOnly     string
	AppendDirname  string
	AppendFilename string
	AppendFSync    string
}

func NewConfig(dir, dbfilename, appendonly, appenddirname, appendfilename, appendfsync *string) *Config {
	config := Config{
		Dir:            DIR,
		DbFileName:     "",
		AppendOnly:     "no",
		AppendDirname:  "appendonlydir",
		AppendFilename: "appendonly.aof",
		AppendFSync:    "everysec",
	}

	if dir != nil && *dir != "" {
		config.Dir = *dir
	}

	if dbfilename != nil && *dbfilename != "" {
		config.DbFileName = *dbfilename
	}

	if appendonly != nil && *appendonly != "" {
		config.AppendOnly = *appendonly
	}

	if appenddirname != nil && *appenddirname != "" {
		config.AppendDirname = *appenddirname
	}

	if appendfilename != nil && *appendfilename != "" {
		config.AppendFilename = *appendfilename
	}

	if appendfsync != nil && *appendfsync != "" {
		config.AppendFSync = *appendfsync
	}

	return &config
}
