package redis

import "testing"

func TestNewOptionalClient(t *testing.T) {
	t.Parallel()

	client, err := NewOptionalClient(nil)
	if err != nil {
		t.Fatalf("nil 配置时不应报错: %v", err)
	}
	if client != nil {
		t.Fatal("nil 配置时不应创建客户端")
	}

	client, err = NewOptionalClient(&Config{Addr: "  "})
	if err != nil {
		t.Fatalf("空地址时不应报错: %v", err)
	}
	if client != nil {
		t.Fatal("空地址时不应创建客户端")
	}

	client, err = NewOptionalClient(&Config{Addr: "localhost:6379"})
	if err != nil {
		t.Fatalf("有效地址应能创建客户端: %v", err)
	}
	if client == nil {
		t.Fatal("有效地址时应创建客户端")
	}
	t.Cleanup(func() { _ = client.Close() })
}
