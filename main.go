package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"time"
)

var (
	timeout      int64
	size         int
	count        int
	typ          uint8 = 8
	code         uint8 = 0
	sendCount    int
	successCount int
	failCount    int
	minTs        int64 = math.MaxInt32
	maxTs        int64
	totalTs      int64
)

// Type8 code0是ping请求固定的字段，其他字段初始化为0

// 定义icmp报文头部,必须按照如下顺序
type ICMP struct {
	Type        uint8
	Code        uint8
	CheckSum    uint16
	ID          uint16
	SequenceNum uint16
}

func main() {
	getCommandArgs()
	desIp := os.Args[len(os.Args)-1] // 最后一个参数是需要ping的ip
	// 使用icmp协议建立连接
	conn, err := net.DialTimeout("ip:icmp", desIp, time.Duration(timeout)*time.Millisecond)
	// 若连接建立失败直接返回
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	fmt.Printf("正在ping %s [%s] 具有 %d 字节的数据\n", desIp, conn.RemoteAddr(), size)

	for i := 0; i < count; i++ {
		sendCount += 1
		t1 := time.Now()
		icmp := &ICMP{
			Type:        typ,
			Code:        code,
			CheckSum:    0,
			ID:          1,
			SequenceNum: 1,
		}

		data := make([]byte, size)
		var buffer bytes.Buffer
		binary.Write(&buffer, binary.BigEndian, icmp) // binary.BigEndian就是大端写入（左边高位右边低位）
		buffer.Write(data)
		data = buffer.Bytes()
		checksum := checkSum(data)
		data[2] = byte(checksum >> 8)
		data[3] = byte(uint8(checksum))

		conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))
		n, err := conn.Write(data) // 将数据写入连接
		if err != nil {
			failCount += 1
			log.Println(err)
			continue
		}
		// 从连接里读数据
		buf := make([]byte, 65535) // 大小为2^16
		n, err = conn.Read(buf)
		if err != nil {
			failCount += 1
			log.Println(err)
			continue
		}
		successCount += 1
		ts := time.Since(t1).Milliseconds()
		if ts < minTs {
			minTs = ts
		}
		if ts > maxTs {
			maxTs = ts
		}
		totalTs = totalTs + ts
		fmt.Printf("来自 %d.%d.%d.%d 的回复：字节=%d 时间=%d ms TTL=%d\n", buf[12], buf[13], buf[14], buf[15], n-28, ts, buf[8])
		time.Sleep(time.Second)
	}
	fmt.Printf("%s 的ping统计信息:\n 数据包：已发送 = %d , 已接受= %d , 丢失 = %d (%.2f%%丢失率) , \n 往返行程的估计时间为（以毫秒为单位）：\n  最短 = %d ms  最长 = %d ms  平均= %d ms \n", conn.RemoteAddr(), sendCount, successCount, failCount, float64(failCount)/float64(sendCount), minTs, maxTs, totalTs/int64(sendCount))

	// fmt.Println(data)
	// 	(base) wangzhujia@wangzhujiadeMacBook-Pro ping % sudo go run main.go -w 150 -l 32 -n 8 www.baidu.com
	// [8 0 0 0 0 1 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]  首部8个加正文32个字节，整整40个

}

// 获取命令行参数
func getCommandArgs() {
	flag.Int64Var(&timeout, "w", 1000, "请求超时时长，单位毫秒")
	flag.IntVar(&size, "l", 32, "请求发送缓冲区大小，单位字节")
	flag.IntVar(&count, "n", 4, "发送请求数")
	flag.Parse() // 定义完参数之后需要parse
}

// 校验和计算
func checkSum(data []byte) uint16 {
	length := len(data)
	index := 0
	var sum uint32 = 0
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		length -= 2
		index += 2
	}
	if length != 0 {
		sum += uint32(data[index])
	}
	hi16 := sum >> 16
	lo16 := uint32(uint16(sum))
	for hi16 != 0 {
		sum = hi16 + lo16
		hi16 = sum >> 16
	}
	return uint16(^sum)
}
