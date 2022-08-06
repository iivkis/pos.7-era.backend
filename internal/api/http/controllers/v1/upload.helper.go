package controller

type uploadAllowedContentType map[string]struct{}

func (u uploadAllowedContentType) Add(contentType string) {
	u[contentType] = struct{}{}
}

func (u uploadAllowedContentType) Allowed(contentType string) bool {
	_, ok := u[contentType]
	return ok
}
