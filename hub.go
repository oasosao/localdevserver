package main

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Comm struct {
	clients         map[*Client]bool
	sendMsgToClient chan []byte
	clientReg       chan *Client
	clientUnReg     chan *Client
}

func newComm() *Comm {
	return &Comm{
		sendMsgToClient: make(chan []byte),
		clientReg:       make(chan *Client),
		clientUnReg:     make(chan *Client),
		clients:         make(map[*Client]bool),
	}
}

func (h *Comm) run() {

	for {
		select {

		case client := <-h.clientReg:
			h.clients[client] = true
		case client := <-h.clientUnReg:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.sendMsgToClient:

			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// 监听文件操作
func (h *Comm) watcherFile() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("监听目录发生错误", "error", err)
		return
	}
	defer watcher.Close()

	filepath.Walk(RootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			slog.Error("监听目录发生错误", "error", err)
			return err
		}

		if info.IsDir() {
			watcher.Add(path)
			fmt.Println("监听目录: ", path)
		}
		return nil
	})

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// 如果在侦听器目录中创建新目录，请将其添加到侦听器中
			if event.Op.String() == "CREATE" {
				fs, err := os.Stat(event.Name)
				if err == nil {
					if fs.IsDir() {
						watcher.Add(event.Name)
					}
				}
			}

			// 侦听文件操作
			if event.Op.String() == "WRITE" {
				if filepath.Ext(event.Name) == ".html" || filepath.Ext(event.Name) == ".htm" {
					h.sendMsgToClient <- []byte("reload")
				} else {
					h.sendMsgToClient <- []byte("update")
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				slog.Error("监听目录发生错误", "error", err)
				return
			}

		}
	}

}
