package projectlint

type Finding struct {
	Path    string
	Line    int
	Column  int
	Message string
}
