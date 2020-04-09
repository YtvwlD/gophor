package main

type ClientManager struct {
    Cmd        chan Command
    Clients    map[*Client]bool
    Register   chan *Client
    Unregister chan *Client
}

func (manager *ClientManager) Init() {
    manager.Cmd        = make(chan Command)
    manager.Clients    = make(map[*Client]bool)
    manager.Register   = make(chan *Client)
    manager.Unregister = make(chan *Client)
}

func (manager *ClientManager) Start() {
    go func() {
        for {
            select {
                case cmd := <-manager.Cmd:
                    /* Received manager command, handle! */
                    switch cmd {
                        case Stop:
                            /* Stop all clients then delete, break out of run loop */
                            for client := range manager.Clients {
                                client.Cmd<-Stop
                                delete(manager.Clients, client)
                            }
                            break

                        case Clean:
                            /* Delete all 'done' clients */
                            for client := range manager.Clients {
                                select {
                                    case <-client.Cmd:
                                        /* Channel closed, client done, delete! */
                                        delete(manager.Clients, client)
                                    default:
                                        /* Don't lock :) */
                                }
                            }
                    }

                case client := <-manager.Register:
                    /* Received new client to register */
                    manager.Clients[client] = true
                    client.Start()

                case client := <-manager.Unregister:
                    /* Received client id to unregister */
                    if _, ok := manager.Clients[client]; ok {
                        client.Cmd<-Stop
                        delete(manager.Clients, client)
                    }
            }
        }
    }()
}
