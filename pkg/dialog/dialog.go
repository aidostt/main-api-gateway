package dialog

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"reservista.kz/pkg/logger"
)

type DialogService interface {
	NewConnection(string) (*grpc.ClientConn, error)
}

type Dialog struct {
	Addresses Addresses
	authority string
}
type Addresses struct {
	Users         string
	Reservations  string
	QRs           string
	Notifications string
}

func NewDialog(authority, users, reservations, qrs, notifications string) *Dialog {
	return &Dialog{authority: authority, Addresses: Addresses{Users: users, Reservations: reservations, QRs: qrs, Notifications: notifications}}
}

func (d *Dialog) NewConnection(address string) (*grpc.ClientConn, error) {
	//cert, err := tls.LoadX509KeyPair("path/to/server.crt", "path/to/server.key")
	//if err != nil {
	//	return nil, err
	//}
	//tlsConfig := &tls.Config{
	//	Certificates: []tls.Certificate{cert},
	//	ServerName:   d.authority,
	//}
	//creds := credentials.NewTLS(tlsConfig)
	//conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds))
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Errorf("Failed to connect: %v", err)
		conn.Close()
		return nil, err
	}
	return conn, nil
}
