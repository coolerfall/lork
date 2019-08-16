module gitlab.com/anbillon/slago/zerolog-to-slago

go 1.12

require (
	github.com/rs/zerolog v1.15.0
	gitlab.com/anbillon/slago/slago-api v0.0.0-00010101000000-000000000000
)

replace gitlab.com/anbillon/slago/slago-api => ../slago-api
