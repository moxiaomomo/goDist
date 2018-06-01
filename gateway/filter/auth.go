package filter

type BaseAuth interface {
	LimitReached() bool
}

type HeaderAuth interface {
	BaseAuth
	IsCrossDomain() bool
}

type CookieAuth interface {
	BaseAuth
	IsCookieValid() bool
}
