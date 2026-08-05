package config

type KCPConfig struct {
	MTU      int `json:"mtu"`
	NoDelay  int `json:"no_delay"`
	Interval int `json:"interval"`
	Resend   int `json:"resend"`
	NC       int `json:"nc"`

	SndWnd int `json:"snd_wnd"`
	RcvWnd int `json:"rcv_wnd"`
}
