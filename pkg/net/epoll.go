package net

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

package main

import (
"fmt"
"net"
"os"
"syscall"

"golang.org/x/sys/unix"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}

	err = syscall.SetNonblock(fd, true)
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}

	err = syscall.Bind(fd, &syscall.SockaddrInet4{Port: 8080})
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}

	err = syscall.Listen(fd, 128)
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}

	epfd, err := unix.EpollCreate1(0)
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}

	var event unix.EpollEvent
	event.Events = unix.EPOLLIN | unix.EPOLLET
	event.Fd = int32(fd)

	err = unix.EpollCtl(epfd, unix.EPOLL_CTL_ADD, fd, &event)
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}

	events := make([]unix.EpollEvent, 1024)

	for {
		n, err := unix.EpollWait(epfd, events, -1)
		if err != nil {
			fmt.Printf("Error: %s", err)
			os.Exit(1)
		}

		for i := 0; i < n; i++ {
			if int(events[i].Fd) == fd {
				connFd, _, err := syscall.Accept(fd)
				if err != nil {
					fmt.Printf("Error: %s", err)
					continue
				}

				fmt.Printf("New connection: %d\n", connFd)

				var connEvent unix.EpollEvent
				connEvent.Events = unix.EPOLLIN | unix.EPOLLET
				connEvent.Fd = int32(connFd)

				err = unix.EpollCtl(epfd, unix.EPOLL_CTL_ADD, connFd, &connEvent)
				if err != nil {
					fmt.Printf("Error: %s", err)
					continue
				}
			} else {
				buf := make([]byte, 1024)

				n, err := syscall.Read(int(events[i].Fd), buf)
				if err != nil {
					fmt.Printf("Error: %s", err)
					syscall.Close(int(events[i].Fd))
					continue
				}

				fmt.Printf("Received data from %d: %s\n", events[i].Fd, string(buf[:n]))
			}
		}
	}
}
