package token

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	red "github.com/redis/go-redis/v9"
)

const redisTxnMaxRetries = 3

var ErrTransactionConflict = errors.New("redis token 事务冲突")

type redisTxnClient interface {
	Watch(ctx context.Context, fn func(redisTx) error, keys ...string) error
}

type redisTx interface {
	Get(ctx context.Context, key string) (string, bool, error)
	TxPipelined(ctx context.Context, fn func(redisPipe) error) error
}

type redisPipe interface {
	Set(ctx context.Context, key, value string, ttl time.Duration)
	Del(ctx context.Context, keys ...string)
}

type goRedisTxnClient struct {
	client *red.Client
}

type redisStore struct {
	client    redisTxnClient
	keyPrefix string
	now       func() time.Time
}

type goRedisTx struct {
	tx *red.Tx
}

type goRedisPipe struct {
	pipe red.Pipeliner
}

func newGoRedisTxnClient(client *red.Client) *goRedisTxnClient {
	return &goRedisTxnClient{client: client}
}

func newRedisStore(client redisTxnClient, keyPrefix string) *redisStore {
	return &redisStore{
		client:    client,
		keyPrefix: keyPrefix,
		now:       time.Now,
	}
}

func (c *goRedisTxnClient) Watch(ctx context.Context, fn func(redisTx) error, keys ...string) error {
	return c.client.Watch(ctx, func(tx *red.Tx) error {
		return fn(&goRedisTx{tx: tx})
	}, keys...)
}

func (tx *goRedisTx) Get(ctx context.Context, key string) (string, bool, error) {
	value, err := tx.tx.Get(ctx, key).Result()
	if err == nil {
		return value, true, nil
	}
	if err == red.Nil {
		return "", false, nil
	}
	return "", false, err
}

func (tx *goRedisTx) TxPipelined(ctx context.Context, fn func(redisPipe) error) error {
	_, err := tx.tx.TxPipelined(ctx, func(pipe red.Pipeliner) error {
		return fn(&goRedisPipe{pipe: pipe})
	})
	return err
}

func (p *goRedisPipe) Set(ctx context.Context, key, value string, ttl time.Duration) {
	p.pipe.Set(ctx, key, value, ttl)
}

func (p *goRedisPipe) Del(ctx context.Context, keys ...string) {
	if len(keys) == 0 {
		return
	}
	p.pipe.Del(ctx, keys...)
}

func (s *redisStore) Save(ctx context.Context, record *Record) error {
	ttl := record.ExpiresAt.Sub(s.now().UTC())
	if ttl <= 0 {
		return ErrInvalidTTL
	}

	serialized, err := marshalRecord(record)
	if err != nil {
		return err
	}

	userKey := s.userKey(record.UserID, record.Kind)
	recordKey := s.recordKey(record.Kind, record.Token)
	for attempt := 0; attempt < redisTxnMaxRetries; attempt++ {
		err = s.client.Watch(ctx, func(tx redisTx) error {
			oldToken, found, err := tx.Get(ctx, userKey)
			if err != nil {
				return fmt.Errorf("读取 token 索引失败: %w", err)
			}

			if err := tx.TxPipelined(ctx, func(pipe redisPipe) error {
				pipe.Set(ctx, recordKey, serialized, ttl)
				pipe.Set(ctx, userKey, record.Token, ttl)
				if found && oldToken != "" && oldToken != record.Token {
					pipe.Del(ctx, s.recordKey(record.Kind, oldToken))
				}
				return nil
			}); err != nil {
				return fmt.Errorf("执行 token 写入事务失败: %w", err)
			}

			return nil
		}, userKey)
		if err == nil {
			return nil
		}
		if !errors.Is(err, red.TxFailedErr) {
			return fmt.Errorf("写入 token 失败: %w", err)
		}
	}

	return fmt.Errorf("写入 token 失败: %w", ErrTransactionConflict)
}

func (s *redisStore) Consume(ctx context.Context, userID uint, kind, tokenValue string, now time.Time) (*Record, error) {
	userKey := s.userKey(userID, kind)
	recordKey := s.recordKey(kind, tokenValue)
	for attempt := 0; attempt < redisTxnMaxRetries; attempt++ {
		var consumed *Record
		err := s.client.Watch(ctx, func(tx redisTx) error {
			activeToken, found, err := tx.Get(ctx, userKey)
			if err != nil {
				return fmt.Errorf("读取 token 索引失败: %w", err)
			}
			if !found {
				return ErrTokenNotFound
			}
			if activeToken != tokenValue {
				return ErrTokenMismatch
			}

			serialized, found, err := tx.Get(ctx, recordKey)
			if err != nil {
				return fmt.Errorf("读取 token 记录失败: %w", err)
			}
			if !found {
				if err := deleteKeysInTransaction(ctx, tx, userKey); err != nil {
					return fmt.Errorf("删除失效 token 索引失败: %w", err)
				}
				return ErrTokenNotFound
			}

			record, err := unmarshalRecord(serialized)
			if err != nil {
				return err
			}
			if record.UserID != userID || record.Kind != kind || record.Token != tokenValue {
				return ErrTokenMismatch
			}
			if !record.ExpiresAt.After(now) {
				if err := deleteKeysInTransaction(ctx, tx, recordKey, userKey); err != nil {
					return fmt.Errorf("删除过期 token 失败: %w", err)
				}
				return ErrTokenExpired
			}

			consumed, err = cloneRecord(record)
			if err != nil {
				return err
			}
			if err := deleteKeysInTransaction(ctx, tx, recordKey, userKey); err != nil {
				return fmt.Errorf("删除已消费 token 失败: %w", err)
			}
			return nil
		}, userKey, recordKey)
		if err == nil {
			return consumed, nil
		}
		if errors.Is(err, red.TxFailedErr) {
			continue
		}
		switch {
		case errors.Is(err, ErrTokenNotFound):
			return nil, ErrTokenNotFound
		case errors.Is(err, ErrTokenExpired):
			return nil, ErrTokenExpired
		case errors.Is(err, ErrTokenMismatch):
			return nil, ErrTokenMismatch
		default:
			return nil, fmt.Errorf("消费 token 失败: %w", err)
		}
	}

	return nil, fmt.Errorf("消费 token 失败: %w", ErrTransactionConflict)
}

func (s *redisStore) recordKey(kind, tokenValue string) string {
	return s.keyPrefix + ":" + kind + ":record:" + tokenValue
}

func (s *redisStore) userKey(userID uint, kind string) string {
	return s.keyPrefix + ":" + kind + ":user:" + strconv.FormatUint(uint64(userID), 10)
}

func marshalRecord(record *Record) (string, error) {
	raw, err := json.Marshal(record)
	if err != nil {
		return "", fmt.Errorf("序列化 token 记录失败: %w", err)
	}
	return string(raw), nil
}

func unmarshalRecord(value string) (*Record, error) {
	record := &Record{}
	if err := json.Unmarshal([]byte(value), record); err != nil {
		return nil, fmt.Errorf("反序列化 token 记录失败: %w", err)
	}
	return record, nil
}

func deleteKeysInTransaction(ctx context.Context, tx redisTx, keys ...string) error {
	return tx.TxPipelined(ctx, func(pipe redisPipe) error {
		pipe.Del(ctx, keys...)
		return nil
	})
}
