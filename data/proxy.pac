function FindProxyForURL(url, host) {
	if (host === "ket")
		return "PROXY localhost:8080"
	//return "PROXY proxy.example.com:8080; DIRECT";
	return "DIRECT"
}