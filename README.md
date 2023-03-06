# golang_ping

# 100行代码实现ping操作

## ping操作原理与ICMP协议

ICMP是Internet控制报文协议

### ICMP报文格式

<img src="/Users/wangzhujia/Library/Application Support/typora-user-images/image-20230306151239771.png" alt="image-20230306151239771" style="zoom:67%;" />

## 实现ping操作的两个关键点

1. 定义ICMP报文头部结构体（按照顺序定义）

   ![image-20230306151649129](/Users/wangzhujia/Library/Application Support/typora-user-images/image-20230306151649129.png)

2. ICMP校验和算法

   * 报文内容、相邻两个字节拼接到一起组成一个16bit的数，将这些数累加求和
   * 若长度为奇数，则将剩余的1个字节也累加到求和
   * 得出总和后，将和值的高16位与低16位不断求和，直到高16位为0
   * 以上结果得出后取反，低16位即为校验和

## 实现ping操作

### 定义常量

```go
var (
	timeout      int64
	size         int
	count        int
	typ          uint8 = 8   // ping请求固定的type
	code         uint8 = 0   // ping请求固定的code
	sendCount    int
	successCount int
	failCount    int
	minTs        int64 = math.MaxInt32
	maxTs        int64
	totalTs      int64
)
```

### 定义icmp报文头部

```go
// 必须按照如下顺序, 一共八个字节
type ICMP struct {
	Type        uint8
	Code        uint8
	CheckSum    uint16
	ID          uint16
	SequenceNum uint16
}
```

### 获取命令行参数

```go
func getCommandArgs() {
	flag.Int64Var(&timeout, "w", 1000, "请求超时时长，单位毫秒")
	flag.IntVar(&size, "l", 32, "请求发送缓冲区大小，单位字节")
	flag.IntVar(&count, "n", 4, "发送请求数")
	flag.Parse() // 定义完参数之后需要parse
}
```

### 校验和计算

```go
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
```

***==详细main方法请看代码main.go==***

