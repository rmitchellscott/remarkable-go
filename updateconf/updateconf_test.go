package updateconf

import "testing"

func TestSetServer(t *testing.T) {
	const url = "http://127.0.0.1:8090/"

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "existing SERVER under [General]",
			in:   "[General]\nSERVER=https://updates.cloud.remarkable.com/\nPOLL_INTERVAL=3600\n",
			want: "[General]\nSERVER=http://127.0.0.1:8090/\n#SERVER=https://updates.cloud.remarkable.com/\nPOLL_INTERVAL=3600\n",
		},
		{
			name: "no [General] section appends at end",
			in:   "POLL_INTERVAL=3600\n",
			want: "POLL_INTERVAL=3600\nSERVER=http://127.0.0.1:8090/\n",
		},
		{
			name: "multiple SERVER lines all commented",
			in:   "[General]\nSERVER=https://a/\nSERVER=https://b/\n",
			want: "[General]\nSERVER=http://127.0.0.1:8090/\n#SERVER=https://a/\n#SERVER=https://b/\n",
		},
		{
			name: "no trailing newline is preserved",
			in:   "[General]\nSERVER=https://updates.cloud.remarkable.com/",
			want: "[General]\nSERVER=http://127.0.0.1:8090/\n#SERVER=https://updates.cloud.remarkable.com/",
		},
		{
			name: "empty input appends",
			in:   "",
			want: "SERVER=http://127.0.0.1:8090/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(SetServer([]byte(tt.in), url))
			if got != tt.want {
				t.Errorf("SetServer()\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestSetServerRoundTrip(t *testing.T) {
	in := "[General]\nSERVER=https://updates.cloud.remarkable.com/\nPOLL_INTERVAL=3600\n"
	out := SetServer([]byte(in), "http://127.0.0.1:8090/")
	if got := CurrentServer(out); got != "http://127.0.0.1:8090/" {
		t.Errorf("CurrentServer after SetServer = %q, want the new server", got)
	}
}

func TestCurrentServer(t *testing.T) {
	if got := CurrentServer([]byte("[General]\nSERVER=https://x/\n")); got != "https://x/" {
		t.Errorf("CurrentServer = %q, want https://x/", got)
	}
	if got := CurrentServer([]byte("[General]\n#SERVER=https://x/\n")); got != "" {
		t.Errorf("CurrentServer should ignore commented SERVER, got %q", got)
	}
}
