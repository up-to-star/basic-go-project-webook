package web

type jwtHandler struct {
	// access_token
	atKey []byte
	// refresh_token
	rtKey []byte
}

func newJwtHandler() jwtHandler {
	return jwtHandler{
		atKey: []byte("BTv_D7]5q+f)9MTLwAA'5N!PJ6d6PNQQ"),
		rtKey: []byte("BTv_D7]5q+f)9MTLwAA'5N!PJ6d6xyad"),
	}
}
