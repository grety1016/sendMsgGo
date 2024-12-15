package eventful

import (
	"fmt"
	"sync"
	"sync/atomic"
)

//#region 事件驱动包

// 事件分发器
type Eventful struct {
	nextId      uint64
	subscribers sync.Map
}

type Topic string

// 构造函数
func NewTopic(value string) Topic { return Topic(value) }

// 获取 Topic 的值
func (t Topic) Value() string { return string(t) }

// 初始化事件分发器
func NewEventful() *Eventful {
	return &Eventful{}
}

// 订阅者结构
type Subscriber struct {
	id      uint64
	topic   Topic
	handler func(interface{})
}

// 订阅事件
func (e *Eventful) Subscribe(topic Topic, handler func(interface{})) uint64 {
	subId := atomic.AddUint64(&e.nextId, 1) - 1
	sub := &Subscriber{id: subId, topic: topic, handler: handler}
	e.subscribers.Store(subId, sub)
	return subId
}

// 取消订阅事件
func (e *Eventful) Unsubscribe(subId uint64) {
	e.subscribers.Delete(subId)
}

// 发布主题事件
func (e *Eventful) Publish(topic Topic, msg interface{}) {
	e.subscribers.Range(func(key, value interface{}) bool {
		sub := value.(*Subscriber)
		if sub.topic == topic {
			sub.handler(msg)
		}
		return true
	})
}

func EventDemo() {
	// 自定义消息类型
	type CustomMessage struct {
		Content string
		Code    int
	}
	eventful := NewEventful()

	// 创建自定义主题
	customTopic := NewTopic("CustomMessage")
	fmt.Println("Custom topic:", customTopic.Value())

	// 定义事件处理函数
	handler := func(msg interface{}) {
		if cm, ok := msg.(CustomMessage); ok {
			fmt.Printf("Received custom message: %s with code %d\n", cm.Content, cm.Code)
		}
	}

	// 订阅自定义主题
	subId := eventful.Subscribe(customTopic, handler)

	// 发布自定义消息
	eventful.Publish(customTopic, CustomMessage{Content: "Hello, World!", Code: 200})

	// 取消订阅
	eventful.Unsubscribe(subId)

	// 再次发布自定义消息，确保订阅者已取消
	eventful.Publish(customTopic, CustomMessage{Content: "Goodbye, World!", Code: 404})
}

//#endregion 事件驱动包
