package cache

import (
	"sync"
)

type Lru struct {
	capacity int
	cache    map[string]*listNode
	head     *listNode
	tail     *listNode
	mu       sync.RWMutex
}

type listNode struct {
	key  string
	prev *listNode
	next *listNode
}

func NewCache(capacity int) *Lru {
	if capacity <= 0 {
		capacity = 100
	}
	return &Lru{
		capacity: capacity,
		cache:    make(map[string]*listNode),
	}
}

func (c *Lru) Add(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if node, exists := c.cache[key]; exists {
		c.moveToFront(node)
		return
	}

	node := &listNode{key: key}
	c.cache[key] = node
	c.moveToFront(node)

	if len(c.cache) > c.capacity {
		c.removeOldest()
	}
}

func (c *Lru) Update(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if node, exists := c.cache[key]; exists {
		c.moveToFront(node)
	}
}

func (c *Lru) RemoveOldest() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.removeOldest()
}

func (c *Lru) IsFull() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache) >= c.capacity
}

func (c *Lru) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}

func (c *Lru) Contains(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.cache[key]
	return exists
}

func (c *Lru) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*listNode)
	c.head = nil
	c.tail = nil
}

func (c *Lru) moveToFront(node *listNode) {
	if c.head == node {
		return
	}

	if node.prev != nil {
		node.prev.next = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	}
	if c.tail == node {
		c.tail = node.prev
	}

	node.next = c.head
	node.prev = nil
	if c.head != nil {
		c.head.prev = node
	}
	c.head = node
	if c.tail == nil {
		c.tail = node
	}
}

func (c *Lru) removeOldest() string {
	if c.tail == nil {
		return ""
	}

	key := c.tail.key
	delete(c.cache, key)

	if c.tail.prev != nil {
		c.tail.prev.next = nil
	}
	c.tail = c.tail.prev
	if c.tail == nil {
		c.head = nil
	}

	return key
}
