package server

import (
	"bufio"
	"github.com/Sirupsen/logrus"
	"github.com/oleiade/lane"
	"net"
	"strconv"
	"sync"
)

type SplitterMap func([]byte) []byte

type ConnectionList struct {
	sync.RWMutex
	connections []net.Conn
}

func (cl *ConnectionList) AddConnection(conn net.Conn) {
	cl.Lock()
	defer cl.Unlock()
	cl.connections = append(cl.connections, conn)
}

func (cl *ConnectionList) removeConnectionAt(i int) {
	// Can only be entered while lock is acquired by Broadcast
	// Close it first
	logrus.Infof("REMOVING %d", i)
	cl.connections[i].Close()
	// Remove it
	cl.connections = append(cl.connections[:i], cl.connections[i+1:]...)
}

func NewConnectionList() *ConnectionList {
	return &ConnectionList{
		connections: make([]net.Conn, 0),
	}
}

func (cl *ConnectionList) Broadcast(bytes []byte) {
	cl.Lock()
	defer cl.Unlock()
	for i, conn := range cl.connections {
		num, err := conn.Write(bytes)
		if num == 0 {
			if err != nil {
				// connection closed, bail?
				defer cl.removeConnectionAt(i)
			}
		}
	}
}

func StartReadListening(readPort int, writePort int, defaultWriters []string, mapOperation SplitterMap) {
	// Create buffer to hold data
	queue := lane.NewQueue()
	// Start listening for writer destinations
	go StartWriteListening(queue, defaultWriters, writePort)

	socket, err := net.Listen("tcp", ":"+strconv.Itoa(readPort))
	if err != nil {
		logrus.Error(err)
	}
	// This will block the main thread
	for {
		// Begin trying to accept connections
		logrus.Debug("Awaiting Connection...")
		//Block and wait for listeners
		conn, err := socket.Accept()
		if err != nil {
			logrus.Error(err)
		} else {
			logrus.Debug("Accepted Connection...")
			go HandleReadConnection(conn, queue, writePort, mapOperation)
		}
	}
}

func StartWriteListening(readQueue *lane.Queue, defaultWriters []string, writePort int) {
	cList := NewConnectionList()
	socket, err := net.Listen("tcp", ":"+strconv.Itoa(writePort))
	if err != nil {
		logrus.Error(err)
	}
	// Begin trying to connect to default endpoints
	for _, writer := range defaultWriters {
		logrus.Debug("Opening connections to endpoints...")
		conn, err := net.Dial("tcp", writer)
		if err != nil {
			logrus.Error(err)
		} else {
			logrus.Debug("Accepted Connection...")
			cList.AddConnection(conn)
			go HandleWriteConnections(cList, readQueue)
		}
	}

	for {
		// Begin trying to accept connections
		logrus.Debug("Awaiting Connection...")
		//Block and wait for listeners
		conn, err := socket.Accept()
		if err != nil {
			logrus.Error(err)
		} else {
			logrus.Debug("Accepted Connection...")
			cList.AddConnection(conn)
			go HandleWriteConnections(cList, readQueue)
		}
	}
}

func HandleReadConnection(conn net.Conn, readQueue *lane.Queue, writePort int, mapOperation SplitterMap) {

	buffConn := bufio.NewReaderSize(conn, 1024)
	buffer := make([]byte, 1024)
	for {
		logrus.Debug("Begining Read")
		bytes, err := buffConn.Read(buffer)

		// Output the content of the bytes to the queue
		if bytes > 0 {
			readQueue.Enqueue(buffer[:bytes])
			logrus.Info(string(mapOperation(buffer[:bytes])[:bytes]))
		}

		if bytes == 0 {
			if err.Error() == "EOF" {
				logrus.Error("End of individual transmission")
				buffConn = nil
				conn.Close()
				return
			}
		}

		if err != nil {
			logrus.Error("Underlying network failure?")
			logrus.Error(err)
			break
		}
	}
}

func HandleWriteConnections(cList *ConnectionList, readQueue *lane.Queue) {
	for {
		data := readQueue.Dequeue()
		if data != nil {
			bytes := data.([]byte)
			cList.Broadcast(bytes)
		}
	}
}
