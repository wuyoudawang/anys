package services

// import (
// 	"net"
// 	"net/http"
// 	"time"

// 	"anys/jobs"
// )

// type HttpServer struct {
// 	http.Server

// 	l   net.Listener
// 	eng *jobs.Engine
// }

// func (h *HttpServer)

// func (h *HttpServer) Serve() {
// 	defer h.l.Close()
// 	var tempDelay time.Duration // how long to sleep on accept failure
// 	for {
// 		rw, e := h.l.Accept()
// 		if e != nil {
// 			if ne, ok := e.(net.Error); ok && ne.Temporary() {
// 				if tempDelay == 0 {
// 					tempDelay = 5 * time.Millisecond
// 				} else {
// 					tempDelay *= 2
// 				}
// 				if max := 1 * time.Second; tempDelay > max {
// 					tempDelay = max
// 				}
// 				srv.logf("http: Accept error: %v; retrying in %v", e, tempDelay)
// 				time.Sleep(tempDelay)
// 				continue
// 			}
// 			return e
// 		}
// 		tempDelay = 0
// 		c, err := srv.newConn(rw)
// 		if err != nil {
// 			continue
// 		}
// 		c.setState(c.rwc, StateNew) // before Serve can return
// 		go c.serve()
// 	}
// }
