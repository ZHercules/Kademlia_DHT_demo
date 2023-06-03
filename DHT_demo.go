package DHT_demo

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"time"
)

const (
	K          = 20
	bucketSize = 3
	numPeers   = 5
)

type Node struct {
	id  string
	dht *DHT
}

type DHT struct {
	store map[string][]byte // Store key-value pairs
	nodes []*Node           // Nodes stored in the DHT
}

type Bucket struct {
	nodes []*Node
}

type KBucket struct {
	buckets []*Bucket
}

type Peer struct {
	kbucket *KBucket
	dht     *DHT
}

func NewNode(id string) *Node {
	return &Node{id: id, dht: &DHT{store: make(map[string][]byte), nodes: make([]*Node, 0, K)}}
}

func NewBucket() *Bucket {
	return &Bucket{nodes: make([]*Node, 0, bucketSize)}
}

func NewKBucket() *KBucket {
	return &KBucket{buckets: make([]*Bucket, K)}
}

func NewPeer(id string) *Peer {
	return &Peer{kbucket: NewKBucket(), dht: &DHT{store: make(map[string][]byte), nodes: make([]*Node, 0, K)}}
}

// SetValue stores a key-value pair in the DHT.
func (n *Node) SetValue(key, value []byte) bool {
	hash := sha256.Sum256(value)
	if hex.EncodeToString(hash[:]) != string(key) {
		return false
	}

	// Check if the key already exists in the DHT
	if _, ok := n.dht.store[string(key)]; ok {
		return true
	}

	// Store the key-value pair in the DHT
	n.dht.store[string(key)] = value

	// Find the closest nodes to the key and call SetValue on them
	closestNodes := n.dht.selectClosestNodes(n.id, 2)
	for _, node := range closestNodes {
		node.SetValue(key, value)
	}

	return true
}

// GetValue retrieves a value for a given key from the DHT.
func (p *Peer) GetValue(key []byte) []byte {
	// Check if the key exists in the DHT
	if value, ok := p.dht.store[string(key)]; ok {
		return value
	}

	// Find the closest nodes to the key and call GetValue on them
	closestNodes := p.kbucket.selectClosestNodes(p.dht.nodes, p.dht.store, string(key), 2)
	for _, node := range closestNodes {
		if value := node.GetValue(key); value != nil {
			hash := sha256.Sum256(value)
			// Check if the value matches the key
			if hex.EncodeToString(hash[:]) == string(key) {
				return value
			}
		}
	}

	return nil
}

// You'll need to implement selectClosestNodes, which selects the n closest nodes to a given node ID.
// This can be done by modifying the existing selectRandomNodes function to sort the nodes by distance
// before returning them.
func (kb *KBucket) selectClosestNodes(nodes []*Node, store map[string][]byte, nodeId string, n int) []*Node {
	// Select n closest nodes from all buckets
	var closestNodes []*Node
	for _, bucket := range kb.buckets {
		closestNodes = append(closestNodes, bucket.nodes...)
	}

	// Sort the nodes by their distance to the target node ID
	sort.Slice(closestNodes, func(i, j int) bool {
		distanceI := calculateDistance(closestNodes[i].id, nodeId)
		distanceJ := calculateDistance(closestNodes[j].id, nodeId)
		return distanceI.Cmp(distanceJ) < 0
	})

	return closestNodes[:n]
}

// You'll also need to implement calculateDistance, which calculates the XOR distance between two node IDs.
func calculateDistance(nodeId1, nodeId2 string) *big.Int {
	// Convert the node IDs from hexadecimal strings to big.Int
	nodeId1Int := new(big.Int)
	nodeId1Int.SetString(nodeId1, 16)

	nodeId2Int := new(big.Int)
	nodeId2Int.SetString(nodeId2, 16)

	// Calculate the XOR distance
	distance := new(big.Int)
	distance.Xor(nodeId1Int, nodeId2Int)

	return distance
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Initialize 100 nodes
	var nodes []*Node
	for i := 0; i < 100; i++ {
		nodes = append(nodes, NewNode(fmt.Sprintf("node%d", i)))
	}

	// Generate 200 random strings and their hashes, and randomly select a node to call SetValue
	keys := make([][]byte, 200)
	values := make(map[string]string, 200)
	for i := 0; i < 200; i++ {
		value := randomString(rand.Intn(10) + 1) // Random length between 1 and 10
		hash := sha256.Sum256([]byte(value))
		key := hash[:]
		keys[i] = key
		values[hex.EncodeToString(key)] = value

		node := nodes[rand.Intn(len(nodes))]
		node.SetValue(key, []byte(value))
	}

	// Randomly select 100 keys and a node for each key to call GetValue
	for i := 0; i < 100; i++ {
		key := keys[rand.Intn(len(keys))]
		node := nodes[rand.Intn(len(nodes))]
		value := node.GetValue(key)

		// Check if the value matches the original value
		if string(value) != values[hex.EncodeToString(key)] {
			fmt.Printf("GetValue returned incorrect value for key %x\n", key)
		}
	}
}
