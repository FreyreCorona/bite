package main

var Scr [64][32]uint8

func clearScr() {
	for i := range 32 {
		for j := range 64 {
			Scr[i][j] = 0
		}
	}
}
