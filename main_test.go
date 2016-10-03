package main

import "testing"

func Test_compactIfName(t *testing.T) {
	t.Run("Ifname", func(t *testing.T) {
		t.Run("FastEthernet", func(t *testing.T) {
			data := compactIfName("FastEthernet0/1")
			if data != "Fa 0/1" {
				t.Error("Error compaction ifname")
			}
		})
		t.Run("GigabitEthernet", func(t *testing.T) {
			data := compactIfName("GigabitEthernet0/1")
			if data != "Gi 0/1" {
				t.Error("Error compaction ifname")
			}
		})
		t.Run("TenGigabitEthernet", func(t *testing.T) {
			data := compactIfName("TenGigabitEthernet0/1")
			if data != "Te 0/1" {
				t.Error("Error compaction ifname")
			}
		})
	})
}
