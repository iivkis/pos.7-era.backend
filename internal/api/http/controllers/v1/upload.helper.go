package controller

// в структуре храняться разрешенные к загрузке на сервер типы файлов (image/jpeg, image/png и т.д.)
type uploadAllowedContentType map[string]struct{}

func (u uploadAllowedContentType) Add(contentType string) {
	u[contentType] = struct{}{}
}

func (u uploadAllowedContentType) Allowed(contentType string) bool {
	_, ok := u[contentType]
	return ok
}
