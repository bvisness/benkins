package shared

type ProjectName string

func NewProjectNameFromPlain(name string) ProjectName {
	return ProjectName(Base64Encode(name))
}

func NewProjectNameFromEncoded(encodedName string) ProjectName {
	return ProjectName(encodedName)
}

func (p ProjectName) Encoded() string {
	return string(p)
}

func (p ProjectName) Decoded() string {
	return Base64Decode(string(p))
}
