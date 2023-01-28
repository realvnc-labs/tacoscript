package tasks

const (
	TaskTypeCmdRun  = "cmd.run"
	FileManaged     = "file.managed"
	FileReplace     = "file.replace"
	PkgInstalled    = "pkg.installed"
	PkgRemoved      = "pkg.removed"
	PkgUpgraded     = "pkg.uptodate"
	WinRegPresent   = "win_reg.present"
	WinRegAbsent    = "win_reg.absent"
	WinRegAbsentKey = "win_reg.absent_key"

	NameField  = "name"
	NamesField = "names"

	RequireField = "require"

	CreatesField = "creates"
	OnlyIfField  = "onlyif"
	UnlessField  = "unless"

	PatternField = "pattern"
	ReplField    = "repl"

	CwdField          = "cwd"
	UserField         = "user"
	ShellField        = "shell"
	EnvField          = "env"
	SourceField       = "source"
	SourceHashField   = "source_hash"
	MakeDirsField     = "makedirs"
	ReplaceField      = "replace"
	SkipVerifyField   = "skip_verify"
	ContentsField     = "contents"
	GroupField        = "group"
	ModeField         = "mode"
	EncodingField     = "encoding"
	AbortOnErrorField = "abort_on_error"
	Version           = "version"
	Refresh           = "refresh"

	CountField             = "count"
	AppendIfNotFoundField  = "append_if_not_found"
	PrependIfNotFoundField = "prepend_if_not_found"
	NotFoundContentField   = "not_found_content"
	BackupExtensionField   = "backup"
	MaxFileSizeField       = "max_file_size"

	RegPathField = "reg_path"
	ValField     = "value"
	ValTypeField = "type"
)
