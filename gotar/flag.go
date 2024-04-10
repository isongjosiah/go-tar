package gotar

type TarFlag struct {
	ListArchivedContent bool

	// UseFile is set to true when a file path is passed
	UseFile bool

	ExtractFromArchive bool
}
