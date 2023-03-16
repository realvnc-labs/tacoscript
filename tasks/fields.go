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
	RealVNCServer   = "realvnc_server.config_update"

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

	EncryptionField              = "encryption"
	AuthenticationField          = "authentication"
	PermissionsField             = "permissions"
	QueryConnectField            = "query_connect"
	QueryOnlyIfLoggedOnField     = "query_only_if_logged_on"
	QueryConnectTimeoutSecsField = "query_connect_timeout"
	BlankScreenField             = "blank_screen"
	ConnNotifyTimeoutSecsField   = "conn_notify_timeout"
	ConnNotifyAlwaysField        = "conn_notify_always"
	IdleTimeoutSecsField         = "idle_timeout"
	LogField                     = "log"
	CaptureMethodField           = "capture_method"

	ConfigFileField = "config_file"
	ServerModeField = "server_mode"
	ExecPathField   = "exec_path"
	ExecCmdField    = "exec_cmd"
	SkipReloadField = "skip_reload"
	SkipBackupField = "skip_backup"
)

var (
	sharedFields = []string{"name", "names", "require", "creates", "onlyif", "unless", "shell"}
)

func sharedField(fieldKey string) (shared bool) {
	for _, f := range sharedFields {
		if fieldKey == f {
			return true
		}
	}
	return false
}
