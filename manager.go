package main

import (
    "os"
    "log"
    "time"
//    "syscall"
)

type ManagerMsg int
const (
    /* FROM manager */
    Done      ManagerMsg = iota

    /* TO manager */
    RemoveOld ManagerMsg = iota

    /* Internal */
    FinishUp  ManagerMsg = iota
)

type ClientManager struct {
    Clients    map[*Client]bool
    Signals    chan os.Signal   // OS sends, manager receives
    Register   chan *Client     // User sends, manager receives
    Message    chan ManagerMsg  // User and manager, send / receive
    SleepTime  time.Duration
}

func (manager *ClientManager) Init() {
    manager.Clients    = make(map[*Client]bool)
    manager.Signals    = make(chan os.Signal)
    manager.Register   = make(chan *Client)
    manager.Message    = make(chan ManagerMsg)
    manager.SleepTime  = time.Second
}

func (manager *ClientManager) Start() {
    /* Internal channel to stop cleaner goroutine */
    cleanup := make(chan bool)

    /* Main manager goroutine */
    go func() {
        defer func() {
            /* We should have exited before this, but :shrug: */
            close(manager.Register)
            close(manager.Signals)
            close(manager.Message)
        }()

        for {
            select {
                /* New client received */
                case client := <-manager.Register:
                    /* TODO: decrease SleepTime for when more connections added */
                    manager.Clients[client] = true
                    client.Start()

                /* Message receieved */
                case msg := <-manager.Message:
                    switch msg {
                        default:
                            /* do nothing */
                    }

                /* OS signal received */
                case sig := <-manager.Signals:
                    log.Printf("SIGNAL RECEIVED: %v\n", sig)
                        log.Printf("Received %v, waiting on cleanup before exit... (ctrl-c again to stop NOW)\n", sig)
                        cleanup<-true
                        select {
                            case sig = <-manager.Signals:
                                log.Printf("Signal received again, exiting now.\n")
                                return
                            case <-cleanup:
                                return
                        }
            }
        }
    }()

     /* Start cleaner goroutine in background */
    go manager.Cleaner(cleanup)
}

func (manager *ClientManager) Cleaner(cleanup chan bool) {
    finishUp := false
    for {
        /* Check for cleanup signal */
        select {
            case _ = <-cleanup:
                finishUp = true
            default:
                /* do nothing */
        }

        for client := range manager.Clients {
            select {
                case <-client.Message:
                    delete(manager.Clients, client)
                default:
                    /* do nothing */
            }
        }

        if finishUp {
            break
        } else {
            time.Sleep(manager.SleepTime)
        }
    }

    /* Final cleanup */
    for client := range manager.Clients {
        <-client.Message
        delete(manager.Clients, client)
    }

    /* Close the cleanup channel -- indicates we're done! */
    close(cleanup)
}
