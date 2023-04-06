package LRUCache

import "testing"

func TestLRU1(t *testing.T) {
	bank := NewLRUCache(3)
	bank.Add(1, "1")
	bank.Add(2, "2")
	bank.Add(3, "3")
	bank.Add(4, "4")

	val, _ := bank.Get(1)
	if val != "" {
		t.Fatalf("key 1 should be removed from cache")
	}
}

func TestLRU2(t *testing.T) {
	bank := NewLRUCache(3)
	bank.Add(1, "1")
	bank.Add(2, "2")
	bank.Add(3, "3")
	bank.Add(4, "4")

	_, ok := bank.Get(1)
	if ok {
		t.Fatal("Key 1 should not exist")
	}
}

func TestLRU3(t *testing.T) {
	bank := NewLRUCache(3)
	bank.Add(1, "1")
	bank.Add(2, "2")
	bank.Add(3, "3")
	bank.Add(4, "4")

	val, _ := bank.Get(2)
	if val != "2" {
		t.Fatalf("2 != 2")
	}
}

func TestLRU4(t *testing.T) {
	bank := NewLRUCache(3)
	bank.Add(1, "1")
	bank.Add(2, "2")
	bank.Add(3, "3")
	bank.Add(4, "4")

	val, _ := bank.Get(2)
	if val != "2" {
		t.Fatalf("2 != 2")
	}

	bank.Remove(2)
	val, _ = bank.Get(2)
	if val != "" {
		t.Fatalf("key 2 should be removed from cache")
	}
	val, _ = bank.Get(3)
	if val != "3" {
		t.Fatalf("3!= 3")
	}
	bank.Remove(1)
	val, _ = bank.Get(1)
	if val != "" {
		t.Fatalf("key 1 should be removed from cache")
	}
}

func TestLRU5(t *testing.T) {
	bank := NewLRUCache(3)
	bank.Add(1, "1")
	bank.Add(2, "2")
	bank.Add(3, "3")
	bank.Add(4, "4")

	val, _ := bank.Get(1)
	if val != "" {
		t.Fatalf("key 1 should be removed from cache")
	}

	bank.Remove(4)
	val, _ = bank.Get(4)
	if val != "" {
		t.Fatalf("key 4 should be removed from cache")
	}

	val, _ = bank.Get(2)
	if val != "2" {
		t.Fatalf("2 != 2")
	}

}
