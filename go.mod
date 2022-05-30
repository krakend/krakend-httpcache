module github.com/devopsfaith/krakend-httpcache/v2

go 1.17

require (
	github.com/luraproject/lura/v2 v2.0.0
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
)

require (
	github.com/krakendio/flatmap v0.0.0-20200601181759-8521186182fc // indirect
	github.com/valyala/fastrand v1.1.0 // indirect
)

replace github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 => github.com/m4ns0ur/httpcache v0.0.0-20200426190423-1040e2e8823f
