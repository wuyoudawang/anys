package storeEngine

type Compaction interface {
	Compact() error
}

type Interface interface {
	MakeRoomForLog() error

	UpLevel() error
	// Lock() (l util.Releaser, err error)

	// Log logs a string. This is used for logging. An implementation
	// may write to a file, stdout or simply do nothing.
	// Log(str string)

	// GetFile returns a file for the given number and type. GetFile will never
	// returns nil, even if the underlying storage is closed.
	// GetFile(num uint64, t FileType) File

	// GetFiles returns a slice of files that match the given file types.
	// The file types may be OR'ed together.
	// GetFiles(t FileType) ([]File, error)

	// GetManifest returns a manifest file. Returns os.ErrNotExist if manifest
	// file does not exist.
	// GetManifest() (File, error)

	// SetManifest sets the given file as manifest file. The given file should
	// be a manifest file type or error will be returned.
	// SetManifest(f File) error

	// Close closes the storage. It is valid to call Close multiple times.
	// Other methods should not be called after the storage has been closed.
	// Close() error
}
