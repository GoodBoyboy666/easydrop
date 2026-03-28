package token

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	red "github.com/redis/go-redis/v9"
)

func mustNewConcreteManager(t *testing.T, cfg *Config, redisClient *red.Client) *manager {
	t.Helper()

	tokenManager, err := NewManager(cfg, redisClient)
	if err != nil {
		t.Fatalf("创建 token manager 失败: %v", err)
	}

	concrete, ok := tokenManager.(*manager)
	if !ok {
		t.Fatal("unexpected token manager implementation")
	}

	return concrete
}

func TestNewManager_BackendSelection(t *testing.T) {
	t.Parallel()

	manager, err := NewManager(nil, nil)
	if err != nil {
		t.Fatalf("创建内存 token manager 失败: %v", err)
	}
	if manager.Backend() != BackendMemory {
		t.Fatalf("后端不符合预期: %s", manager.Backend())
	}

	client := red.NewClient(&red.Options{Addr: "127.0.0.1:6379"})
	t.Cleanup(func() { _ = client.Close() })

	manager, err = NewManager(&Config{KeyPrefix: "custom"}, client)
	if err != nil {
		t.Fatalf("创建 redis token manager 失败: %v", err)
	}
	if manager.Backend() != BackendRedis {
		t.Fatalf("后端不符合预期: %s", manager.Backend())
	}
}

func TestManagerIssueAndConsume_Memory(t *testing.T) {
	t.Parallel()

	manager := mustNewConcreteManager(t, nil, nil)

	recordedNow := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return recordedNow }

	payload := "{\"email\":\"alice@example.com\",\"meta\":{\"source\":\"test\"}}"

	tokenValue, err := manager.Issue(context.Background(), 42, KindVerifyEmail, 5*time.Minute, payload)
	if err != nil {
		t.Fatalf("签发 token 失败: %v", err)
	}
	if tokenValue == "" {
		t.Fatal("期望生成非空 token")
	}

	record, err := manager.Consume(context.Background(), KindVerifyEmail, tokenValue)
	if err != nil {
		t.Fatalf("消费 token 失败: %v", err)
	}
	if record.UserID != 42 {
		t.Fatalf("user id 不符合预期: %d", record.UserID)
	}
	if record.Kind != KindVerifyEmail {
		t.Fatalf("kind 不符合预期: %s", record.Kind)
	}
	if record.Token != tokenValue {
		t.Fatalf("token 不符合预期: %s", record.Token)
	}
	if record.CreatedAt != recordedNow {
		t.Fatalf("created_at 不符合预期: %s", record.CreatedAt)
	}
	if want := recordedNow.Add(5 * time.Minute); !record.ExpiresAt.Equal(want) {
		t.Fatalf("expires_at 不符合预期: %s", record.ExpiresAt)
	}

	if record.Payload != payload {
		t.Fatalf("payload 不符合预期: %q", record.Payload)
	}

	_, err = manager.Consume(context.Background(), KindVerifyEmail, tokenValue)
	if !errors.Is(err, ErrTokenNotFound) {
		t.Fatalf("期望消费后 token 不存在，实际为: %v", err)
	}
}

func TestManagerIssue_ReplacesExistingTokenOfSameKind(t *testing.T) {
	t.Parallel()

	manager := mustNewConcreteManager(t, nil, nil)

	firstToken, err := manager.Issue(context.Background(), 7, KindResetPassword, time.Minute, "{\"step\":1}")
	if err != nil {
		t.Fatalf("签发第一个 token 失败: %v", err)
	}

	secondToken, err := manager.Issue(context.Background(), 7, KindResetPassword, time.Minute, "{\"step\":2}")
	if err != nil {
		t.Fatalf("签发第二个 token 失败: %v", err)
	}
	if secondToken == firstToken {
		t.Fatal("期望新 token 与旧 token 不同")
	}

	_, err = manager.Consume(context.Background(), KindResetPassword, firstToken)
	if !errors.Is(err, ErrTokenNotFound) {
		t.Fatalf("期望旧 token 失效，实际为: %v", err)
	}

	record, err := manager.Consume(context.Background(), KindResetPassword, secondToken)
	if err != nil {
		t.Fatalf("消费新 token 失败: %v", err)
	}
	if record.Payload != "{\"step\":2}" {
		t.Fatalf("新 token payload 不符合预期: %q", record.Payload)
	}
}

func TestManagerIssue_AllowsDifferentKinds(t *testing.T) {
	t.Parallel()

	manager, err := NewManager(nil, nil)
	if err != nil {
		t.Fatalf("创建 token manager 失败: %v", err)
	}

	resetToken, err := manager.Issue(context.Background(), 9, KindResetPassword, time.Minute, "reset")
	if err != nil {
		t.Fatalf("签发重置密码 token 失败: %v", err)
	}

	verifyToken, err := manager.Issue(context.Background(), 9, KindVerifyEmail, time.Minute, "verify")
	if err != nil {
		t.Fatalf("签发邮箱验证 token 失败: %v", err)
	}

	resetRecord, err := manager.Consume(context.Background(), KindResetPassword, resetToken)
	if err != nil {
		t.Fatalf("消费重置密码 token 失败: %v", err)
	}
	if resetRecord.Payload != "reset" {
		t.Fatalf("重置密码 token payload 不符合预期: %q", resetRecord.Payload)
	}

	verifyRecord, err := manager.Consume(context.Background(), KindVerifyEmail, verifyToken)
	if err != nil {
		t.Fatalf("消费邮箱验证 token 失败: %v", err)
	}
	if verifyRecord.Payload != "verify" {
		t.Fatalf("邮箱验证 token payload 不符合预期: %q", verifyRecord.Payload)
	}
}

func TestManagerConsume_ExpiredToken(t *testing.T) {
	t.Parallel()

	manager := mustNewConcreteManager(t, nil, nil)

	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return now }

	tokenValue, err := manager.Issue(context.Background(), 12, KindChangeEmail, time.Minute, "")
	if err != nil {
		t.Fatalf("签发 token 失败: %v", err)
	}

	manager.now = func() time.Time { return now.Add(2 * time.Minute) }

	_, err = manager.Consume(context.Background(), KindChangeEmail, tokenValue)
	if !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("期望过期错误，实际为: %v", err)
	}
}

func TestManagerIssueAndConsume_EmptyPayload(t *testing.T) {
	t.Parallel()

	manager := mustNewConcreteManager(t, nil, nil)

	tokenValue, err := manager.Issue(context.Background(), 13, KindChangeEmail, time.Minute, "")
	if err != nil {
		t.Fatalf("签发空 payload token 失败: %v", err)
	}

	record, err := manager.Consume(context.Background(), KindChangeEmail, tokenValue)
	if err != nil {
		t.Fatalf("消费空 payload token 失败: %v", err)
	}
	if record.Payload != "" {
		t.Fatalf("空 payload 不符合预期: %q", record.Payload)
	}
}

func TestRedisStoreSaveAndConsume(t *testing.T) {
	t.Parallel()

	client := newFakeRedisTxnClient()
	store := newRedisStore(client, "unit-token")
	store.now = func() time.Time {
		return time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	}

	record := &Record{
		UserID:    20,
		Kind:      KindResetPassword,
		Token:     "token-a",
		CreatedAt: time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 20, 10, 10, 0, 0, time.UTC),
		Payload:   "alice@example.com",
	}

	if err := store.Save(context.Background(), record); err != nil {
		t.Fatalf("写入 redis token 失败: %v", err)
	}

	recordKey := store.recordKey(record.Kind, record.Token)
	userKey := store.userKey(record.UserID, record.Kind)
	if client.values[userKey] != record.Token {
		t.Fatalf("用户索引不符合预期: %s", client.values[userKey])
	}
	if client.ttls[recordKey] <= 0 || client.ttls[userKey] <= 0 {
		t.Fatalf("ttl 未写入: record=%s user=%s", client.ttls[recordKey], client.ttls[userKey])
	}

	consumed, err := store.Consume(context.Background(), record.Kind, record.Token, record.CreatedAt.Add(time.Minute))
	if err != nil {
		t.Fatalf("消费 redis token 失败: %v", err)
	}
	if consumed.Payload != "alice@example.com" {
		t.Fatalf("payload 不符合预期: %q", consumed.Payload)
	}
	if _, ok := client.values[recordKey]; ok {
		t.Fatalf("消费后 record key 未删除: %s", recordKey)
	}
}

func TestRedisStoreSave_ReplacesExistingRecord(t *testing.T) {
	t.Parallel()

	client := newFakeRedisTxnClient()
	store := newRedisStore(client, "unit-token")
	store.now = func() time.Time {
		return time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	}

	first := &Record{
		UserID:    30,
		Kind:      KindVerifyEmail,
		Token:     "token-old",
		CreatedAt: time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 20, 10, 5, 0, 0, time.UTC),
	}
	if err := store.Save(context.Background(), first); err != nil {
		t.Fatalf("写入首个 token 失败: %v", err)
	}

	second := &Record{
		UserID:    30,
		Kind:      KindVerifyEmail,
		Token:     "token-new",
		CreatedAt: time.Date(2026, 3, 20, 10, 1, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 20, 10, 6, 0, 0, time.UTC),
	}
	if err := store.Save(context.Background(), second); err != nil {
		t.Fatalf("写入替换 token 失败: %v", err)
	}

	if _, ok := client.values[store.recordKey(first.Kind, first.Token)]; ok {
		t.Fatalf("旧 token 记录未删除")
	}
	if client.values[store.userKey(second.UserID, second.Kind)] != second.Token {
		t.Fatalf("用户索引未更新")
	}
}

func TestRedisStoreSave_RetriesOnTransactionConflict(t *testing.T) {
	t.Parallel()

	client := newFakeRedisTxnClient()
	client.watchFailures = 1
	store := newRedisStore(client, "unit-token")
	store.now = func() time.Time {
		return time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	}

	record := &Record{
		UserID:    40,
		Kind:      KindResetPassword,
		Token:     "token-retry",
		CreatedAt: time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 20, 10, 5, 0, 0, time.UTC),
	}

	if err := store.Save(context.Background(), record); err != nil {
		t.Fatalf("事务重试后写入 token 失败: %v", err)
	}
	if client.watchCalls != 2 {
		t.Fatalf("watch 次数不符合预期: %d", client.watchCalls)
	}
}

func TestRedisStoreConsume_RetriesOnTransactionConflict(t *testing.T) {
	t.Parallel()

	client := newFakeRedisTxnClient()
	store := newRedisStore(client, "unit-token")
	store.now = func() time.Time {
		return time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	}

	record := &Record{
		UserID:    41,
		Kind:      KindVerifyEmail,
		Token:     "token-retry-consume",
		CreatedAt: time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 20, 10, 5, 0, 0, time.UTC),
	}
	if err := store.Save(context.Background(), record); err != nil {
		t.Fatalf("预写入 token 失败: %v", err)
	}

	client.watchFailures = 1
	consumed, err := store.Consume(context.Background(), record.Kind, record.Token, record.CreatedAt.Add(time.Minute))
	if err != nil {
		t.Fatalf("事务重试后消费 token 失败: %v", err)
	}
	if consumed.Token != record.Token {
		t.Fatalf("消费结果不符合预期: %s", consumed.Token)
	}
}

func TestRedisStoreConsume_ExpiredTokenDeletesKeys(t *testing.T) {
	t.Parallel()

	client := newFakeRedisTxnClient()
	store := newRedisStore(client, "unit-token")
	store.now = func() time.Time {
		return time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	}

	record := &Record{
		UserID:    42,
		Kind:      KindChangeEmail,
		Token:     "expired-token",
		CreatedAt: time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 20, 10, 1, 0, 0, time.UTC),
	}
	if err := store.Save(context.Background(), record); err != nil {
		t.Fatalf("预写入 token 失败: %v", err)
	}

	_, err := store.Consume(context.Background(), record.Kind, record.Token, record.CreatedAt.Add(2*time.Minute))
	if !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("期望过期错误，实际为: %v", err)
	}
	if _, ok := client.values[store.recordKey(record.Kind, record.Token)]; ok {
		t.Fatalf("过期 token 的 record key 未删除")
	}
}

type fakeRedisTxnClient struct {
	mu            sync.Mutex
	watchCalls    int
	watchFailures int
	values        map[string]string
	ttls          map[string]time.Duration
}

type fakeRedisTx struct {
	client      *fakeRedisTxnClient
	pending     []func()
	hasPipeline bool
}

type fakeRedisPipe struct {
	tx *fakeRedisTx
}

func newFakeRedisTxnClient() *fakeRedisTxnClient {
	return &fakeRedisTxnClient{
		values: make(map[string]string),
		ttls:   make(map[string]time.Duration),
	}
}

func (f *fakeRedisTxnClient) Watch(_ context.Context, fn func(redisTx) error, _ ...string) error {
	f.mu.Lock()
	f.watchCalls++
	f.mu.Unlock()

	tx := &fakeRedisTx{client: f}
	err := fn(tx)
	if !tx.hasPipeline {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.watchFailures > 0 {
		f.watchFailures--
		return red.TxFailedErr
	}
	for _, op := range tx.pending {
		op()
	}
	return err
}

func (tx *fakeRedisTx) Get(_ context.Context, key string) (string, bool, error) {
	tx.client.mu.Lock()
	defer tx.client.mu.Unlock()

	value, ok := tx.client.values[key]
	return value, ok, nil
}

func (tx *fakeRedisTx) TxPipelined(_ context.Context, fn func(redisPipe) error) error {
	tx.hasPipeline = true
	pipe := &fakeRedisPipe{tx: tx}
	return fn(pipe)
}

func (p *fakeRedisPipe) Set(_ context.Context, key, value string, ttl time.Duration) {
	p.tx.pending = append(p.tx.pending, func() {
		p.tx.client.values[key] = value
		p.tx.client.ttls[key] = ttl
	})
}

func (p *fakeRedisPipe) Del(_ context.Context, keys ...string) {
	p.tx.pending = append(p.tx.pending, func() {
		for _, key := range keys {
			delete(p.tx.client.values, key)
			delete(p.tx.client.ttls, key)
		}
	})
}

func TestRedisStoreSave_ReturnsConflictAfterRetryExhausted(t *testing.T) {
	t.Parallel()

	client := newFakeRedisTxnClient()
	client.watchFailures = redisTxnMaxRetries
	store := newRedisStore(client, "unit-token")
	store.now = func() time.Time {
		return time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	}

	record := &Record{
		UserID:    43,
		Kind:      KindVerifyEmail,
		Token:     "token-conflict",
		CreatedAt: time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 20, 10, 5, 0, 0, time.UTC),
	}

	err := store.Save(context.Background(), record)
	if !errors.Is(err, ErrTransactionConflict) {
		t.Fatalf("期望事务冲突错误，实际为: %v", err)
	}
	if client.watchCalls != redisTxnMaxRetries {
		t.Fatalf("watch 次数不符合预期: %d", client.watchCalls)
	}
}

func TestRedisStoreConsume_ReturnsConflictAfterRetryExhausted(t *testing.T) {
	t.Parallel()

	client := newFakeRedisTxnClient()
	store := newRedisStore(client, "unit-token")
	store.now = func() time.Time {
		return time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	}

	record := &Record{
		UserID:    44,
		Kind:      KindVerifyEmail,
		Token:     "token-consume-conflict",
		CreatedAt: time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 20, 10, 5, 0, 0, time.UTC),
	}
	if err := store.Save(context.Background(), record); err != nil {
		t.Fatalf("预写入 token 失败: %v", err)
	}

	client.watchFailures = redisTxnMaxRetries
	_, err := store.Consume(context.Background(), record.Kind, record.Token, record.CreatedAt.Add(time.Minute))
	if !errors.Is(err, ErrTransactionConflict) {
		t.Fatalf("期望事务冲突错误，实际为: %v", err)
	}
	if client.watchCalls != 1+redisTxnMaxRetries {
		t.Fatalf("watch 次数不符合预期: %d", client.watchCalls)
	}
}

func TestRedisStoreConsume_IgnoresDanglingUserKeyWhenRecordMissing(t *testing.T) {
	t.Parallel()

	client := newFakeRedisTxnClient()
	store := newRedisStore(client, "unit-token")

	userKey := store.userKey(55, KindResetPassword)
	client.values[userKey] = "dangling-token"
	client.ttls[userKey] = time.Minute

	_, err := store.Consume(context.Background(), KindResetPassword, "dangling-token", time.Date(2026, 3, 20, 10, 1, 0, 0, time.UTC))
	if !errors.Is(err, ErrTokenNotFound) {
		t.Fatalf("期望 token 不存在，实际为: %v", err)
	}
	if _, ok := client.values[userKey]; !ok {
		t.Fatalf("悬挂 user key 不应被当前消费逻辑修改")
	}
}

func TestFakeRedisTxnClient_DoesNotApplyPendingOpsOnConflict(t *testing.T) {
	t.Parallel()

	client := newFakeRedisTxnClient()
	client.watchFailures = 1

	err := client.Watch(context.Background(), func(tx redisTx) error {
		return tx.TxPipelined(context.Background(), func(pipe redisPipe) error {
			pipe.Set(context.Background(), "k", "v", time.Minute)
			return nil
		})
	}, "k")
	if !errors.Is(err, red.TxFailedErr) {
		t.Fatalf("期望事务冲突，实际为: %v", err)
	}
	if _, ok := client.values["k"]; ok {
		t.Fatalf("冲突时不应提交事务内容")
	}
}
