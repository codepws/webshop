package steup

import (
	"common/settings"
	"log"
	"webshop-service/inventory-srv/servicer"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

func SteupRocketMQReg(RocketMQ *settings.RocketMQ) error {

	log.Println("MQ订单库存归还消息者：", RocketMQ.OrderInvReback.CGroupName, RocketMQ.OrderInvReback.Topic, RocketMQ.OrderInvReback.Tag)

	//订单支付超时 Consumer
	cer, err := rocketmq.NewPushConsumer(
		consumer.WithGroupName(RocketMQ.OrderInvReback.CGroupName),
		consumer.WithNsResolver(primitive.NewPassthroughResolver(RocketMQ.NameServers)),
		//consumer.WithConsumerOrder(true), //顺序消费
		consumer.WithAutoCommit(false),                  //手动提交
		consumer.WithConsumerModel(consumer.Clustering), //集群模式：同一个消费组的同一个Topic的多个消费，只能被多个消费者分摊消费

		//consumer.WithConsumeFromWhere(consumer.ConsumeFromFirstOffset),
		//consumer.WithConsumerModel(consumer.BroadCasting), 	//广播模式

		//consumer.WithMaxReconsumeTimes(3),	//consumer 客户端参数， 最大重试次数，超过最大重试次数，消息将被转 移到私信队列。 作用域：无序和顺序的集群消费起作用。 设置方式：默认值 无序消息 16 次，顺序消息 -1 表示无限次本地重试。
		consumer.WithRetry(3),
	)
	if err != nil {
		log.Println("rocketmq.NewPushConsumer失败：", err.Error())
		return err
	}
	defer cer.Shutdown()

	topic := RocketMQ.OrderInvReback.Topic
	tag := RocketMQ.OrderInvReback.Tag
	selector := consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: tag,
	}
	err = cer.Subscribe(topic, selector, servicer.MQOrderInvRebackCallback)
	if err != nil {
		log.Println("PushConsumer.Subscribe失败：", err.Error())
		return err
	}
	// Note: start after subscribe
	err = cer.Start()
	if err != nil {
		log.Println("start after subscribe error:", err.Error())
		return err
	}
	return nil
}
