package load

import (
	"crypto"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/go-redis/redis/v7"
	"github.com/marmotedu/errors"
	"github.com/nico612/iam-demo/pkg/log"
	"github.com/nico612/iam-demo/pkg/storage"
)

// NotificationCommand defines a new notification type.
type NotificationCommand string

// Define Redis pub/sub events.
const (
	RedisPubSubChannel                      = "iam.cluster.notifications"
	NoticePolicyChanged NotificationCommand = "PolicyChanged"
	NoticeSecretChanged NotificationCommand = "SecretChanged"
)

// Notification is a type that encodes a message published to a pub sub channel (shared between implementations).
type Notification struct {
	Command       NotificationCommand `json:"command"`
	Payload       string              `json:"payload"`
	Signature     string              `json:"signature"`
	SignatureAlgo crypto.Hash         `json:"algorithm"`
}

// Sign Notification with SHA256 algorithm.
func (n *Notification) Sign() {
	n.SignatureAlgo = crypto.SHA256
	hash := sha256.Sum256([]byte(string(n.Command) + n.Payload))
	n.Signature = hex.EncodeToString(hash[:])
}

// 处理 redis 事件
func handleRedisEvent(v interface{}, handled func(NotificationCommand), reloaded func()) {
	message, ok := v.(*redis.Message)
	if !ok {
		return
	}

	notif := Notification{}
	if err := json.Unmarshal([]byte(message.Payload), &notif); err != nil {
		log.Errorf("Unmarshalling message body failed, malformed: ", err)

		return
	}

	log.Infow("receive redis message", "command", notif.Command, "payload", message.Payload)

	switch notif.Command {
	case NoticePolicyChanged, NoticeSecretChanged:
		log.Infof("Reloading secrets and polices")
		reloadQueue <- reloaded // 收到通知写入 reloaded 事件
	default:
		log.Warnf("Unknown notification command: %q", notif.Command)
		return
	}

	if handled != nil { // 回调
		// went through. all others shoul have returned early.
		handled(notif.Command)
	}
}

// RedisNotifier will use redis pub/sub channels to send notifications.
type RedisNotifier struct {
	store   *storage.RedisCluster
	channel string
}

// Notify will send a notification to a channel.
func (r *RedisNotifier) Notify(notify interface{}) bool {
	if n, ok := notify.(Notification); ok {
		n.Sign() // 签名消息
		notify = n
	}

	toSend, err := json.Marshal(notify)
	if err != nil {
		log.Errorf("Problem marshaling notification: %s", err.Error())

		return false
	}

	log.Debugf("Sending notification: %v", notify)

	// 发送消息到指定的 redis 通道中
	if err := r.store.Publish(r.channel, string(toSend)); err != nil {
		if !errors.Is(err, storage.ErrRedisIsDown) {
			log.Errorf("Could not send notification: %s", err.Error())
		}
		return false
	}

	return true
}
