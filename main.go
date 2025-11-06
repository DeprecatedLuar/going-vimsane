package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Message struct {
	LayerChange struct {
		New string `json:"new"`
	} `json:"LayerChange"`
}

func getColorForLayer(layer string) Color {
	switch layer {
	case "default":
		return Blue
	case "vim-normal":
		return Cyan
	case "visual-mode":
		return Magenta
	case "vim-shifted":
		return Orange
	case "visual-shifted":
		return Orange
	case "delete-ops":
		return Red
	case "yank-ops":
		return Yellow
	case "g-ops":
		return Green
	case "meta-layer":
		return Purple
	case "escape":
		return Color{R: 0.5, G: 0.5, B: 0.5}
	default:
		return Cyan
	}
}

func main() {
	// Initialize overlay window
	if err := InitOverlay(); err != nil {
		fmt.Printf("Failed to create overlay: %v\n", err)
		fmt.Println("Make sure you're running X11 (not Wayland)")
		os.Exit(1)
	}
	defer Cleanup()

	// Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		Cleanup()
		os.Exit(0)
	}()

	for {
		// Connect to kanata
		conn, err := net.Dial("tcp", "127.0.0.1:5829")
		if err != nil {
			fmt.Printf("Can't connect to kanata: %v\n", err)
			fmt.Println("Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Println("Connected to kanata tcp")

		// Monitor layer changes
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			var msg Message
			if err := json.Unmarshal(scanner.Bytes(), &msg); err == nil {
				if msg.LayerChange.New != "" {
					layer := msg.LayerChange.New

					if layer == "default" || layer == "meta-layer" {
						fmt.Printf("Layer: %-15s → Hidden\n", layer)
						HideWindow()
					} else {
						color := getColorForLayer(layer)
						fmt.Printf("Layer: %-15s → Color: RGB(%.1f, %.1f, %.1f)\n",
							layer, color.R, color.G, color.B)
						ShowWindow()
						DrawBorder(color, BorderWidth)
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading from kanata: %v\n", err)
		}

		conn.Close()
		fmt.Println("Connection lost. Retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
	}
}
