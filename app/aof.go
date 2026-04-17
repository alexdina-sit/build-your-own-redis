package main

type AOF struct {
	Dir            string
	AppendOnly     string
	AppendDirName  string
	AppendFileName string
	AppendFSync    string
}

func NewAOF(dir, appendOnly, appendDirName, appendFileName, appendFSync string) *AOF {
	return &AOF{
		Dir:            dir,
		AppendOnly:     appendOnly,
		AppendDirName:  appendDirName,
		AppendFileName: appendFileName,
		AppendFSync:    appendFSync,
	}
}
