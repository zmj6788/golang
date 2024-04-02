# 动手写分布式缓存 - GeeCache第四天 一致性哈希(hash)

- ## 一致性哈希(consistent hashing)的原理以及为什么要使用一致性哈希

- ## 实现一致性哈希代码，添加相应的测试用例

## 一致性哈希原理

一致性哈希算法将 key 映射到 2^32 的空间中，将这个数字首尾相连，形成一个环。

- 计算节点/机器(通常使用节点的名称、编号和 IP 地址)的哈希值，放置在环上。
- 计算 key 的哈希值，放置在环上，顺时针寻找到的第一个节点，就是应选取的节点/机器。

![dd_pee](Z:\浏览器下载\add_peer.jpg)

环上有 peer2，peer4，peer6 三个节点，`key11`，`key2`，`key27` 均映射到 peer2，`key23` 映射到 peer4。此时，如果新增节点/机器 peer8，假设它新增位置如图所示，那么只有 `key27` 从 peer2 调整到 peer8，其余的映射均没有发生改变。

也就是说，一致性哈希算法，在新增/删除节点时，只需要重新定位该节点附近的一小部分数据，而不需要重新定位所有的节点，这就解决了上述的问题。我该访问谁？和节点数量变化了怎么办？

## 数据倾斜问题

如果服务器的节点过少，容易引起 key 的倾斜。例如上面例子中的 peer2，peer4，peer6 分布在环的上半部分，下半部分是空的。那么映射到环下半部分的 key 都会被分配给 peer2，key 过度向 peer2 倾斜，缓存节点间负载不均。

为了解决这个问题，引入了虚拟节点的概念，一个真实节点对应多个虚拟节点。

假设 1 个真实节点对应 3 个虚拟节点，那么 peer1 对应的虚拟节点是 peer1-1、 peer1-2、 peer1-3（通常以添加编号的方式实现），其余节点也以相同的方式操作。

- 第一步，计算虚拟节点的 Hash 值，放置在环上。
- 第二步，计算 key 的 Hash 值，在环上顺时针寻找到应选取的虚拟节点，例如是 peer2-1，那么就对应真实节点 peer2。

虚拟节点扩充了节点的数量，解决了节点较少的情况下数据容易倾斜的问题。而且代价非常小，只需要增加一个字典(map)维护真实节点与虚拟节点的映射关系即可。

1.consistenthash.go

自定义函数类型 Hash

```
type Hash func(data []byte) uint32
```

Map结构体，实现节点匹配机制，暴漏给用户使用

```
type Map struct {
 hash     Hash
 replicas int
 keys     []int // Sorted
 hashMap  map[int]string
}
```

实例化Map，设置默认Hash算法

```
func New(replicas int, fn Hash) *Map {
 m := &Map{
  replicas: replicas,
  hash:     fn,
  hashMap:  make(map[int]string),
 }
 if m.hash == nil {
  m.hash = crc32.ChecksumIEEE
 }
 return m
}
```

Add方法增加需要匹配的真实节点对应的虚拟节点到匹配环keys上，添加虚拟节点与真实节点映射关系

```
func (m *Map) Add(keys ...string) {
 for _, key := range keys {
  for i := 0; i < m.replicas; i++ {
   hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
   m.keys = append(m.keys, hash)
   m.hashMap[hash] = key
  }
 }
 sort.Ints(m.keys)
}
```

Get根据key在keys环上匹配真实节点

```
func (m *Map) Get(key string) string {
 if len(m.keys) == 0 {
  return ""
 }

 hash := int(m.hash([]byte(key)))
 // Binary search for appropriate replica.
 idx := sort.Search(len(m.keys), func(i int) bool {
  return m.keys[i] >= hash
 })

 return m.hashMap[m.keys[idx%len(m.keys)]]
}
```

至此一致性哈希匹配机制完成

2.consistenthash_test.go

测试一致性哈希匹配机制

测试目标：能否每次匹配同一节点? 新增节点后，能否匹配同一节点呢？

测试步骤：

1.实例化Map对象hash，设置虚拟节点倍数和自定义简单hash算法

2.为实例化对象，添加节点，根据自定义hash算法和添加节点信息，设置已知应匹配节点的测试数据

3.将测试数据循环匹配，比较期待值与实际值，相同则成功验证匹配机制，否则返回错误

4.再次添加节点后，某个测试数据的期待值变化，再次循环验证，相同则成功验证匹配机制，否则返回错误

至此验证结束

```
func TestHashing(t *testing.T) {
 hash := New(3, func(key []byte) uint32 {
  i, _ := strconv.Atoi(string(key))
  return uint32(i)
 })

 // Given the above hash function, this will give replicas with "hashes":
 // 2, 4, 6, 12, 14, 16, 22, 24, 26
 hash.Add("6", "4", "2")

 testCases := map[string]string{
  "2":  "2",
  "11": "2",
  "23": "4",
  "27": "2",
 }

 for k, v := range testCases {
  if hash.Get(k) != v {
   t.Errorf("Asking for %s, should have yielded %s", k, v)
  }
 }

 // Adds 8, 18, 28
 hash.Add("8")

 // 27 should now map to 8.
 testCases["27"] = "8"

 for k, v := range testCases {
  if hash.Get(k) != v {
   t.Errorf("Asking for %s, should have yielded %s", k, v)
  }
 }

}
```
